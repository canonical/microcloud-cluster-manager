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

## Git hooks

To enable pre-commit checks, run:

```
make add-hooks
```

After this, each commit runs backend lint checks with `golangci-lint` and UI checks with `yarn lint-staged`.

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

### Enrolling test clusters

The `scripts/lxd-container-setup.sh` script creates single-member MicroCloud clusters in
LXD containers and enrolls each of them with the Cluster Manager running on your
machine. It is useful to test the Cluster Manager against real MicroCloud clusters
without having to set up multiple machines.

With `make dev` running, build the Cluster Manager binary and run the script:

```bash
make build
./scripts/lxd-container-setup.sh --count 2
```

The script creates a dedicated LXD project named `microcloud-mcm`. It shares the default
project's images, networks, and profiles so that the containers use the host's existing
network and storage configuration. All host-side container commands explicitly select
this project; the active LXD project is not changed.
Remove any same-named containers in other projects before running the script, because
container DNS names must be unique on the shared network.

For every container the script installs the LXD and MicroCloud snaps, initializes a
single-member MicroCloud cluster, creates a join token with the Cluster Manager CLI,
joins the cluster to the Cluster Manager and enables the reverse tunnel. It also points
the LXD of every cluster at the same identity provider as the Cluster Manager, so that
the OIDC tokens the Cluster Manager verifies are issued by a provider that the cluster's
LXD trusts too. The `oidc.issuer`, `oidc.client.id`, `oidc.client.secret` and
`oidc.audience` values come from the environment, and the groups claim is set to
`mcm-idp-groups`, which is the claim the Cluster Manager reads. The client secret is
needed because the identity provider is registered as a confidential client, and LXD uses
the same client when it handles its own login callback.

The script binds each LXD HTTPS API to its cluster address instead of all interfaces.
This prevents LXD from advertising nested bridge addresses, such as the address of
`lxdfan0`, which are not covered by the cluster certificate.
When reusing an enrolled cluster, the script restarts the MicroCloud daemon and waits
for it to become ready. This briefly interrupts the reverse tunnel, allowing it to
reconnect with the current LXD address and certificate instead of retaining old values.
The cluster address is also added to the SANs of its LXD cluster certificate. The
requests that arrive over the reverse tunnel are proxied by `microcloudd` to the LXD of the
cluster, and `microcloudd` verifies that connection against the cluster certificate. LXD
only puts the hostname and the loopback addresses in the SANs of the certificate it
generates, so without this the verification fails with `certificate is valid for 127.0.0.1,
::1, not <address>` and every tunnel request is rejected.

The identity provider group `admins` is then mapped to the LXD group `admins`, which LXD
creates with the server admin entitlement. The Cluster Manager only accepts users in the
OIDC group `admins`, and the mapping is what lets such a user authenticate against a
cluster's LXD with permission to do anything, because access is granted by the claim in
the token rather than by a separate LXD login.

The script loads the development environment from `scripts/setup-env.sh`, because the
Cluster Manager CLI reads the same configuration as the services. The file is sourced
with its side effects stubbed out, so it only exports the variables and does not start
the Postgres and Prometheus containers again.

Before it creates the containers, the script compares the certificate that the running
Cluster Connector serves with the certificate of this checkout, and stops when they
differ. The join token carries the certificate fingerprint of the Cluster Connector, and
MicroCloud refuses to join when the connector answers with a different certificate. This
only happens when the connector was started from another checkout, which is easy to do
with git worktrees; in that case, restart the development environment with `make dev`
from this checkout.

Use `--microcloud-channel` to install the MicroCloud snap from another channel:

```bash
./scripts/lxd-container-setup.sh --count 2 --microcloud-channel 3/edge
```

The containers are named `microcloud-01`, `microcloud-02`, ... and the MicroCloud
cluster inside a container is named after the container's hostname. The script can be
run again: it reuses existing containers in `microcloud-mcm`, skips the snaps that are
already installed, preserves initialized clusters and their enrollment, and rewrites the
Cluster Manager host entry only when it changed. Existing containers in other projects
are not moved or reused.

To list the containers:

```bash
lxc --project microcloud-mcm list
```

The logs of a run are written to a temporary directory. The directory is removed when
the run succeeds and kept when it fails, and the script then prints where the logs are.
The script waits for the MicroCloud daemon before running preseed. If preseed fails,
its log includes the exit status (or an explicit timeout message) and recent MicroCloud
daemon logs, even when preseed itself produces no output.

To force-delete the default number of containers, including their data:

```bash
./scripts/lxd-container-setup.sh --nuke
```

Use `-n` or `--count` to select how many containers to delete, starting with
`microcloud-01`. For example, to delete the first three containers:

```bash
./scripts/lxd-container-setup.sh --nuke --count 3
```

Cleanup only targets the selected containers in `microcloud-mcm`. It keeps the project
and does not require the Cluster Manager development environment or binary.
Cluster Manager records are not removed; remove the corresponding entries in the UI
before enrolling replacement clusters with the same names.

Run `./scripts/lxd-container-setup.sh --help` for all options.

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
