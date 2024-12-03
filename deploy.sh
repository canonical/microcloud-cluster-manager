#!/bin/bash

# pre-requisites
# 1. clone project
# lxc file push -rp lxd-cluster-manager lxd-cm/home/lxd-cluster-manager/
# 2. Setup k8s cluster
# snap install microk8s --classic --channel=1.31/stable
# microk8s status --wait-ready
# alias k="microk8s kubectl"

set -e
