# Contributing

## Required dependencies

### Automatic dependency installation

For convenience, two `make` targets are prepared to install all required dependencies for the development environment. You may simply run the following command at the project root:

```
make install-core
```

then run

```
make install-deps
```

**CAUTION**: The `install-core` and `install-deps` targets have been tested only in an Ubuntu Linux environment and may not work on other operating systems. It is strongly recommended that you avoid running this directly on your host machine. Instead, use it as a convenient method for setting up a VM-based development environment.

### Manual dependency installation

If you choose to set up the dependencies manually, in order to run the local development environment, the following dependenicies of the LXD Cluster Manager project must first be installed and configured:

#### Make

The project is managed using the `makefile` targets located at the project root. You will need to install `make`.

```
sudo apt update
sudo apt install make
```

#### Golang

You will need to [install Go](https://go.dev/dl/) on your host for developing the backend. Consult [go.mod](go.mod) for the required go version. Make sure you add the Go binary to `PATH` after installation.

#### Docker

Docker is required to build and run the service containers for the Kubernetes cluster. It is highly recommended that you install the docker snap package as it does not have network inteference with `lxd` (another dependency required for building rocks).

```
sudo snap install docker
```

By default, Docker is only accessible with root privileges (sudo). We want to be able to use Docker commands as a regular user:

```
sudo addgroup --system docker
sudo adduser $USER docker
newgrp docker
sudo snap disable docker
sudo snap enable docker
```

#### Kubectl

You will need to install `kubectl` for deploying resources to the local Kubernetes cluster. You can follow the installation instructions from the official [docs](https://kubernetes.io/docs/tasks/tools/).

#### Kind

Kind is required for setting up the local Kubernetes cluster inside a docker container. You can follow instructions from this [guide](https://kind.sigs.k8s.io/docs/user/quick-start/) to install kind.

#### Skaffold

We use skaffold for managing resources in the local Kubernetes cluster with hot reloading and debugging functionalities. You can install it with the following commands.

```
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && \
sudo install skaffold /usr/local/bin/
```

#### Node version manager (NVM)

Node.js is required to run the development UI. Instead of installing Nodejs directly, we recommend that you install NVM for simpler management of multiple Node.js versions. You can follow instructions from the [official GitHub repository](https://github.com/nvm-sh/nvm) to install NVM.

#### Dotrun

Dotrun is required to spin up a development container for the UI. You can install it by following the instructions [here](https://github.com/canonical/dotrun#installation).

#### Juju

Juju is required to deploy the Canonical Observability [Stack](https://charmhub.io/topics/canonical-observability-stack) to our development k8s cluster.
To install Juju, run the following command:

```
make install-juju
```

## Optional dependencies

Generally, the required dependencies cover all the essentials for setting up the local development environment. Below are additional optional dependencies for more advanced tasks:

### LXD

LXD is required to build a rock using Rockcraft. It must be installed and initialized:

```
sudo snap install lxd
lxd init --auto
```

### Rockcraft

For local development, the Kubernetes cluster image is built using Docker. However, for production and CI, the image is built as a rock using Rockcraft. To install Rockcraft, run the following command:

```
sudo snap install rockcraft --classic
```

## Running the local development environment

### Run the backend cluster

Once you have all the required dependencies installed, to get the local development environment up and running, run the following commands at the project root:

```
make dev
```

Once the cluster is ready, you should see terminal output logs similar to the following example:

```
Deployments stabilized in 31.094 seconds
Port forwarding service/management-api-svc in namespace default, remote port management-api -> http://127.0.0.1:9000
Port forwarding service/cluster-connector-svc in namespace default, remote port cluster-conn -> http://127.0.0.1:9001
Port forwarding service/db-svc in namespace default, remote port db -> http://127.0.0.1:5432
Listing files to watch...
 - microcloud-cluster-manager
Press Ctrl+C to exit
...
Watching for changes...
```

**NOTE**: If it's your first time starting the development environment, it may take a while for all the resources and images to be pulled into the cluster.

### Run the UI in a separate terminal

First, add the local development hosts `ma.lxd-cm.local` and `cc.lxd-cm.local` to your `/etc/hosts` file. You can do this with the following command:

```
sudo make add-hosts
```

Then start the UI development server:

```
make ui
```

Once the UI is ready, you should see terminal output logs similar to the following example:

```
cd ui && dotrun
Checking for dotrun image updates...
- Yarn dependencies have changed, reinstalling

[ $ yarn install ]
...
[ $ yarn run start ]

yarn run v1.22.21
$ ./entrypoint 0.0.0.0:${PORT}
[WARNING] 344/114722 (65) : parsing [haproxy-local.cfg:11] : 'bind ma.lxd-cm.local:8414' :
  unable to load default 1024 bits DH parameter for certificate 'keys/lxd-cm.pem'.
  , SSL library will use an automatically generated DH parameter.
[WARNING] 344/114722 (65) : Setting tune.ssl.default-dh-param to 1024 by default, if your workload permits it you should set it to at least 2048. Please set a value >= 1024 to make this warning disappear.
Listening at https://ma.lxd-cm.local:8414
Re-optimizing dependencies because lockfile has changed

  VITE v5.4.8  ready in 164 ms

  ➜  press h + enter to show help
```

**NOTE**: You can reach the UI in your browser at `https://ma.lxd-cm.local:8414`.

### Stop and cleanup the development cluster

Unfortunately, skaffold creates new images for each rebuild due to detected code changes. Those images are not automatically deleted from your local docker image registry. This can cause unbounded disk utilisation on your host machine. To prevent this from happening, you should always run the following command after you are done with development for the day to clean up unused images:

```
make nuke
```

## Run the backend cluster with rock

If you need to work on building the image with Rockcraft, you can test out the rock with the backend cluster by running the following command:

```
make dev-rock
```

**NOTE**: You must have installed the optional dependencies for the above to work.

## End-to-end (e2e) tests

The e2e tests steps through the OIDC authentication flow using a pre-configured auth0 account with admin access. You can run both backend and frontend e2e tests against the running development cluster with the following command:

### Backend e2e tests

```
make test-e2e
```

### UI e2e tests

For the tests to work locally, you will need to create a file at the path `ui/.env.local` containing the following variables:

```
OIDC_USER="cluster-manager-e2e-tests@example.org"
OIDC_PASSWORD="cluster-manager-e2e-password"
```

If it's the first time you are running UI e2e tests, you should first install playwright and its browsers:

```
cd ui && npx playwright install
```

You can then run the UI e2e tests with:

```
make test-ui-e2e
```
