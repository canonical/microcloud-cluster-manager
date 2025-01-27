#!/bin/bash

set -e

image="$1"

configure_microk8s() {
  sudo microk8s config > ~/.kube/config
  sudo microk8s enable hostpath-storage dns
  IPADDR=$(ip -4 -j route get 2.2.2.2 | jq -r '.[] | .prefsrc')
  sudo microk8s enable metallb:$IPADDR-$IPADDR
}

setup_juju_controller() {
  make install-juju
  mkdir -p ~/.local/share
  sudo sh -c "mkdir -p /var/snap/juju/current/microk8s/credentials"
  sudo sh -c "microk8s config | tee /var/snap/juju/current/microk8s/credentials/client.config"
  sudo chown -f -R $USER:$USER /var/snap/juju/current/microk8s/credentials/client.config
  sudo juju bootstrap microk8s cm-controller
}

deploy_cos_lite() {
  sudo juju add-model cos && sudo juju switch cos
  sudo juju deploy cos-lite --trust
}

deploy_k8s_resources() {
  make deploy-ci-k8s-cluster IMAGE_NAME=$image
}

main() {
  configure_microk8s
  setup_juju_controller
  deploy_k8s_resources
  deploy_cos_lite
}

main
