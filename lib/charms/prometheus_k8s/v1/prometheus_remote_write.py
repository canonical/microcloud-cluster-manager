# Copyright 2021 Canonical Ltd.
# See LICENSE file for licensing details.
"""# Prometheus remote-write library.

This library facilitates the integration of the prometheus_remote_write interface.

Source code can be found on GitHub at:
 https://github.com/canonical/prometheus-k8s-operator/tree/main/lib/charms/prometheus_k8s

Charms that need to push data to a charm exposing the Prometheus remote_write API,
should use the `PrometheusRemoteWriteConsumer`. Charms that operate software that exposes
the Prometheus remote_write API, that is, they can receive metrics data over remote_write,
should use the `PrometheusRemoteWriteProducer`.
"""

import json
import logging
import os
import platform
import re
import socket
import subprocess
import tempfile
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, Tuple, Union

import yaml
from cosl import JujuTopology
from cosl.rules import AlertRules, generic_alert_groups
from ops.charm import (
    CharmBase,
    HookEvent,
    RelationBrokenEvent,
    RelationEvent,
    RelationMeta,
    RelationRole,
)
from ops.framework import BoundEvent, EventBase, EventSource, Object, ObjectEvents
from ops.model import Relation

# The unique Charmhub library identifier, never change it
LIBID = "f783823fa75f4b7880eb70f2077ec259"

# Increment this major API version when introducing breaking changes
LIBAPI = 1

# Increment this PATCH version before using `charmcraft publish-lib` or reset
# to 0 if you are raising the major API version
LIBPATCH = 7

PYDEPS = ["cosl"]


logger = logging.getLogger(__name__)


DEFAULT_RELATION_NAME = "receive-remote-write"
DEFAULT_CONSUMER_NAME = "send-remote-write"
RELATION_INTERFACE_NAME = "prometheus_remote_write"

DEFAULT_ALERT_RULES_RELATIVE_PATH = "./src/prometheus_alert_rules"


class RelationNotFoundError(Exception):
    """Raised if there is no relation with the given name."""

    def __init__(self, relation_name: str):
        self.relation_name = relation_name
        self.message = "No relation named '{}' found".format(relation_name)

        super().__init__(self.message)


class RelationInterfaceMismatchError(Exception):
    """Raised if the relation with the given name has a different interface."""

    def __init__(
        self,
        relation_name: str,
        expected_relation_interface: str,
        actual_relation_interface: str,
    ):
        self.relation_name = relation_name
        self.expected_relation_interface = expected_relation_interface
        self.actual_relation_interface = actual_relation_interface
        self.message = (
            "The '{}' relation has '{}' as its interface rather than the expected '{}'".format(
                relation_name, actual_relation_interface, expected_relation_interface
            )
        )

        super().__init__(self.message)


class RelationRoleMismatchError(Exception):
    """Raised if the relation with the given name has a different direction."""

    def __init__(
        self,
        relation_name: str,
        expected_relation_role: RelationRole,
        actual_relation_role: RelationRole,
    ):
        self.relation_name = relation_name
        self.expected_relation_interface = expected_relation_role
        self.actual_relation_role = actual_relation_role
        self.message = "The '{}' relation has role '{}' rather than the expected '{}'".format(
            relation_name, repr(actual_relation_role), repr(expected_relation_role)
        )

        super().__init__(self.message)


class InvalidAlertRuleEvent(EventBase):
    """Event emitted when alert rule files are not parsable.

    Enables us to set a clear status on the provider.
    """

    def __init__(self, handle, errors: str = "", valid: bool = False):
        super().__init__(handle)
        self.errors = errors
        self.valid = valid

    def snapshot(self) -> Dict:
        """Save alert rule information."""
        return {
            "valid": self.valid,
            "errors": self.errors,
        }

    def restore(self, snapshot):
        """Restore alert rule information."""
        self.valid = snapshot["valid"]
        self.errors = snapshot["errors"]


def _is_official_alert_rule_format(rules_dict: dict) -> bool:
    """Are alert rules in the upstream format as supported by Prometheus.

    Alert rules in dictionary format are in "official" form if they
    contain a "groups" key, since this implies they contain a list of
    alert rule groups.

    Args:
        rules_dict: a set of alert rules in Python dictionary format

    Returns:
        True if alert rules are in official Prometheus file format.
    """
    return "groups" in rules_dict


def _is_single_alert_rule_format(rules_dict: dict) -> bool:
    """Are alert rules in single rule format.

    The Prometheus charm library supports reading of alert rules in a
    custom format that consists of a single alert rule per file. This
    does not conform to the official Prometheus alert rule file format
    which requires that each alert rules file consists of a list of
    alert rule groups and each group consists of a list of alert
    rules.

    Alert rules in dictionary form are considered to be in single rule
    format if in the least it contains two keys corresponding to the
    alert rule name and alert expression.

    Returns:
        True if alert rule is in single rule file format.
    """
    # one alert rule per file
    return set(rules_dict) >= {"alert", "expr"}


def _validate_relation_by_interface_and_direction(
    charm: CharmBase,
    relation_name: str,
    expected_relation_interface: str,
    expected_relation_role: RelationRole,
):
    """Verifies that a relation has the necessary characteristics.

    Verifies that the `relation_name` provided: (1) exists in metadata.yaml,
    (2) declares as interface the interface name passed as `relation_interface`
    and (3) has the right "direction", i.e., it is a relation that `charm`
    provides or requires.

    Args:
        charm: a `CharmBase` object to scan for the matching relation.
        relation_name: the name of the relation to be verified.
        expected_relation_interface: the interface name to be matched by the
            relation named `relation_name`.
        expected_relation_role: whether the `relation_name` must be either
            provided or required by `charm`.

    Raises:
        RelationNotFoundError: If there is no relation in the charm's metadata.yaml
            with the same name as provided via `relation_name` argument.
        RelationInterfaceMismatchError: The relation with the same name as provided
            via `relation_name` argument does not have the same relation interface
            as specified via the `expected_relation_interface` argument.
        RelationRoleMismatchError: If the relation with the same name as provided
            via `relation_name` argument does not have the same role as specified
            via the `expected_relation_role` argument.
    """
    if relation_name not in charm.meta.relations:
        raise RelationNotFoundError(relation_name)

    relation: RelationMeta = charm.meta.relations[relation_name]

    actual_relation_interface = relation.interface_name
    if actual_relation_interface != expected_relation_interface:
        raise RelationInterfaceMismatchError(
            relation_name, expected_relation_interface, actual_relation_interface or "None"
        )

    if expected_relation_role == RelationRole.provides:
        if relation_name not in charm.meta.provides:
            raise RelationRoleMismatchError(
                relation_name, RelationRole.provides, RelationRole.requires
            )
    elif expected_relation_role == RelationRole.requires:
        if relation_name not in charm.meta.requires:
            raise RelationRoleMismatchError(
                relation_name, RelationRole.requires, RelationRole.provides
            )
    else:
        raise Exception("Unexpected RelationDirection: {}".format(expected_relation_role))


class PrometheusRemoteWriteEndpointsChangedEvent(EventBase):
    """Event emitted when Prometheus remote_write endpoints change."""

    def __init__(self, handle, relation_id):
        super().__init__(handle)
        self.relation_id = relation_id

    def snapshot(self):
        """Save scrape Prometheus remote_write information."""
        return {"relation_id": self.relation_id}

    def restore(self, snapshot):
        """Restore scrape Prometheus remote_write information."""
        self.relation_id = snapshot["relation_id"]


class InvalidAlertRulePathError(Exception):
    """Raised if the alert rules folder cannot be found or is otherwise invalid."""

    def __init__(
        self,
        alert_rules_absolute_path: str,
        message: str,
    ):
        self.alert_rules_absolute_path = alert_rules_absolute_path
        self.message = message

        super().__init__(self.message)


def _resolve_dir_against_charm_path(charm: CharmBase, *path_elements: str) -> str:
    """Resolve the provided path items against the directory of the main file.

    Look up the directory of the main .py file being executed. This is normally
    going to be the charm.py file of the charm including this library. Then, resolve
    the provided path elements and, if the result path exists and is a directory,
    return its absolute path; otherwise, return `None`.
    """
    charm_dir = Path(str(charm.charm_dir))
    if not charm_dir.exists() or not charm_dir.is_dir():
        # Operator Framework does not currently expose a robust
        # way to determine the top level charm source directory
        # that is consistent across deployed charms and unit tests
        # Hence for unit tests the current working directory is used
        # TODO: updated this logic when the following ticket is resolved
        # https://github.com/canonical/operator/issues/643
        charm_dir = Path(os.getcwd())

    alerts_dir_path = charm_dir.absolute().joinpath(*path_elements)

    if not alerts_dir_path.exists():
        raise InvalidAlertRulePathError(str(alerts_dir_path), "directory does not exist")
    if not alerts_dir_path.is_dir():
        raise InvalidAlertRulePathError(str(alerts_dir_path), "is not a directory")

    return str(alerts_dir_path)


class PrometheusRemoteWriteConsumerEvents(ObjectEvents):
    """Event descriptor for events raised by `PrometheusRemoteWriteConsumer`."""

    endpoints_changed = EventSource(PrometheusRemoteWriteEndpointsChangedEvent)
    alert_rule_status_changed = EventSource(InvalidAlertRuleEvent)


class PrometheusRemoteWriteConsumer(Object):
    """API that manages a required `prometheus_remote_write` relation.

     The `PrometheusRemoteWriteConsumer` is intended to be used by charms that need to push data to
     other charms over the Prometheus remote_write API.

     The `PrometheusRemoteWriteConsumer` object can be instantiated as follows in your charm:

     ```
     from charms.prometheus_k8s.v1.prometheus_remote_write import PrometheusRemoteWriteConsumer

     def __init__(self, *args):
         ...
         self.remote_write_consumer = PrometheusRemoteWriteConsumer(self)
         ...
     ```

     The `PrometheusRemoteWriteConsumer` assumes that, in the `metadata.yaml` of your charm,
     you declare a required relation as follows:

     ```
     requires:
         send-remote-write:  # Relation name
             interface: prometheus_remote_write  # Relation interface
     ```

     The charmed operator is expected to use the `PrometheusRemoteWriteConsumer` as follows:

     ```
     def __init__(self, *args):
         ...
         self.remote_write_consumer = PrometheusRemoteWriteConsumer(self)
         ...

         self.framework.observe(
             self.remote_write_consumer.on.endpoints_changed,
             self._handle_endpoints_changed,
         )
     ```
     The `endpoints_changed` event will fire in situations such as provider ip change (e.g.
     relation created, provider upgrade, provider pod churn) or provider config change (e.g.
     metadata settings).

     Then, inside the logic of `_handle_endpoints_changed`, the updated endpoint list is
     retrieved with:

     ```
     self.remote_write_consumer.endpoints
     ```

     which returns a dictionary structured like the Prometheus configuration object (see
     https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write).

     Regarding the default relation name, `send-remote-write`: if you choose to change it,
     you would need to explicitly provide it to the `PrometheusRemoteWriteConsumer` via the
     `relation_name` constructor argument. (The relation interface, on the other hand, is
     fixed and, if you were to change it, your charm would not be able to relate with other
     charms using the correct relation interface. The library prevents you from doing that by
     raising an exception.)

     In any case, it is strongly discouraged to change the relation name: having consistent
     relation names across charms that do similar things is good practice and more
     straightforward for the users of your charm. The one exception to the rule above,
     is if your charm needs to both consume and provide a relation using the
     `prometheus_remote_write` interface, in which case changing the relation name to
     differentiate between "incoming" and "outgoing" remote write interactions is necessary.

     It is also possible to specify alert rules. By default, this library will search
     `<charm_parent_dir>/prometheus_alert_rules`, which in standard charm
     layouts resolves to `src/prometheus_alert_rules`. Each set of alert rules, grouped
     by the topology identifier, goes into a separate `*.rule` file.

     If the syntax of a rule is invalid, the `MetricsEndpointProvider` logs an error and
     does not load the particular rule.

     To avoid false positives and false negatives the library will inject label filters
     automatically in the PromQL expression. For example if the charm provides an
     alert rule with an `expr` like this one:

     ```yaml
     expr: up < 1
     ```

    it will be modified with label filters ensuring that
     the only timeseries evaluated are those scraped from this charm, and no other.


     ```yaml
     expr: up{juju_application="traefik",
              juju_charm="traefik-k8s",
              juju_model="cos",
              juju_model_uuid="b5ed878d-2671-42e8-873a-e8d58c0ec325"
           } < 1
     labels:
       juju_application: traefik
       juju_charm: traefik-k8s
       juju_model: cos
       juju_model_uuid: b5ed878d-2671-42e8-873a-e8d58c0ec325
     ```
    """

    on = PrometheusRemoteWriteConsumerEvents()  # pyright: ignore

    def __init__(
        self,
        charm: CharmBase,
        relation_name: str = DEFAULT_CONSUMER_NAME,
        alert_rules_path: str = DEFAULT_ALERT_RULES_RELATIVE_PATH,
        refresh_event: Optional[Union[BoundEvent, List[BoundEvent]]] = None,
        *,
        forward_alert_rules: bool = True,
    ):
        """API to manage a required relation with the `prometheus_remote_write` interface.

        Args:
            charm: The charm object that instantiated this class.
            relation_name: Name of the relation with the `prometheus_remote_write` interface as
                defined in metadata.yaml.
            alert_rules_path: Path of the directory containing the alert rules.
            refresh_event: an optional bound event or list of bound events which
                will be observed to re-set alerts data.
            forward_alert_rules: Flag to toggle forwarding of charmed alert rules.

        Raises:
            RelationNotFoundError: If there is no relation in the charm's metadata.yaml
                with the same name as provided via `relation_name` argument.
            RelationInterfaceMismatchError: The relation with the same name as provided
                via `relation_name` argument does not have the `prometheus_scrape` relation
                interface.
            RelationRoleMismatchError: If the relation with the same name as provided
                via `relation_name` argument does not have the `RelationRole.requires`
                role.
        """
        _validate_relation_by_interface_and_direction(
            charm, relation_name, RELATION_INTERFACE_NAME, RelationRole.requires
        )

        try:
            alert_rules_path = _resolve_dir_against_charm_path(charm, alert_rules_path)
        except InvalidAlertRulePathError as e:
            logger.debug(
                "Invalid Prometheus alert rules folder at %s: %s",
                e.alert_rules_absolute_path,
                e.message,
            )

        super().__init__(charm, relation_name)
        self._charm = charm
        self._relation_name = relation_name
        self._alert_rules_path = alert_rules_path
        self._forward_alert_rules = forward_alert_rules

        self.topology = JujuTopology.from_charm(charm)

        on_relation = self._charm.on[self._relation_name]

        self.framework.observe(on_relation.relation_joined, self._handle_endpoints_changed)
        self.framework.observe(on_relation.relation_changed, self._handle_endpoints_changed)
        self.framework.observe(on_relation.relation_departed, self._handle_endpoints_changed)
        self.framework.observe(on_relation.relation_broken, self._on_relation_broken)
        self.framework.observe(on_relation.relation_joined, self._push_alerts_on_relation_joined)
        self.framework.observe(
            self._charm.on.leader_elected, self._push_alerts_to_all_relation_databags
        )
        self.framework.observe(
            self._charm.on.upgrade_charm, self._push_alerts_to_all_relation_databags
        )
        if refresh_event:
            if not isinstance(refresh_event, list):
                refresh_event = [refresh_event]
            for ev in refresh_event:
                self.framework.observe(ev, self._push_alerts_to_all_relation_databags)

    def _on_relation_broken(self, event: RelationBrokenEvent) -> None:
        self.on.endpoints_changed.emit(relation_id=event.relation.id)

    def _handle_endpoints_changed(self, event: RelationEvent) -> None:
        if self._charm.unit.is_leader() and event.app is not None:
            ev = json.loads(event.relation.data[event.app].get("event", "{}"))

            if ev:
                valid = bool(ev.get("valid", True))
                errors = ev.get("errors", "")

                if valid and not errors:
                    self.on.alert_rule_status_changed.emit(valid=valid)
                else:
                    self.on.alert_rule_status_changed.emit(valid=valid, errors=errors)

        self.on.endpoints_changed.emit(relation_id=event.relation.id)

    def _push_alerts_on_relation_joined(self, event: RelationEvent) -> None:
        self._push_alerts_to_relation_databag(event.relation)

    def _push_alerts_to_all_relation_databags(self, _: Optional[HookEvent]) -> None:
        for relation in self.model.relations[self._relation_name]:
            self._push_alerts_to_relation_databag(relation)

    def _push_alerts_to_relation_databag(self, relation: Relation) -> None:
        if not self._charm.unit.is_leader():
            return

        alert_rules = AlertRules(query_type="promql", topology=self.topology)
        if self._forward_alert_rules:
            alert_rules.add_path(self._alert_rules_path)
            alert_rules.add(
                generic_alert_groups.aggregator_rules, group_name_prefix=self.topology.identifier
            )

        alert_rules_as_dict = alert_rules.as_dict()

        relation.data[self._charm.app]["alert_rules"] = json.dumps(alert_rules_as_dict)

    def reload_alerts(self) -> None:
        """Reload alert rules from disk and push to relation data."""
        self._push_alerts_to_all_relation_databags(None)

    @property
    def endpoints(self) -> List[Dict[str, str]]:
        """A config object ready to be dropped into a prometheus config file.

        The format of the dict is specified in the official prometheus docs:
        https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write

        Returns:
            A list of dictionaries where each dictionary provides information about
            a single remote_write endpoint.
        """
        endpoints = []
        for relation in self.model.relations[self._relation_name]:
            for unit in relation.units:
                if unit.app is self._charm.app:
                    # This is a peer unit
                    continue

                remote_write = relation.data[unit].get("remote_write")
                if remote_write:
                    deserialized_remote_write = json.loads(remote_write)
                    endpoints.append(
                        {
                            "url": deserialized_remote_write["url"],
                        }
                    )

        return endpoints


class PrometheusRemoteWriteAlertsChangedEvent(EventBase):
    """Event emitted when Prometheus remote_write alerts change."""

    def __init__(self, handle, relation_id):
        super().__init__(handle)
        self.relation_id = relation_id

    def snapshot(self):
        """Save Prometheus remote_write information."""
        return {"relation_id": self.relation_id}

    def restore(self, snapshot):
        """Restore Prometheus remote_write information."""
        self.relation_id = snapshot["relation_id"]


class PrometheusRemoteWriteProviderConsumersChangedEvent(EventBase):
    """Event emitted when Prometheus remote_write alerts change."""


class PrometheusRemoteWriteProviderEvents(ObjectEvents):
    """Event descriptor for events raised by `PrometheusRemoteWriteProvider`."""

    alert_rules_changed = EventSource(PrometheusRemoteWriteAlertsChangedEvent)
    consumers_changed = EventSource(PrometheusRemoteWriteProviderConsumersChangedEvent)


class PrometheusRemoteWriteProvider(Object):
    """API that manages a provided `prometheus_remote_write` relation.

    The `PrometheusRemoteWriteProvider` is intended to be used by charms whose workloads need
    to receive data from other charms' workloads over the Prometheus remote_write API.

    The `PrometheusRemoteWriteProvider` object can be instantiated as follows in your charm:

    ```
    from charms.prometheus_k8s.v1.prometheus_remote_write import PrometheusRemoteWriteProvider

    def __init__(self, *args):
        ...
        self.remote_write_provider = PrometheusRemoteWriteProvider(self)
        ...
    ```

    The `PrometheusRemoteWriteProvider` assumes that, in the `metadata.yaml` of your charm,
    you declare a provided relation as follows:

    ```
    provides:
        receive-remote-write:  # Relation name
            interface: prometheus_remote_write  # Relation interface
    ```

    About the name of the relation managed by this library: technically, you *could* change
    the relation name, `receive-remote-write`, but that requires you to provide the new
    relation name to the `PrometheusRemoteWriteProducer` via the `relation_name` constructor
    argument. (The relation interface, on the other hand, is immutable and, if you were to change
    it, your charm would not be able to relate with other charms using the right relation
    interface. The library prevents you from doing that by raising an exception.) In any case, it
    is strongly discouraged to change the relation name: having consistent relation names across
    charms that do similar things is a very good thing for the people that will use your charm.
    The one exception to the rule above, is if you charm needs to both consume and provide a
    relation using the `prometheus_remote_write` interface, in which case changing the relation
    name to differentiate between "incoming" and "outgoing" remote write interactions is necessary.
    """

    on = PrometheusRemoteWriteProviderEvents()  # pyright: ignore

    def __init__(
        self,
        charm: CharmBase,
        relation_name: str = DEFAULT_RELATION_NAME,
        *,
        server_url_func: Callable[[], str] = lambda: f"http://{socket.getfqdn()}:9090",
        endpoint_path: str = "/api/v1/write",
    ):
        """API to manage a provided relation with the `prometheus_remote_write` interface.

        Args:
            charm: The charm object that instantiated this class.
            relation_name: Name of the relation with the `prometheus_remote_write` interface as
                defined in metadata.yaml.
            server_url_func: A callable returning the URL for your prometheus server.
            endpoint_path: The path of the server's remote_write endpoint.

        Raises:
            RelationNotFoundError: If there is no relation in the charm's metadata.yaml
                with the same name as provided via `relation_name` argument.
            RelationInterfaceMismatchError: The relation with the same name as provided
                via `relation_name` argument does not have the `prometheus_scrape` relation
                interface.
            RelationRoleMismatchError: If the relation with the same name as provided
                via `relation_name` argument does not have the `RelationRole.requires`
                role.
        """
        _validate_relation_by_interface_and_direction(
            charm, relation_name, RELATION_INTERFACE_NAME, RelationRole.provides
        )

        super().__init__(charm, relation_name)
        self._charm = charm
        self._tool = CosTool(self._charm)
        self._relation_name = relation_name
        self._get_server_url = server_url_func
        self._endpoint_path = endpoint_path

        on_relation = self._charm.on[self._relation_name]
        self.framework.observe(
            on_relation.relation_created,
            self._on_consumers_changed,
        )
        self.framework.observe(
            on_relation.relation_joined,
            self._on_consumers_changed,
        )
        self.framework.observe(
            on_relation.relation_changed,
            self._on_relation_changed,
        )

    def _on_consumers_changed(self, event: RelationEvent) -> None:
        if not isinstance(event, RelationBrokenEvent):
            self.update_endpoint(event.relation)
        self.on.consumers_changed.emit()

    def _on_relation_changed(self, event: RelationEvent) -> None:
        """Flag Providers that data has changed, so they can re-read alerts."""
        self.on.alert_rules_changed.emit(event.relation.id)

    def update_endpoint(self, relation: Optional[Relation] = None) -> None:
        """Triggers programmatically the update of the relation data.

        This method should be used when the charm relying on this library needs
        to update the relation data in response to something occurring outside
        the `prometheus_remote_write` relation lifecycle, e.g., in case of a
        host address change because the charmed operator becomes connected to an
        Ingress after the `prometheus_remote_write` relation is established.

        Args:
            relation: An optional instance of `class:ops.model.Relation` to update.
                If not provided, all instances of the `prometheus_remote_write`
                relation are updated.
        """
        relations = [relation] if relation else self.model.relations[self._relation_name]

        for relation in relations:
            self._set_endpoint_on_relation(relation)

    def _set_endpoint_on_relation(self, relation: Relation) -> None:
        """Set the remote_write endpoint on relations.

        Args:
            relation: The relation whose data to update.
        """
        relation.data[self._charm.unit]["remote_write"] = json.dumps(
            {
                "url": self._get_server_url().rstrip("/") + "/" + self._endpoint_path.strip("/"),
            }
        )

    @property
    def alerts(self) -> dict:
        """Fetch alert rules from all relations.

        A Prometheus alert rules file consists of a list of "groups". Each
        group consists of a list of alerts (`rules`) that are sequentially
        executed. This method returns all the alert rules provided by each
        related metrics provider charm. These rules may be used to generate a
        separate alert rules file for each relation since the returned list
        of alert groups are indexed by relation ID. Also, for each relation ID
        associated scrape metadata such as Juju model, UUID and application
        name are provided so the unique name may be generated for the rules
        file. For each relation the structure of data returned is a dictionary
        with four keys

        - groups
        - model
        - model_uuid
        - application

        The value of the `groups` key is such that it may be used to generate
        a Prometheus alert rules file directly using `yaml.dump` but the
        `groups` key itself must be included as this is required by Prometheus,
        for example as in `yaml.safe_dump({"groups": alerts["groups"]})`.

        The `PrometheusRemoteWriteProvider` accepts a list of rules and these
        rules are all placed into one group.

        Returns:
            a dictionary mapping the name of an alert rule group to the group.
        """
        alerts = {}  # type: Dict[str, dict] # mapping b/w juju identifiers and alert rule files
        for relation in self._charm.model.relations[self._relation_name]:
            if not relation.units or not relation.app:
                continue

            alert_rules = json.loads(relation.data[relation.app].get("alert_rules", "{}"))
            if not alert_rules:
                continue

            alert_rules = self._inject_alert_expr_labels(alert_rules)

            identifier, topology = self._get_identifier_by_alert_rules(alert_rules)
            if not topology:
                try:
                    scrape_metadata = json.loads(relation.data[relation.app]["scrape_metadata"])
                    identifier = JujuTopology.from_dict(scrape_metadata).identifier
                    alerts[identifier] = self._tool.apply_label_matchers(alert_rules)  # type: ignore

                except KeyError as e:
                    logger.debug(
                        "Relation %s has no 'scrape_metadata': %s",
                        relation.id,
                        e,
                    )

            if not identifier:
                logger.error(
                    "Alert rules were found but no usable group or identifier was present."
                )
                continue

            _, errmsg = self._tool.validate_alert_rules(alert_rules)
            if errmsg:
                logger.error(f"Invalid alert rule file: {errmsg}")
                if self._charm.unit.is_leader():
                    data = json.loads(relation.data[self._charm.app].get("event", "{}"))
                    data["errors"] = errmsg
                    relation.data[self._charm.app]["event"] = json.dumps(data)
                continue

            alerts[identifier] = alert_rules

        return alerts

    def _get_identifier_by_alert_rules(
        self, rules: Dict[str, Any]
    ) -> Tuple[Union[str, None], Union[JujuTopology, None]]:
        """Determine an appropriate dict key for alert rules.

        The key is used as the filename when writing alerts to disk, so the structure
        and uniqueness is important.

        Args:
            rules: a dict of alert rules
        Returns:
            A tuple containing an identifier, if found, and a JujuTopology, if it could
            be constructed.
        """
        if "groups" not in rules:
            logger.debug("No alert groups were found in relation data")
            return None, None

        # Construct an ID based on what's in the alert rules if they have labels
        for group in rules["groups"]:
            try:
                labels = group["rules"][0]["labels"]
                topology = JujuTopology(
                    # Don't try to safely get required constructor fields. There's already
                    # a handler for KeyErrors
                    model_uuid=labels["juju_model_uuid"],
                    model=labels["juju_model"],
                    application=labels["juju_application"],
                    unit=labels.get("juju_unit", ""),
                    charm_name=labels.get("juju_charm", ""),
                )
                return topology.identifier, topology
            except KeyError:
                logger.debug("Alert rules were found but no usable labels were present")
                continue

        logger.warning(
            "No labeled alert rules were found, and no 'scrape_metadata' "
            "was available. Using the alert group name as filename."
        )
        try:
            for group in rules["groups"]:
                return group["name"], None
        except KeyError:
            logger.debug("No group name was found to use as identifier")

        return None, None

    def _inject_alert_expr_labels(self, rules: Dict[str, Any]) -> Dict[str, Any]:
        """Iterate through alert rules and inject topology into expressions.

        Args:
            rules: a dict of alert rules
        """
        if "groups" not in rules:
            return rules

        modified_groups = []
        for group in rules["groups"]:
            # Copy off rules, so we don't modify an object we're iterating over
            rules_copy = group["rules"]
            for idx, rule in enumerate(rules_copy):
                labels = rule.get("labels")

                if labels:
                    try:
                        topology = JujuTopology(
                            # Don't try to safely get required constructor fields. There's already
                            # a handler for KeyErrors
                            model_uuid=labels["juju_model_uuid"],
                            model=labels["juju_model"],
                            application=labels["juju_application"],
                            unit=labels.get("juju_unit", ""),
                            charm_name=labels.get("juju_charm", ""),
                        )

                        # Inject topology and put it back in the list
                        rule["expr"] = self._tool.inject_label_matchers(
                            re.sub(r"%%juju_topology%%,?", "", rule["expr"]),
                            topology.alert_expression_dict,
                        )
                    except KeyError:
                        # Some required JujuTopology key is missing. Just move on.
                        pass

                    group["rules"][idx] = rule

            modified_groups.append(group)

        rules["groups"] = modified_groups
        return rules


# Copy/pasted from prometheus_scrape.py
class CosTool:
    """Uses cos-tool to inject label matchers into alert rule expressions and validate rules."""

    _path = None
    _disabled = False

    def __init__(self, charm):
        self._charm = charm

    @property
    def path(self):
        """Lazy lookup of the path of cos-tool."""
        if self._disabled:
            return None
        if not self._path:
            self._path = self._get_tool_path()
            if not self._path:
                logger.debug("Skipping injection of juju topology as label matchers")
                self._disabled = True
        return self._path

    def apply_label_matchers(self, rules) -> dict:
        """Will apply label matchers to the expression of all alerts in all supplied groups."""
        if not self.path:
            return rules
        for group in rules["groups"]:
            rules_in_group = group.get("rules", [])
            for rule in rules_in_group:
                topology = {}
                # if the user for some reason has provided juju_unit, we'll need to honor it
                # in most cases, however, this will be empty
                for label in [
                    "juju_model",
                    "juju_model_uuid",
                    "juju_application",
                    "juju_charm",
                    "juju_unit",
                ]:
                    if label in rule["labels"]:
                        topology[label] = rule["labels"][label]

                rule["expr"] = self.inject_label_matchers(rule["expr"], topology)
        return rules

    def validate_alert_rules(self, rules: dict) -> Tuple[bool, str]:
        """Will validate correctness of alert rules, returning a boolean and any errors."""
        if not self.path:
            logger.debug("`cos-tool` unavailable. Not validating alert correctness.")
            return True, ""

        with tempfile.TemporaryDirectory() as tmpdir:
            rule_path = Path(tmpdir + "/validate_rule.yaml")
            rule_path.write_text(yaml.dump(rules))

            args = [str(self.path), "validate", str(rule_path)]
            # noinspection PyBroadException
            try:
                self._exec(args)
                return True, ""
            except subprocess.CalledProcessError as e:
                logger.debug("Validating the rules failed: %s", e.output)
                return False, ", ".join(
                    [
                        line
                        for line in e.output.decode("utf8").splitlines()
                        if "error validating" in line
                    ]
                )

    def inject_label_matchers(self, expression, topology) -> str:
        """Add label matchers to an expression."""
        if not topology:
            return expression
        if not self.path:
            logger.debug("`cos-tool` unavailable. Leaving expression unchanged: %s", expression)
            return expression
        args = [str(self.path), "transform"]
        args.extend(
            ["--label-matcher={}={}".format(key, value) for key, value in topology.items()]
        )

        args.extend(["{}".format(expression)])
        # noinspection PyBroadException
        try:
            return self._exec(args)
        except subprocess.CalledProcessError as e:
            logger.debug('Applying the expression failed: "%s", falling back to the original', e)
            return expression

    def _get_tool_path(self) -> Optional[Path]:
        arch = platform.machine()
        arch = "amd64" if arch == "x86_64" else arch
        res = "cos-tool-{}".format(arch)
        try:
            path = Path(res).resolve(strict=True)
            return path
        except (FileNotFoundError, OSError):
            logger.debug('Could not locate cos-tool at: "{}"'.format(res))
        return None

    def _exec(self, cmd) -> str:
        result = subprocess.run(cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        return result.stdout.decode("utf-8").strip()
