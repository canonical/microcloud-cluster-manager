# MicroCloud Cluster Manager

Cluster Manager is a tool for viewing and managing multiple MicroCloud deployments. It includes the [Canonical Observability Stack](https://charmhub.io/topics/canonical-observability-stack) for monitoring and alerting with Grafana and Prometheus, along with a web UI for viewing information about the registered MicroClouds.

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

# Charm

Useful commands

```bash
# connect to the containers
kubectl exec -it microcloud-cluster-manager-k8s-0 -c microcloud-cluster-manager -n cluster-manager-juju-dev -- /bin/bash
kubectl exec -it microcloud-cluster-manager-k8s-0 -c charm -n cluster-manager-juju-dev -- /bin/bash

# manual port forwarding
kubectl port-forward pod/microcloud-cluster-manager-k8s-0 8414:9100 -n cluster-manager-juju-dev
kubectl port-forward pod/microcloud-cluster-manager-k8s-0 31000:9000 -n cluster-manager-juju-dev
sudo socat TCP-LISTEN:32000,fork TCP:localhost:31000

# run migrations
eval $(cat /proc/136/environ | tr '\0' '\n' | grep -v '^SERVICE=' | sed 's/^/export /') && export SERVICE=admin && microcloud-cluster-manager
```