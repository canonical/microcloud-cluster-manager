#!/bin/bash

set -e

# Install Cert-Manager if it's not already installed
echo "Checking if Cert-Manager is installed..."
if ! kubectl get ns cert-manager &>/dev/null; then
  echo "Cert-Manager not found, installing it..."
  kubectl apply -f deployment/k8s/base/cert/cert-manager.yaml
else
  echo "Cert-Manager is already installed."
fi

# Wait for Cert-Manager to be deployed and ready
echo "Waiting for Cert-Manager deployment to become available..."
kubectl wait --for=condition=available --timeout=300s deployment --all -n cert-manager

# Apply the ClusterIssuer if not already created
echo "Applying ClusterIssuer..."
kubectl apply -f deployment/k8s/base/cert/cert-issuer.yaml

# Apply the Certificate resource
echo "Applying Certificates..."
kubectl apply -f deployment/k8s/base/cert/management-cert.yaml
kubectl apply -f deployment/k8s/base/cert/control-cert.yaml

# Wait for the cert Secret to be created (it contains the certificate)
echo "Waiting for the certificate Secrets to be created..."
kubectl wait --for=create --timeout=600s secret/management-cert-secret -n default
kubectl wait --for=create --timeout=600s secret/control-cert-secret -n default

echo "Certificates is ready!"
echo "Proceeding to application deployment..."