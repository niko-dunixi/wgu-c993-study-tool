#!/usr/bin/env bash
set -ex
./stop.sh
if docker volume ls --format '{{.Name}}' | grep oracle-persistance > /dev/null; then
  docker volume rm -f oracle-persistance
fi