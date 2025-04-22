# MicroCloud Cluster Manager

Cluster Manager is a tool for viewing and managing multiple MicroCloud deployments. It includes the [Canonical Observability Stack](https://charmhub.io/topics/canonical-observability-stack) for monitoring and alerting with Grafana and Prometheus and a web UI for viewing information of the registered MicroClouds.

# Development Setup

To start the development environment, run the commands:

```bash
make install-deps
sudo make add-hosts
make dev
```
and in a separate terminal

```bash
make ui
```

Now you can access the UI at [ma.lxd-cm.local:8414](https://ma.lxd-cm.local:8414). For more information on the local development, please see [contributing guidelines](CONTRIBUTING.md).

# Architecture

Cluster Manager is a distributed web application with a Go backend and React Typescript UI. The application is running in Kubernetes. For an overview of the system, see the [architecture documentation](ARCHITECTURE.md).
