#!/bin/bash

function err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
}

#######################################
# Build and push the bookstore image with the tag provided
# Arguments:
#   Tag to use for image
# Returns:
#   0 if docker build was pushed, non-zero on error.
#######################################
function docker_build_push() {
  local tag=$1

  docker build -t "bookstore:${tag}" . && \
    docker tag "bookstore:${tag}" "registry.digitalocean.com/at-docker/bookstore:${tag}" && \
    docker push "registry.digitalocean.com/at-docker/bookstore:${tag}"
}

tag=$1

if [[ -z "${tag}" ]]; then
  tag="latest"
fi 

if ! docker_build_push $tag; then
  err "Unable to push docker build"
  exit 1
fi