#!/usr/bin/env bash
set -ex
./reset.sh
if docker image ls --format '{{.Repository}}:{{.Tag}}' | grep "my-oracle:latest" > /dev/null; then
  docker image rm --force --no-prune my-oracle:latest
fi
