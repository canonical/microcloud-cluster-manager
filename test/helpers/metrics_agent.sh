#!/bin/bash

# To setup metrics for lxd, make sure the following is done:
# 1. Install LXD on the host machine
# 2. Initialise LXD and have core.https_address set to ":8443"
# 3. Disable authentication for metrics "lxc config set core.metrics_authentication false"
# 4. Make sure at least one instance is created on the lxd server to get some metrics

set -ex

LXD_SERVER_URL="https://localhost:8443"
CLUSTER_CONNECTOR_URL="https://cc.lxd-cm.local:30000/1.0"
MANAGEMENT_API_URL="https://ma.lxd-cm.local:30000/1.0"
INTERVAL=15
CERT_DIR="/tmp/cm"
CLUSTER_CONNECTOR_SERVICE="cluster-connector"

ensure_service_running() {
  # Restart the deployment
  echo "Restarting deployment map-api"
  kubectl rollout restart deployment management-api-depl
  kubectl rollout status deployment management-api-depl --timeout 120s; \

  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to restart deployment."
    return 1
  fi

  echo "Deployment restarted successfully."
}

create_remote_cluster_join_token() {
  local cluster_name="$1"
  local expiry="$2"

  local payload="{\"cluster_name\":\"$cluster_name\""
  if [[ -n "$expiry" ]]; then
    payload+=",\"expiry\":\"$expiry\""
  fi
  payload+="}"

  local endpoint="$MANAGEMENT_API_URL/remote-cluster-join-token"

  local response=$(curl -sk \
    -H "Content-Type: application/json" \
    -X POST \
    -d "$payload" \
    "$endpoint")

  # Extract token from response
  local token=$(echo $response | jq -r '.metadata.token')
  if [[ "$token" == "null" || -z "$token" ]]; then
    echo "Error: Failed to retrieve token from response."
    exit 1
  fi

  echo "$token"
}

decode_token() {
  local token="$1"

  # Decode the base64 token
  local decoded_token=$(echo "$token" | base64 --decode)
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to decode base64 token."
    exit 1
  fi

  # Parse the JSON structure
  echo $decoded_token
}

send_join_request() {
  local decoded_token="$1"
  local server_name=$(echo $decoded_token | jq -r '.server_name')
  local secret=$(echo $decoded_token | jq -r '.secret')

  # Generate a dedicated lxd cert for cluster manager
  local client_cert="$CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/$server_name.pem"
  local client_key="$CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/$server_name-key.pem"
  openssl req -x509 -nodes -newkey rsa:2048 -keyout $client_key -out $client_cert -subj "/CN=$server_name" -days 365
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to generate LXD certificate."
    exit 1
  fi

  local remote_cluster_certificate=$(awk '/BEGIN CERTIFICATE/,/END CERTIFICATE/' $client_cert | sed ':a;N;$!ba;s/\n/\\n/g')
  local payload=$(printf '{"cluster_name":"%s","cluster_certificate":"%s"}' "$server_name" "$remote_cluster_certificate")

  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to create JSON payload."
    exit 1
  fi

  local hmac=$(echo -n $payload | openssl dgst -sha256 -hmac $secret -binary | base64)
  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to compute HMAC signature."
    exit 1
  fi

  # Send the request
  local endpoint="$CLUSTER_CONNECTOR_URL/remote-cluster"
  local response=$(curl -sk \
    -H "Content-Type: application/json" \
    -H "X-CLUSTER-SIGNATURE: $hmac" \
    -X POST \
    -d "$payload" \
    "$endpoint")

  if [[ $? -ne 0 ]]; then
    echo "Error: Failed to send join request."
    exit 1
  fi

  echo "Join request successful."
}

approve_join_request() {
  local decoded_token="$1"
  local cluster_name=$(echo $decoded_token | jq -r '.server_name')

  local payload='{"status":"ACTIVE"}'

  curl -sk \
    -H "Content-Type: application/json" \
    -X PATCH \
    -d "$payload" \
    "$MANAGEMENT_API_URL/remote-cluster/$cluster_name"
}

get_certificate() {
    local secret_name="$1"
    local service_name="$2"

    mkdir -p $CERT_DIR/$service_name
    kubectl get secret $secret_name -o jsonpath='{.data.tls\.crt}' | base64 -d > $CERT_DIR/$service_name/tls.crt
    kubectl get secret $secret_name -o jsonpath='{.data.tls\.key}' | base64 -d > $CERT_DIR/$service_name/tls.key
    kubectl get secret $secret_name -o jsonpath='{.data.ca\.crt}' | base64 -d > $CERT_DIR/$service_name/ca.crt
}

send_status_updates() {
    local decoded_token="$1"
    local server_name=$(echo $decoded_token | jq -r '.server_name')
    local client_cert="$CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/$server_name.pem"

    while true; do
        echo "Fetching metrics from LXD server..."
        metrics=$(curl -sk "$LXD_SERVER_URL/1.0/metrics")

        payload=$(cat <<EOF
{
    "cpu_total_count": 0,
    "cpu_load_1": "0.0",
    "cpu_load_5": "0.0",
    "cpu_load_15": "0.0",
    "memory_total_amount": 0,
    "memory_usage": 0,
    "disk_total_size": 0,
    "disk_usage": 0,
    "member_statuses": [],
    "instance_statuses": [],
    "metrics": $(jq -Rs . <<< "$metrics")
}
EOF
        )

        echo "Sending metrics to Go server..."
        curl -s --cert $CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/tls.crt --cacert $CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/ca.crt \
            -X POST "$CLUSTER_CONNECTOR_URL" \
            -H "Content-Type: application/json" \
            -d "$payload"

        sleep "$INTERVAL"
    done
}

# main ========================================================================
ensure_service_running
get_certificate "cluster-connector-cert-secret" "$CLUSTER_CONNECTOR_SERVICE"
token=$(create_remote_cluster_join_token "lxd-cluster-$(uuidgen)")
decoded_token=$(decode_token $token)
send_join_request $decoded_token
approve_join_request $decoded_token

# send status updates every 15s
server_name=$(echo $decoded_token | jq -r '.server_name')
client_cert="$CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/$server_name.pem"
client_key="$CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/$server_name-key.pem"

while true; do
    echo "Fetching metrics from LXD server..."
    metrics=$(curl -sk "$LXD_SERVER_URL/1.0/metrics")

    payload=$(cat <<EOF
{
    "cpu_total_count": 0,
    "cpu_load_1": "0.0",
    "cpu_load_5": "0.0",
    "cpu_load_15": "0.0",
    "memory_total_amount": 0,
    "memory_usage": 0,
    "disk_total_size": 0,
    "disk_usage": 0,
    "member_statuses": [],
    "instance_statuses": [],
    "metrics": $(jq -Rs . <<< "$metrics")
}
EOF
    )

    echo "Sending metrics to Go server..."
    curl -sv --cert $client_cert --key $client_key --cacert $CERT_DIR/$CLUSTER_CONNECTOR_SERVICE/ca.crt \
        -X POST "$CLUSTER_CONNECTOR_URL/remote-cluster/status" \
        -H "Content-Type: application/json" \
        -d "$payload"

    sleep "$INTERVAL"
done
