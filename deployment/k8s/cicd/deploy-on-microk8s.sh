#!/bin/bash

set -e
set -x

image="$1"

configure_microk8s() {
  sudo microk8s config > ~/.kube/config
  sudo microk8s enable hostpath-storage dns
  IPADDR=$(ip -4 -j route get 2.2.2.2 | jq -r '.[] | .prefsrc')
  sudo microk8s enable metallb:$IPADDR-$IPADDR
}

deploy_prometheus() {
  sudo docker run -d \
  --name prometheus \
  -p 9090:9090 \
  prom/prometheus \
  --web.enable-remote-write-receiver \
  --config.file=/etc/prometheus/prometheus.yml
}

deploy_k8s_resources() {
  PROMETHEUS_ADDRESS="$(hostname -I | awk '{print $1}'):9090"
  echo "Prometheus address: $PROMETHEUS_ADDRESS"
  make deploy-ci-k8s-cluster IMAGE_NAME=$image PROMETHEUS_ADDRESS=$PROMETHEUS_ADDRESS
}

main() {
  configure_microk8s
  deploy_prometheus
  deploy_k8s_resources
}

main
