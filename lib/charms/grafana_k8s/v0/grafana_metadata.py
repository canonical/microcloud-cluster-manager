"""grafana_metadata.

This library implements endpoint wrappers for the grafana-metadata interface.  The grafana-metadata interface is used to
transfer information about an instance of Grafana, such as how to access and uniquely identify it.  Typically, this is
useful for charms that operate a Grafana instance to give other applications access to its API.

## Usage

### Requirer

GrafanaMetadataRequirer is a wrapper for pulling data from the grafana-metadata interface.  To use it in your charm:

* observe the relation-changed event for this relation wherever your charm needs to use this data (this endpoint wrapper
  DOES NOT automatically observe any events)
* wherever you need access to the data, call `GrafanaMetadataRequirer(...).get_data()`

An example implementation is:

```python
class FooCharm(CharmBase):
    def __init__(self, framework):
        super().__init__(framework)

        self.grafana_metadata = GrafanaMetadataRequirer(self.model.relations, "grafana-metadata")

        self.framework.observe(self.on["grafana-metadata"].relation_changed, self._on_grafana_metadata_changed)

    def _on_grafana_metadata_changed(self):
        data = self.grafana_metadata.get_data()
        ...
```

Where you also add relation to your `charmcraft.yaml` or `metadata.yaml` (note that GrafanaMetadataRequirer is designed
for relating to a single application and must be used with limit=1 as shown below):

```yaml
requires:
  grafana-metadata:
    limit: 1
    interface: grafana_metadata
```

### Provider

GrafanaMetadataProvider is a wrapper for publishing data to charms related using the grafana-metadata interface.  Note
that `GrafanaMetadataProvider` *does not* manage any events, but instead provides a `publish` method for sending data to
all related applications.  Triggering `publish` appropriately is left to the charm author, although generally you want
to do this at least during the `relation_joined` and `leader_elected` events.  An example implementation is:

```python
class FooCharm(CharmBase):
    def __init__(self, framework):
        super().__init__(framework)
        self.grafana_metadata = GrafanaMetadataProvider(
            relations=self.model.relations,
            relation_name="grafana-metadata",
            app=self.app,
        )

        self.framework.observe(self.on.leader_elected, self.do_something_to_publish)
        self.framework.observe(self._charm.on["grafana-metadata"].relation_joined, self.do_something_to_publish)
        self.framework.observe(self.on.some_event_that_changes_grafana_metadata, self.do_something_to_publish)

    def do_something_to_publish(self, e):
        self.grafana_metadata.publish(...)
```

Where you also add the following to your `charmcraft.yaml` or `metadata.yaml`:

```yaml
provides:
  grafana-metadata:
    interface: grafana_metadata
```
"""

import logging
from typing import Optional, Union, List

from ops import RelationMapping, Application, Relation
from pydantic import AnyHttpUrl, BaseModel, Field, AfterValidator
from typing_extensions import Annotated

# The unique Charmhub library identifier, never change it
LIBID = "cee305b7fe0546548cbfe81c3bd20e33"

# Increment this major API version when introducing breaking changes
LIBAPI = 0

# Increment this PATCH version before using `charmcraft publish-lib` or reset
# to 0 if you are raising the major API version
LIBPATCH = 3

PYDEPS = ["pydantic>=2"]

log = logging.getLogger(__name__)

DEFAULT_RELATION_NAME = "grafana-metadata"

# Define a custom type that accepts AnyHttpUrl and string, but converts to AnyHttpUrl and raises an exception if the
# string is not a valid URL
AnyHttpUrlOrStrUrl = Annotated[Union[AnyHttpUrl, str], AfterValidator(lambda v: AnyHttpUrl(v))]


class GrafanaMetadataAppData(BaseModel):
    """Data model for the grafana-metadata interface."""

    ingress_url: Optional[AnyHttpUrlOrStrUrl] = Field(
        default=None,
        description="The non-internal URL at which this application can be reached.  Typically, this is an ingress URL.",
    )
    direct_url: AnyHttpUrlOrStrUrl = Field(
        description="The cluster-internal URL at which this application can be reached.  Typically, this is a"
        " Kubernetes FQDN like name.namespace.svc.cluster.local for connecting to the prometheus api"
        " from inside the cluster, with scheme."
    )
    grafana_uid: str = Field(description="The UID of this Grafana application.")


class GrafanaMetadataRequirer:
    """Endpoint wrapper for the requirer side of the grafana-metadata relation."""

    def __init__(
        self,
        relation_mapping: RelationMapping,
        relation_name: str = DEFAULT_RELATION_NAME,
    ) -> None:
        """Initialize the GrafanaMetadataRequirer object.

        This object is for accessing data from relations that use the grafana-metadata interface.  It **does not**
        autonomously handle the events associated with that relation.  It is up to the charm using this object to
        observe those events as they see fit.  Typically, that charm should observe this relation's relation-changed
        event.

        This object is for interacting with a relation that has limit=1 set in charmcraft.yaml.  In particular, the
        get_data method will raise if more than one related application is available.

        Args:
            relation_mapping: The RelationMapping of a charm (typically `self.model.relations` from within a charm
                              object).
            relation_name: The name of the wrapped relation.
        """
        self._charm_relation_mapping = relation_mapping
        self._relation_name = relation_name

    @property
    def relations(self) -> List[Relation]:
        """Return the relation instances for applications related to us on the monitored relation."""
        return self._charm_relation_mapping.get(self._relation_name, [])

    def get_data(self) -> Optional[BaseModel]:
        """Return data for at most one related application, raising if more than one is available.

        Useful for charms that always expect exactly one related application.  It is recommended that those charms also
        set limit=1 for that relation in charmcraft.yaml.  Returns None if no data is available (either because no
        applications are related to us, or because the related application has not sent data).
        """
        relations = self.relations
        if len(relations) == 0:
            return None
        if len(relations) > 1:
            raise ValueError("Cannot get_info when more than one application is related.")

        # Being a little cautious here using getattr and get, since some funny things have happened with relation data
        # in the past.
        raw_data_dict = getattr(relations[0], "data", {}).get(relations[0].app)
        if not raw_data_dict:
            return None

        return GrafanaMetadataAppData.model_validate(raw_data_dict)


class GrafanaMetadataProvider:
    """The provider side of the grafana-metadata relation."""

    def __init__(
        self,
        relation_mapping: RelationMapping,
        app: Application,
        relation_name: str = DEFAULT_RELATION_NAME,
    ):
        """Initialize the GrafanaMetadataProvider object.

        This object is for serializing and sending data to a relation that uses the grafana-metadata interface - it does
        not automatically observe any events for that relation.  It is up to the charm using this to call publish when
        it is appropriate to do so, typically on at least the charm's leader_elected event and this relation's
        relation_joined event.

        Args:
            relation_mapping: The RelationMapping of a charm (typically `self.model.relations` from within a charm object).
            app: This application.
            relation_name: The name of the relation.
        """
        self._charm_relation_mapping = relation_mapping
        self._app = app
        self._relation_name = relation_name

    @property
    def relations(self):
        """Return the applications related to us under the monitored relation."""
        return self._charm_relation_mapping.get(self._relation_name, ())

    def publish(
        self, grafana_uid: str, direct_url: AnyHttpUrlOrStrUrl, ingress_url: Optional[AnyHttpUrlOrStrUrl] = None
    ):
        """Post grafana-metadata to all related applications.

        This method writes to the relation's app data bag, and thus should never be called by a unit that is not the
        leader otherwise ops will raise an exception.

        Args:
            grafana_uid: The UID of this Grafana instance.
            direct_url: The cluster-internal URL at which this application can be reached.  Typically, this is a
                        Kubernetes FQDN like name.namespace.svc.cluster.local for connecting to the prometheus api
                        from inside the cluster, with scheme.
            ingress_url: The non-internal URL at which this application can be reached.  Typically, this is an ingress
                         URL.
        """
        data = GrafanaMetadataAppData(
            grafana_uid=grafana_uid, direct_url=direct_url, ingress_url=ingress_url
        ).model_dump(mode="json", by_alias=True, exclude_defaults=True, round_trip=True)

        for relation in self.relations:
            databag = relation.data[self._app]
            databag.update(data)
