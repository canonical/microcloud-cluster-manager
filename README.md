# MicroCloud Cluster Manager

Cluster Manager is the entry point for all your MicroClouds. It can be extended for monitoring and alerting.

# Get started

This is an early version of cluster manager. Get started as an early adopter today!

You need to a [juju controlled k8s environment](https://juju.is/). Deploy MicroCloud Cluster Manager along with its dependencies:

```
juju add-model cluster-manager

juju deploy postgresql-k8s --channel 14/stable --trust
juju deploy self-signed-certificates --trust
juju deploy traefik-k8s --trust
juju deploy microcloud-cluster-manager-k8s --channel edge --trust

juju integrate postgresql-k8s:database microcloud-cluster-manager-k8s
juju integrate self-signed-certificates:certificates microcloud-cluster-manager-k8s
juju integrate traefik-k8s:traefik-route microcloud-cluster-manager-k8s
```

For authentication you need an OIDC provider. You can use Auth0, Ory Hydra, Keycloak, Microsoft Entra among others. See [LXD documentation on OIDC](https://documentation.ubuntu.com/lxd/latest/howto/oidc/) for how to configure the provider side. The callback path is `/oidc/callback`, but in contrast to LXD it must be without the port 8443 as cluster manager listens on the default port 443.

Configure the cluster manager charm with your auth provider:
```
juju config microcloud-cluster-manager-k8s oidc-issuer=https://example.com/
juju config microcloud-cluster-manager-k8s oidc-client-id=ababab
juju config microcloud-cluster-manager-k8s oidc-audience=https://example.com/api/v2/
```

Configure the domain for the management api and the cluster connector. You can also use externally exposed IP addresses, but using domains are recommended. The `management-api-domain` is the address for the management interface and its api. The `cluster-connector-domain` is used by clusters to send heartbeats to the cluster manager, it does not need to be exposed to public internet.

```
juju config microcloud-cluster-manager-k8s management-api-domain=example.com
juju config microcloud-cluster-manager-k8s cluster-connector-domain=cluster-connector.example.com
```

You might want to set the domain for your traefic controller
```
juju config traefik-k8s external_hostname=example.com
```

Now you can access the web ui via https://example.com

<img width="1434" height="701" alt="image" src="https://github.com/user-attachments/assets/18ddfef1-db66-4971-bbcf-5af7067f3581" />

## Enroll your first cluster

<img width="1435" height="745" alt="image" src="https://github.com/user-attachments/assets/987942d6-d53f-470e-b1a9-081b171a23f7" />

## Explore your first clusters details

<img width="1435" height="745" alt="image" src="https://github.com/user-attachments/assets/e1998f49-2c6d-42f1-9042-9a37cdab05ce" />

## Extending with observability

You can extend Cluster Manager with the [Canonical Observability Stack](https://charmhub.io/topics/canonical-observability-stack) to have Grafana and Prometheus integrated:

```
juju add-model cos
juju deploy cos-lite --trust
juju offer prometheus:receive-remote-write
juju offer grafana:grafana-dashboard grafana-db
juju offer grafana:grafana-metadata
juju switch cluster-manager
juju integrate microcloud-cluster-manager-k8s:send-remote-write admin/cos.prometheus
juju integrate microcloud-cluster-manager-k8s:grafana-dashboard admin/cos.grafana-db
juju integrate microcloud-cluster-manager-k8s:grafana-metadata admin/cos.grafana
```

This will make a LXD dashboard available in Grafana. With the relation enabled, Cluster Manager starts forwarding metrics to the observability stack on each cluster heartbeat received.

To access Grafana, you can fetch the admin password with:

```
juju run --model cos grafana/leader get-admin-password
```

In the Cluster Manager web interface, you will see a new button on the cluster details page deep linking into the Grafana dashboard.

<img width="1921" height="959" alt="cluster-manager-grafana-cos" src="https://github.com/user-attachments/assets/d015b440-8725-453f-90b4-72ffda4d0ba1" />

# Development setup

**CAUTION**: The `install-deps` target has been tested only in an Ubuntu Linux environment and may not work on other operating systems. It is strongly recommended that you avoid running this directly on your host machine. Instead, use it as a convenient method for setting up a VM-based development environment.

To start the development environment, run these commands:

```bash
make install-deps
sudo make add-hosts
make dev
```

Then in a separate terminal, run:

```bash
make ui
```

Now you can access the UI at [ma.microcloud-cm-example.local:8414](https://ma.microcloud-cm-example.local:8414). For more information on local development, please see the [contributing guidelines](CONTRIBUTING.md).

# Architecture

Cluster Manager is a distributed web application with a Go backend and a React Typescript UI. The application runs in Kubernetes. For an overview of the system, see the [architecture documentation](ARCHITECTURE.md).
