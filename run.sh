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
  -p 1521:1521 \
  -p 5500:5500 \
  "my-oracle:latest"
function is-healthy() {
  docker_status="$(docker ps --filter name=my-oracle --format '{{.Status}}')"
  [[ "${docker_status}" =~ .*(healthy) ]]
}
set +x
echo "Please wait, oracle is performing it's setup/startup"
until is-healthy; do
  sleep 2s
  printf '.'
done
echo "!"
echo "Oracle Database is ready, performing my database seeding."
docker exec -it my-oracle /bin/data-generator