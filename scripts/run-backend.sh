#!/bin/bash

set -e
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "$SCRIPT_DIR/setup-env.sh"

# build golang app
make build

# run the cluster connector
export SERVICE=cluster-connector
export SERVER_PORT=9000
export STATUS_PORT=9009
nohup cmd/app > >(cat) 2> >(cat >&2) &
CLUSTER_CONNECTOR_PID=$!

cleanup() {
  echo "Caught Ctrl+C — cleaning up..."

  # kill app processes
  echo "stopping services..."
  kill $CLUSTER_CONNECTOR_PID

  stop_postgres
  stop_prometheus
}

trap cleanup INT


# run the management api
export SERVICE=management-api
export SERVER_PORT=30000
export STATUS_PORT=30003
cmd/app
