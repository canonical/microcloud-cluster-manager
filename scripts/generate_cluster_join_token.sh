#!/usr/bin/env bash
set -e

# if flag --local exists then use go run to run lxd-cluster-mgr
if [ "$1" == "--local" ]; then
    COMMAND="go run ./cmd/lxd-cluster-mgr"
else
    COMMAND="./lxd-cluster-mgr"
fi

$COMMAND --state-dir ./state/dir1 remote-cluster-join-token add cluster1 --expiry 24h
