#!/bin/bash

set -e
set -x

# setup /etc/hosts entries for local domains
make add-hosts

# generate keys if they don't exist
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
KEY_DIR="$SCRIPT_DIR/keys"
if [ ! -d "$KEY_DIR" ]; then
  mkdir -p $KEY_DIR
  openssl req -nodes -x509 -newkey ec \
    -pkeyopt ec_paramgen_curve:secp384r1 \
    -sha384 \
    -keyout $KEY_DIR/tls.key \
    -out $KEY_DIR/tls.crt \
    -subj "/C=GB/ST=London/L=London/O=Microcloud Cluster Manager/OU=dev/CN=ma.lxd-cm.local" \
    -days 3000 \
    -addext "subjectAltName=DNS:ma.lxd-cm.local,DNS:cc.lxd-cm.local"

  cp $KEY_DIR/tls.crt $KEY_DIR/ca.crt
fi
export management_api_cert_secret="$KEY_DIR"
export cluster_connector_cert_secret="$KEY_DIR"

# run postgres
docker rm -f my-postgres 2>/dev/null || true
docker run -d \
  --name my-postgres \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=cm \
  -p 5432:5432 \
  postgres:latest

# run prometheus
docker rm -f prometheus 2>/dev/null || true
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  prom/prometheus \
  --web.enable-remote-write-receiver \
  --config.file=/etc/prometheus/prometheus.yml
PROMETHEUS_ADDRESS="$(hostname -I | awk '{print $1}'):9090"

# build golang app
make build

# set environment variables
export DB_HOST=localhost
export DB_PASSWORD=admin
export OIDC_AUDIENCE=https://dev-h6c02msuggpi6ijh.eu.auth0.com/api/v2/
export OIDC_CLIENT_ID=RYDnMpkygLAMfeo17lU7LYwWGxisRuRR
export OIDC_CLIENT_SECRET=CNKX4UmrZKZJq5rJy5VM_JfcNPqkws1rwWWQk_q0oyZ8gABARr19ic7xrhPssGA1
export OIDC_ISSUER=https://dev-h6c02msuggpi6ijh.eu.auth0.com/
export MANAGEMENT_API_TLS_PATH="$KEY_DIR"
export CLUSTER_CONNECTOR_TLS_PATH="$KEY_DIR"
export CLUSTER_CONNECTOR_PORT=9000
export PROMETHEUS_BASE_URL=http://$PROMETHEUS_ADDRESS/api/v1/write

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

  # stop postgres
  echo "stop postgres..."
  CONTAINER_ID=$(docker ps -q -f name=my-postgres)
  docker stop "$CONTAINER_ID"
  docker rm "$CONTAINER_ID"
}

trap cleanup INT


# run the management api
export SERVICE=management-api
export SERVER_PORT=30000
export STATUS_PORT=30003
cmd/app
