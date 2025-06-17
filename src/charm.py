#!/usr/bin/env python3
# Copyright 2025 Canonical Ltd.
# See LICENSE file for licensing details.

import logging
import ops

from charms.data_platform_libs.v0.data_interfaces import DatabaseRequires
from charms.grafana_k8s.v0.grafana_metadata import GrafanaMetadataAppData, GrafanaMetadataRequirer
from charms.prometheus_k8s.v1.prometheus_remote_write import PrometheusRemoteWriteConsumer
from charms.tls_certificates_interface.v4.tls_certificates import (
    Certificate,
    CertificateRequestAttributes,
    Mode,
    PrivateKey,
    TLSCertificatesRequiresV4,
)
from typing import Optional

CERTS_DIR_PATH = "/etc/ssl/management-api"
PRIVATE_KEY_NAME = "tls.key"
CERTIFICATE_NAME = "tls.crt"
CA_NAME = "ca.crt"

logger = logging.getLogger(__name__)


class ClusterManagerCharm(ops.CharmBase):
    def __init__(self, framework: ops.Framework) -> None:
        super().__init__(framework)
        self.container = self.unit.get_container("microcloud-cluster-manager")

        # TLS certificates relation
        self.certificates = TLSCertificatesRequiresV4(
            charm=self,
            relationship_name="certificates",
            certificate_requests=[self._get_certificate_request_attributes()],
            mode=Mode.UNIT,
        )
        framework.observe(self.certificates.on.certificate_available, self._on_certificates_available)

        # Database relation
        self.database = DatabaseRequires(self, relation_name="database", database_name="cluster_manager_db")
        framework.observe(self.database.on.database_created, self._on_database_changed)
        framework.observe(self.database.on.endpoints_changed, self._on_database_changed)

        # Prometheus remote write relation
        self.prometheus_remote_write = PrometheusRemoteWriteConsumer(self)
        self.framework.observe(self.prometheus_remote_write.on.endpoints_changed, self._on_endpoints_changed)

        # Grafana metadata relation
        self.grafana_metadata = GrafanaMetadataRequirer(relation_mapping=self.model.relations)
        self.framework.observe(self.on["grafana-metadata"].relation_changed, self._on_grafana_metadata_changed)

        # Main service
        self.pebble_service_name = "mcm-management-api"
        framework.observe(self.on.collect_unit_status, self._on_collect_status)
        framework.observe(self.on.microcloud_cluster_manager_pebble_ready,
                          self._on_microcloud_cluster_manager_pebble_ready)

    def _get_certificate_request_attributes(self) -> CertificateRequestAttributes:
        return CertificateRequestAttributes(
            common_name="cc.lxd-cm.local",
            sans_dns=frozenset(["cc.lxd-cm.local"]),
        )

    def _on_collect_status(self, event: ops.CollectStatusEvent):
        if not self._is_relation_created("certificates"):
            event.add_status(
                ops.BlockedStatus("certificates integration not created")
            )
            return
        if not self.model.get_relation('database'):
            event.add_status(ops.BlockedStatus('Waiting for database relation'))
            return
        if not self.database.fetch_relation_data():
            event.add_status(ops.WaitingStatus('Waiting for database relation'))
            return
        try:
            status = self.container.get_service(self.pebble_service_name)
        except (ops.pebble.APIError, ops.pebble.ConnectionError, ops.ModelError):
            event.add_status(ops.MaintenanceStatus('Waiting for Pebble in workload container'))
        else:
            if not status.is_running():
                event.add_status(ops.MaintenanceStatus('Waiting for the service to start up'))
        # If nothing is wrong, then the status is active.
        event.add_status(ops.ActiveStatus())

    def _on_certificates_available(self, _: ops.EventBase):
        if not self._is_relation_created("certificates"):
            return
        if not self._is_certificate_available():
            return
        certificate_update_required = self._check_and_update_certificate()
        if certificate_update_required:
            logger.info("Certificates configured.")
            self.update_pebble_layer()

    def _is_relation_created(self, relation_name: str) -> bool:
        return bool(self.model.relations.get(relation_name))

    def _is_certificate_available(self) -> bool:
        cert, key = self.certificates.get_assigned_certificate(
            certificate_request=self._get_certificate_request_attributes()
        )
        return bool(cert and key)

    def _check_and_update_certificate(self) -> bool:
        """Check if the certificate or private key needs an update and perform the update.

        This method retrieves the currently assigned certificate and private key associated with
        the charm's TLS relation. It checks whether the certificate or private key has changed
        or needs to be updated. If an update is necessary, the new certificate or private key is
        stored.

        Returns:
            bool: True if either the certificate or the private key was updated, False otherwise.
        """
        provider_certificate, private_key = self.certificates.get_assigned_certificate(
            certificate_request=self._get_certificate_request_attributes()
        )
        if not provider_certificate or not private_key:
            logger.debug("Certificate or private key is not available")
            return False
        if certificate_update_required := self._is_certificate_update_required(
                provider_certificate.certificate
        ):
            self._store_certificate(certificate=provider_certificate.certificate)
        if private_key_update_required := self._is_private_key_update_required(private_key):
            self._store_private_key(private_key=private_key)
        return certificate_update_required or private_key_update_required

    def _is_certificate_update_required(self, certificate: Certificate) -> bool:
        return self._get_existing_certificate() != certificate

    def _is_private_key_update_required(self, private_key: PrivateKey) -> bool:
        return self._get_existing_private_key() != private_key

    def _get_existing_certificate(self) -> Optional[Certificate]:
        return self._get_stored_certificate() if self._certificate_is_stored() else None

    def _get_existing_private_key(self) -> Optional[PrivateKey]:
        return self._get_stored_private_key() if self._private_key_is_stored() else None

    def _certificate_is_stored(self) -> bool:
        return self.container.exists(path=f"{CERTS_DIR_PATH}/{CERTIFICATE_NAME}")

    def _private_key_is_stored(self) -> bool:
        return self.container.exists(path=f"{CERTS_DIR_PATH}/{PRIVATE_KEY_NAME}")

    def _get_stored_certificate(self) -> Certificate:
        cert_string = str(self.container.pull(path=f"{CERTS_DIR_PATH}/{CERTIFICATE_NAME}").read())
        return Certificate.from_string(cert_string)

    def _get_stored_private_key(self) -> PrivateKey:
        key_string = str(self.container.pull(path=f"{CERTS_DIR_PATH}/{PRIVATE_KEY_NAME}").read())
        return PrivateKey.from_string(key_string)

    def _store_certificate(self, certificate: Certificate) -> None:
        """Store certificate in workload."""
        if not self.container.exists(path=f"{CERTS_DIR_PATH}"):
            self.container.make_dir(path=CERTS_DIR_PATH, make_parents=True)
            logger.info(f"Created directory {CERTS_DIR_PATH} in workload")
        self.container.push(path=f"{CERTS_DIR_PATH}/{CERTIFICATE_NAME}", source=str(certificate))
        self.container.push(path=f"{CERTS_DIR_PATH}/{CA_NAME}", source=str(certificate))
        logger.info("Pushed certificate pushed to workload")

    def _store_private_key(self, private_key: PrivateKey) -> None:
        """Store private key in workload."""
        if not self.container.exists(path=f"{CERTS_DIR_PATH}"):
            self.container.make_dir(path=CERTS_DIR_PATH, make_parents=True)
            logger.info(f"Created directory {CERTS_DIR_PATH} in workload")
        self.container.push(
            path=f"{CERTS_DIR_PATH}/{PRIVATE_KEY_NAME}",
            source=str(private_key),
        )
        logger.info("Pushed private key to workload")

    def _on_microcloud_cluster_manager_pebble_ready(self, _: ops.PebbleReadyEvent) -> None:
        logger.info("Pebble is ready.")
        self.update_pebble_layer()

    def _on_endpoints_changed(self, _: ops.EventBase):
        logger.info("Prometheus endpoints changed.")
        self.update_pebble_layer()

    def _on_grafana_metadata_changed(self, _: ops.EventBase):
        logger.info("Grafana metadata changed.")
        self.update_pebble_layer()

    def _on_database_changed(self, _: ops.EventBase) -> None:
        logger.info("Postgres database created or updated.")
        self.update_pebble_layer()

    def update_pebble_layer(self) -> None:
        if not self._certificate_is_stored():
            logger.info("TLS certificate is not stored, skipping Pebble layer update.")
            self.unit.status = ops.WaitingStatus("Waiting for TLS certificate to be available")
            return
        if not self.get_postgres_relation_data():
            logger.info("Postgres relation data is not available, skipping Pebble layer update.")
            self.unit.status = ops.WaitingStatus("Waiting for database relation to be available")
            return

        self.container.add_layer('microcloud_cluster_manager', self._pebble_layer, combine=True)
        self.container.replan()
        self.unit.status = ops.ActiveStatus()

    @property
    def _pebble_layer(self) -> ops.pebble.Layer:
        mgmt_environment = self.app_environment
        mgmt_environment['SERVICE'] = 'management-api'
        mgmt_environment['SERVER_PORT'] = '9100'
        mgmt_environment['STATUS_PORT'] = '11000'
        mgmt_service = ops.pebble.ServiceDict(override="replace",
                                              summary="microcloud cluster manager management-api service",
                                              command="microcloud-cluster-manager",
                                              startup="enabled",
                                              environment=mgmt_environment)

        cluster_environment = self.app_environment
        cluster_environment['SERVICE'] = 'cluster-connector'
        cluster_service = ops.pebble.ServiceDict(override="replace",
                                                 summary="microcloud cluster manager cluster-connector service",
                                                 command="microcloud-cluster-manager",
                                                 startup="enabled",
                                                 environment=cluster_environment)

        pebble_layer = ops.pebble.LayerDict(summary='microcloud cluster manager services',
                                            description='cluster connector and management api services',
                                            services={"mcm-management-api": mgmt_service,
                                                      "mcm-cluster-connector": cluster_service})

        return ops.pebble.Layer(pebble_layer)

    def get_prometheus_write_endpoint(self):
        """Return a sorted list of remote-write endpoints."""
        if not hasattr(self, "prometheus_remote_write"):
            return None

        endpoints = getattr(self.prometheus_remote_write, "endpoints", None)
        if not isinstance(endpoints, list) or not endpoints:
            return None

        first = endpoints[0]
        if not isinstance(first, dict):
            return None

        return first.get("url")

    def get_grafana_metadata(self) -> GrafanaMetadataAppData:
        """Return the metadata for the related Grafana."""
        return self.grafana_metadata.get_data()

    def get_postgres_relation_data(self) -> dict[str, str]:
        """Fetch postgres relation data."""
        relations = self.database.fetch_relation_data()
        logger.debug('Got following database data: %s', relations)
        for data in relations.values():
            if not data:
                continue
            logger.debug('New PSQL database endpoint is %s', data['endpoints'])
            host, port = data['endpoints'].split(':')
            db_data = {
                'db_host': host,
                'db_port': port,
                'db_username': data['username'],
                'db_password': data['password'],
                'db_name': data['database'],
            }
            return db_data
        return {}

    @property
    def app_environment(self) -> dict[str, str]:
        db_data = self.get_postgres_relation_data()
        prometheus_url = self.get_prometheus_write_endpoint()

        grafana_metadata = self.get_grafana_metadata()
        # todo this should come from grafana relation
        grafana_url = "http://0.0.0.0:3000/d/bGY-LSB7k/lxd?orgId=1"
        if grafana_metadata:
            grafana_internal_url = str(grafana_metadata.direct_url)
            grafana_external_url = str(grafana_metadata.ingress_url)
            grafana_uid = grafana_metadata.grafana_uid

            logger.info("Grafana metadata: %s", grafana_metadata)
            logger.info("Grafana internal URL: %s", grafana_internal_url)
            logger.info("Grafana external URL: %s", grafana_external_url)
            logger.info("Grafana UID: %s", grafana_uid)

        env = {
            key: value
            for key, value in {
                "CLUSTER_CONNECTOR_ADDRESS": "cc.lxd-cm.local:32000",
                "CLUSTER_CONNECTOR_TLS_PATH": CERTS_DIR_PATH,
                "MANAGEMENT_API_TLS_PATH": CERTS_DIR_PATH,
                "DB_DISABLE_TLS": "true",
                "DB_HOST": db_data.get("db_host", None),
                "DB_PORT": db_data.get("db_port", None),
                "DB_USER": db_data.get("db_username", None),
                "DB_PASSWORD": db_data.get("db_password", None),
                "DB_NAME": db_data.get("db_name", None),
                "DB_MAX_IDLE": "2",
                "DB_MAX_OPEN": "5",
                "OIDC_AUDIENCE": "https://lxd-ui-demo.us.auth0.com/api/v2/",
                "OIDC_CLIENT_ID": "OZSAeCbqAXZid3LL1gRQEkLXP9KlwZtJ",
                "OIDC_ISSUER": "https://lxd-ui-demo.us.auth0.com/",
                "PROMETHEUS_BASE_URL": prometheus_url or "",
                "GRAFANA_BASE_URL": grafana_url or "",
                "SERVER_HOST": "0.0.0.0",
                "SERVER_PORT": "9000",
                "TEST_MODE": "false",
                "VERSION": "development",
            }.items()
            if value is not None
        }
        return env

if __name__ == "__main__":  # pragma: nocover
    ops.main(ClusterManagerCharm)
