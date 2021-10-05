#!/bin/sh

set -eux

IS_CONTAINER=${IS_CONTAINER:-false}
CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-podman}"

if [ "${IS_CONTAINER}" != "false" ]; then
  export XDG_CACHE_HOME="/tmp/.cache"

  gosec -severity medium --confidence medium -quiet ./...
else
  "${CONTAINER_RUNTIME}" run --rm \
    --env IS_CONTAINER=TRUE \
    --volume "${PWD}:/go/src/github.com/yashpatil17/baremetal-operator:ro,z" \
    --entrypoint sh \
    --workdir /go/src/github.com/yashpatil17/baremetal-operator \
    registry.hub.docker.com/securego/gosec:latest \
    /go/src/github.com/yashpatil17/baremetal-operator/hack/gosec.sh "${@}"
fi;
