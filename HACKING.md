# Prerequisites

LXD Cluster Manager has the following dependencies:

## Backend

### Golang

You will need to [install Go](https://go.dev/dl/) on your host for developing the backend.

### DQLite

You will need to install the development version of dqlite as the main database engine for the backend cluster system.
```
sudo add-apt-repository ppa:dqlite/dev -y --no-update
sudo apt-get update

sudo apt-get install --no-install-recommends -y libdqlite-dev
```

### lxd-generate

[lxd-generate](https://pkg.go.dev/github.com/canonical/lxd/lxd/db/generate#section-readme) is a code generation tool used by the lxd team for generating database helper functions for DQLite. If you are working with database related tasks and need to run `make update-schema`, then you will need to setup `lxd-generate` with the following steps.

1. clone the [lxd](https://github.com/canonical/lxd) repo
2. navigate to `./lxd/db/generate` inside the proejct
3. run `go build -o lxd-generate`. This should create the binary for `lxd-generate` inside the directory you are in.
4. Then move that binary to your system's path for go binaries, usually `/usr/go/bin`

## Frontend

### Nodejs

You should [install nodejs](https://nodejs.org/en/download/package-manager/current) for frontend development.

### Yarn

You will need to install [yar](https://yarnpkg.com/) for package management. 

### Dotrun

You will need to install [dotrun](https://github.com/canonical/dotrun#installation) for starting the local dev server inside docker container.

### Playwright

You will need to install playwright for running UI e2e tests. If you have installed Nodejs in the previous step, you should be able to simply run `npx playwright install` to install Playwright and its browser dependencies.

# Environment variables

Before running the development servers, you will need to create a file at the path `ui/.env.local`. It should have the following variables, please reach out to the project maintainers for the `OIDC_` variable values.

```
CLUSTER_MANAGER_BACKEND_IP=
OIDC_USER=
OIDC_PASSWORD=
OIDC_ISSUER=
OIDC_CLIENT_ID=
OIDC_AUDIENCE=
NUM_MEMBERS=1
GLOBAL_ADDRESS=""
POPULATE_MEMBER_EXTERNAL_ADDRESSES=false
```
The `CLUSTER_MANAGER_BACKEND_IP` in the file should match your docker bridge (docker should have been installed as part of the `dotrun` setup process). You can find the correct address with the command `ip address show`. On macOS you might be able to skip this step as the IP matches the contents of the `.env` file that is checked into the repo. 

# Running development backend

Run the backend cluster daemon with yarn on your host:
    
    cd ui
    yarn backend-run

The above command builds the backend daemon and starts it in a go process. If successful, you should see the following terminal output:

    WARNING[2024-08-02T15:49:56+02:00] microcluster database is uninitialized   

In a separate termiinal, bootstrap the cluster database with the following command:

    cd ui
    yarn backend-init

Wait for the init command to finish and once done, you should see the following terminal output:

    => Query 0:

    Rows affected: 200

    => Query 1:

    Rows affected: 200

    => Query 2:

    Rows affected: 0

    Loading environment variables from ./ui/.env.local
    The token can not be retrieved at a later stage, please save it now.
    eyJzZWNyZXQiOiIwNjU4NGViNzUzOWU2ODk5ODQwYjJhZTMyNzg4ODNmMzAwMDJmNThlMjEzNmQ2ZjBkZGVkMmVmNzBmZWUxYzU0IiwiZXhwaXJlc19hdCI6IjIwMjQtMDgtMDNUMTU6NTc6NTUuNDExMTI3NTUzKzAyOjAwIiwiYWRkcmVzc2VzIjpbIjAuMC4wLjA6OTAxMCJdLCJzZXJ2ZXJfbmFtZSI6ImNsdXN0ZXIxIiwiZmluZ2VycHJpbnQiOiI1M2FiYzA0MzE2NzVkNjAyMWUwZmUzNzExMDlkNWQ4YzhlNzVlZGNhNTAxYTQ3ZGU3MzU4ZmYwMDkxM2U5NDUwIn0=
    Done in 4.52s.

# Running development frontend

To start the UI development server, run the following command:

```
cd ui
dotrun
```

You should be able to browse the ui via https://0.0.0.0:8414/

# Join additional cluster members

To add more members to the LXD Clutser Manager, you can run the following commands:
```
go run ./cmd/lxd-cluster-mgrd --state-dir ./state/dir2 &

token_node2=$(go run ./cmd/lxd-cluster-mgr --state-dir ./state/dir1 tokens add "member2")

go run ./cmd/lxd-cluster-mgr --state-dir ./state/dir2 init "member2" 0.0.0.0:9002 --token ${token_node2} --control-address 0.0.0.0:9020
```

# End-to-end tests

### Backend e2e

To run backend e2e test suite. You can run `make test-e2e` from the root of the poject.

### Frontend e2e

To run UI e2e tests, you will need to have the UI development server running first. Execute `dotrun` in the `ui` directory first then run the tests with:

```
cd ui
yarn test-e2e
```

Note we use playwright to run the UI e2e test suites, you should have playwright installed as mentioned previously before running the tests.