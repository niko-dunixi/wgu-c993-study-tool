#!/usr/bin/env bash
set -ex

function image-exists() {
  docker image ls --format '{{.Repository}}:{{.Tag}}' | grep "${1}" > /dev/null
}

if ! image-exists "oracle/database:12.2.0.1-ee"; then
  if ! image-exists "my-oracle-builder:latest"; then
    read -p "Oracle Username: " oracle_username
    read -sp "Oracle Password: " oracle_password
    read -p "Do you agree to oracle terms of service? (Only 'I agree' will work)" oracle_tos
    docker build \
      --build-arg "ORACLE_USERNAME=${oracle_username}" \
      --build-arg "ORACLE_PASSWORD=${oracle_password}" \
      --build-arg "ORACLE_AGREE_TO_TERMS_OF_SERVICE=${oracle_tos}" \
      --target ORACLE_BUILDER \
      -t my-oracle-builder:latest \
      -f Dockerfile.builder
      .
  fi
  docker run \
    --rm -it \
    -v /var/run/docker.sock:/var/run/docker.sock \
    --name my-oracle-builder \
    my-oracle-builder:latest
fi
if ! image-exists "my-oracle:latest"; then
  docker build --target MY_ORACLE -t my-oracle:latest .
fi
