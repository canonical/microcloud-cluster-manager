# Prerequisites

LXD Cluster Manager has the following dependencies:

## Golang

You will need to [install Go](https://go.dev/dl/) on your host for developing the backend. Consult [go.mod](go.mod) for the minimum require go version.

## docker

Docker is required to run the application on your local machine. Information on installing Docker can be found [here](https://docs.docker.com/get-started/get-docker/).

## kubctl

You will need to install `kubectl` for deploying k8s resources to the local cluster. You can follow the instructions from the official [docs](https://kubernetes.io/docs/tasks/tools/).

## kind

Kind is required for setting up a local k8s cluster inside a docker container. You can follow instructions from this [guide](https://kind.sigs.k8s.io/docs/user/quick-start/) to install kind.

## skaffold

We use skaffold for setting up the local k8s development environment with hot reloading and debugging functionalities. You can install it with the following commands.
```
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && \
sudo install skaffold /usr/local/bin/
```

# Running the local development cluster

Once you have all the prerequisite dependencies installed. You can start the development clutser by running the following commands in project root:

### Start the development cluster
```
make dev
```

### Stop and cleanup the development cluster
```
make nuke
```