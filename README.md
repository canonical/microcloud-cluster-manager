# MicroCloud Cluster Manager

Cluster Manager is the entry point for all your MicroClouds. It can be extended with the [Canonical Observability Stack](https://charmhub.io/topics/canonical-observability-stack) for monitoring and alerting.

# Get started

This is an early version of cluster manager. Glad you want to try it as an early adopter!

You need to have a juju k8s environment ready. Deploy MicroCloud Cluster Manager along with its dependencies:

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

For authentication you need an OIDC provider, you can use [Auth0](https://documentation.ubuntu.com/lxd/latest/howto/oidc_auth0/), [ory hydra](https://documentation.ubuntu.com/lxd/latest/howto/oidc_ory/), [Keycloak](https://documentation.ubuntu.com/lxd/latest/howto/oidc_keycloak/), [Microsoft Entra](https://documentation.ubuntu.com/lxd/latest/howto/oidc_entra_id/) or others. Configure the cluster manager charm with your auth provider:

```
juju config microcloud-cluster-manager-k8s oidc-issuer=https://example.com/
juju config microcloud-cluster-manager-k8s oidc-client-id=ababab
juju config microcloud-cluster-manager-k8s oidc-audience=https://example.com/api/v2/
```

You might want to set a domain for your traefic controller
```
juju config traefik-k8s external_hostname=example.com
```

Now you can access the web ui via https://example.com

<img width="1434" height="701" alt="image" src="https://github.com/user-attachments/assets/18ddfef1-db66-4971-bbcf-5af7067f3581" />

After login through your OIDC provider, you can enroll your first cluster

<img width="1435" height="745" alt="image" src="https://github.com/user-attachments/assets/987942d6-d53f-470e-b1a9-081b171a23f7" />

You can now explore your first cluster

<img width="1435" height="745" alt="image" src="https://github.com/user-attachments/assets/e1998f49-2c6d-42f1-9042-9a37cdab05ce" />

You can extend with the observability charm to have grafana and prometheus integrated:

```
juju add-model cos
juju deploy cos-lite --trust
juju offer prometheus:receive-remote-write
juju offer grafana:grafana-dashboard grafana-db
juju offer grafana:grafana-metadata
juju switch cluster-manager-juju-dev
juju integrate microcloud-cluster-manager-k8s:send-remote-write admin/cos.prometheus
juju integrate microcloud-cluster-manager-k8s:grafana-dashboard admin/cos.grafana-db
juju integrate microcloud-cluster-manager-k8s:grafana-metadata admin/cos.grafana
```

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

Now you can access the UI at [ma.lxd-cm.local:8414](https://ma.lxd-cm.local:8414). For more information on local development, please see the [contributing guidelines](CONTRIBUTING.md).

# Architecture

Cluster Manager is a distributed web application with a Go backend and a React Typescript UI. The application runs in Kubernetes. For an overview of the system, see the [architecture documentation](ARCHITECTURE.md).
