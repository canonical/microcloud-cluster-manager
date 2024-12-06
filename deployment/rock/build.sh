#!/bin/sh

# The script requires:
# - rockcraft
# - docker

set -e

rockcraft pack -v

# Use the value of IMAGE if set, otherwise call `make docker-image-name`
# Locally, $IMAGE may be set by skaffold for custom artifact builders
IMAGE=${IMAGE:-$(make docker-image-name)}

rockcraft.skopeo --insecure-policy copy \
    oci-archive:$(make rock-name) \
    docker-daemon:$IMAGE --debug

# remove rock after copying over to docker daemon
rm $(make rock-name)