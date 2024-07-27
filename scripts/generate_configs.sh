#!/usr/bin/env bash
set -e

# if flag --local exists then use go run to run lxd-cluster-mgr
if [ "$1" == "--local" ]; then
    COMMAND="go run ./cmd/lxd-cluster-mgr"
else
    COMMAND="./lxd-cluster-mgr"
fi

# default environmental values if not defined in .env.local
NUM_MEMBERS=1
GLOBAL_ADDRESS=""
POPULATE_MEMBER_EXTERNAL_ADDRESSES=false

# Load environment configs
ENV_FILE="./ui/.env.local"

if [ -f "$ENV_FILE" ]; then
  echo "Loading environment variables from $ENV_FILE"
  set -o allexport; source "$ENV_FILE"; set +o allexport
else
  echo "Environment file $ENV_FILE not found!"
  exit 1
fi

# populate manager oidc configs, these should always be set for Cluster Manager to work
$COMMAND \
    --state-dir ./state/dir1 \
    config set \
        oidc.issuer=$OIDC_ISSUER \
        oidc.client.id=$OIDC_CLIENT_ID \
        oidc.audience=$OIDC_AUDIENCE

# optionally populate manager global address
if [ -n "$GLOBAL_ADDRESS" ]; then
    $COMMAND \
        --state-dir ./state/dir1 \
        config set global.address=$GLOBAL_ADDRESS
fi

# populate manager member configs
# optionally populate member external_address
for i in $(seq 1 $NUM_MEMBERS); do
    MEMBER_NAME="member$i"

    if [ $POPULATE_MEMBER_EXTERNAL_ADDRESSES = true ]; then
        $COMMAND \
            --state-dir ./state/dir$i \
            config set $MEMBER_NAME external_address=0.0.0.0:901$i
    fi
done
