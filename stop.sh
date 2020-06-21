#!/usr/bin/env bash
set -ex
if docker container ls --format '{{.Names}}' | grep "my-oracle" >/dev/null; then
  docker stop --time 0 "my-oracle"
  docker rm -f "my-oracle"
fi
