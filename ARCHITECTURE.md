# Cluster Manager Architecture

## System overview
LXD Cluster Manager is a centralized tool for viewing and managing LXD clusters. It is a highly available web application with a browser-based UI that facilitates user interaction with the system.

Since LXD clusters are often hosted in air-gapped environments, it is assumed that Cluster Manager cannot directly reach an LXD cluster. This implies that network communication is unidirectional, from LXD clusters to Cluster Manager only.

The Cluster Manager requires an OIDC provider to be set up for user authentication. Once authenticated, users will be able to access the UI fully and manage registered LXD clusters.

To connect an LXD cluster, a join token must be generated in Cluster Manager and manually sent to an LXD admin. The join token will need to be consumed by an LXD cluster to register Cluster Manager details such as available network addresses. The LXD cluster will then send a join request to Cluster Manager, with the payload signed using an HMAC key generated from the join token secret. Once Cluster Manager receives the join request, it will validate the HMAC key. If successful, it will store the LXD cluster details with a "PENDING_APPROVAL" status. A Cluster Manager user will need to approve or reject the join request.

Once the LXD cluster is connected, it will send status updates to Cluster Manager at periodic intervals. The data will be stored by Cluster Manager and displayed via the browser UI. Communication between an LXD cluster and Cluster Manager after the initial join request will be authenticated using mTLS.

<!-- TODO: add section for CICD -->
<!-- TODO: add section for testing -->
<!-- TODO: add section for k8s architecture -->