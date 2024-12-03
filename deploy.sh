#!/bin/bash

# pre-requisites
# 1. clone project
# lxc file push -rp lxd-cluster-manager lxd-cm/~/lxd-cluster-manager/
# 2. Setup k8s cluster
# snap install microk8s --classic --channel=1.31/stable
# microk8s status --wait-ready
# alias k="microk8s kubectl"
# 3. Setup lxd for rockcraft
# lxd init --auto
# 4. Install rockcraft
# snap install rockcraft --classic
# 5. Install and configure docker for running the rock
# snap install docker
# sudo addgroup --system docker
# sudo adduser $USER docker
# newgrp docker
# sudo snap disable docker
# sudo snap enable docker

set -e
