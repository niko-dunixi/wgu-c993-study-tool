#!/usr/bin/env bash
set -e
source ./bash-common.sh
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
  -p 1521:1521 \
  -p 5500:5500 \
  "my-oracle:latest"
wait-healthy
docker exec -it my-oracle /bin/database-hydrator