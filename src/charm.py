#!/usr/bin/env python3
# Copyright 2025 Canonical Ltd.
# See LICENSE file for licensing details.

import logging
import ops

from typing import Optional
from charms.data_platform_libs.v0.data_interfaces import DatabaseCreatedEvent
from charms.data_platform_libs.v0.data_interfaces import DatabaseRequires
from charms.tls_certificates_interface.v4.tls_certificates import (
    Certificate,
    CertificateRequestAttributes,
    Mode,
    PrivateKey,
    TLSCertificatesRequiresV4,
)

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
        framework.observe(self.certificates.on.certificate_available, self._configure)
        framework.observe(self.on.collect_unit_status, self._on_collect_status)

        # Database relation
        self.database = DatabaseRequires(self, relation_name="database", database_name="names_db")
        framework.observe(self.database.on.database_created, self._on_database_created)
        framework.observe(self.database.on.endpoints_changed, self._on_database_created)

        # Main service
        self.pebble_service_name = "microcloud-cluster-manager"
        framework.observe(self.on.microcloud_cluster_manager_pebble_ready, self._on_microcloud_cluster_manager_pebble_ready)

    def _on_database_created(self, event: DatabaseCreatedEvent) -> None:
        """Event is fired when postgres database is created."""
        if self._certificate_is_stored():
            logger.info("Database created and certificate is available, replanning")
            self.container.replan()
        else:
            logger.info("Database created but certificate is not available, setting unit to waiting status")
            self.unit.status = ops.WaitingStatus("Waiting for TLS certificate to be available")

    def fetch_postgres_relation_data(self) -> dict[str, str]:
        """Fetch postgres relation data.

        This function retrieves relation data from a postgres database using
        the `fetch_relation_data` method of the `database` object. The retrieved data is
        then logged for debugging purposes, and any non-empty data is processed to extract
        endpoint information, username, and password. This processed data is then returned as
        a dictionary. If no data is retrieved, the unit is set to waiting status and
        the program exits with a zero status code."""
        relations = self.database.fetch_relation_data()
        logger.debug('Got following database data: %s', relations)
        for data in relations.values():
            if not data:
                continue
            logger.info('New PSQL database endpoint is %s', data['endpoints'])
            host, port = data['endpoints'].split(':')
            db_data = {
                'db_host': host,
                'db_port': port,
                'db_username': data['username'],
                'db_password': data['password'],
            }
            return db_data
        return {}

    def _get_certificate_request_attributes(self) -> CertificateRequestAttributes:
        return CertificateRequestAttributes(common_name="example.com")

    def _on_collect_status(self, event: ops.CollectStatusEvent):
        if not self._relation_created("certificates"):
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

    def _configure(self, _: ops.EventBase):
        if not self._relation_created("certificates"):
            return
        if not self._certificate_is_available():
            return
        certificate_update_required = self._check_and_update_certificate()
        if certificate_update_required:
            self.container.replan()

    def _relation_created(self, relation_name: str) -> bool:
        return bool(self.model.relations.get(relation_name))

    def _certificate_is_available(self) -> bool:
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

    def _on_microcloud_cluster_manager_pebble_ready(self, event: ops.PebbleReadyEvent) -> None:
        self.container.add_layer('microcloud_cluster_manager', self._pebble_layer, combine=True)
        self.container.replan()
        self.unit.status = ops.ActiveStatus()

    @property
    def _pebble_layer(self) -> ops.pebble.Layer:
        management_api_environment = self.app_environment
        management_api_environment['SERVICE'] = 'management-api'

        cluster_connector_environment = self.app_environment
        cluster_connector_environment['SERVICE'] = 'cluster-connector'

        pebble_layer: ops.pebble.LayerDict = {
            'summary': 'management api service',
            'description': 'pebble config layer for management api service of microcloud cluster manager',
            'services': {
                "mcm-management-api": {
                    'override': 'replace',
                    'summary': 'microcloud cluster manager management-api',
                    'command': 'microcloud-cluster-manager',
                    'startup': 'enabled',
                    'environment': management_api_environment,
                },
                "mcm-cluster-connector": {
                    'override': 'replace',
                    'summary': 'microcloud cluster manager cluster-connector',
                    'command': 'microcloud-cluster-manager',
                    'startup': 'enabled',
                    'environment': cluster_connector_environment,
                }
            },
        }
        return ops.pebble.Layer(pebble_layer)

    @property
    def app_environment(self) -> dict[str, str]:
        db_data = self.fetch_postgres_relation_data()

        env = {
            key: value
            for key, value in {
                "CLUSTER_CONNECTOR_ADDRESS": "cc.lxd-cm.local:30000",
                "CLUSTER_CONNECTOR_TLS_PATH": CERTS_DIR_PATH,
                "MANAGEMENT_API_TLS_PATH": CERTS_DIR_PATH,
                "DB_DISABLE_TLS": "true",
                "DB_HOST": db_data.get("db_host", None),
                "DB_PORT": db_data.get("db_port", None),
                "DB_USER": db_data.get("db_username", None),
                "DB_PASSWORD": db_data.get("db_password", None),
                "DB_MAX_IDLE": "2",
                "DB_MAX_OPEN": "5",
                "DB_NAME": "cm",
                "OIDC_AUDIENCE": "https://lxd-ui-demo.us.auth0.com/api/v2/",
                "OIDC_CLIENT_ID": "OZSAeCbqAXZid3LL1gRQEkLXP9KlwZtJ",
                "OIDC_ISSUER": "https://lxd-ui-demo.us.auth0.com/",
                # "PROMETHEUS_BASE_URL": "http://192.168.1.133",
                "SERVER_HOST": "0.0.0.0",
                "SERVER_PORT": "9000",
                "TEST_MODE": "false",
                "VERSION": "development",
                "SERVICE": "management-api",
            }.items()
            if value is not None
        }
        return env


if __name__ == "__main__":  # pragma: nocover
    ops.main(ClusterManagerCharm)
