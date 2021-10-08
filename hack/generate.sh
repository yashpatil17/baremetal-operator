#!/bin/sh

# Ignore the rule that says we should always quote variables, because
# in this script we *do* want globbing.
# shellcheck disable=SC2086

set -eux

IS_CONTAINER=${IS_CONTAINER:-false}
ARTIFACTS=${ARTIFACTS:-/tmp}
CONTAINER_RUNTIME="${CONTAINER_RUNTIME:-podman}"

if [ "${IS_CONTAINER}" != "false" ]; then
  eval "$(go env)"
  cd "${GOPATH}"/src/github.com/yashpatil17/baremetal-operator
  export XDG_CACHE_HOME="/tmp/.cache"

  INPUT_FILES="$(git ls-files config) $(git ls-files | grep zz_generated)"
  cksum $INPUT_FILES > "$ARTIFACTS/lint.cksums.before"
  export VERBOSE="--verbose"
  make generate manifests
  cksum $INPUT_FILES > "$ARTIFACTS/lint.cksums.after"
  diff "$ARTIFACTS/lint.cksums.before" "$ARTIFACTS/lint.cksums.after"

else
  "${CONTAINER_RUNTIME}" run --rm \
    --env IS_CONTAINER=TRUE \
    --env DEPLOY_KERNEL_URL=https://tarballs.opendev.org/openstack/ironic-python-agent/dib/files/ipa-centos8-stable-wallaby.kernel \
    --env DEPLOY_RAMDISK_URL=https://tarballs.opendev.org/openstack/ironic-python-agent/dib/files/ipa-centos8-stable-wallaby.initramfs \
    --env IRONIC_ENDPOINT=http://localhost:6385/v1/ \
    --env IRONIC_INSPECTOR_ENDPOINT=http://localhost:5050/v1/ \
    --volume "${PWD}:/go/src/github.com/yashpatil17/baremetal-operator:rw,z" \
    --entrypoint sh \
    --workdir /go/src/github.com/yashpatil17/baremetal-operator \
    registry.hub.docker.com/library/golang:1.16 \
    /go/src/github.com/yashpatil17/baremetal-operator/hack/generate.sh "${@}"
fi;
