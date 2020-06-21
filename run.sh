#!/usr/bin/env bash
set -ex
./build.sh
if ! docker volume ls --format '{{.Name}}' | grep oracle-persistance > /dev/null; then
  docker volume create oracle-persistance
fi
docker run \
  --name "my-oracle" \
  --restart=always \
  --detach \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v oracle-persistance:/opt/oracle/oradata \
  --env "ORACLE_PWD=password1234" \
  -p 1521:1521 \
  -p 5500:5500 \
  "my-oracle:latest"
