#!/usr/bin/env bash

function is-healthy() {
  docker_status="$(docker ps --filter name=my-oracle --format '{{.Status}}')"
  [[ "${docker_status}" =~ .*(healthy) ]]
}

function wait-healthy() {
  echo "Please wait, oracle is performing it's setup/startup"
  until is-healthy; do
    sleep 0.25s
    printf '.'
  done
  echo "!"
  echo "Oracle Database is ready"
}

function is-ready() {
  docker exec my-oracle /bin/healthcheck --check-user-ready
}

function wait-is-ready() {
  echo "Please wait, checking for database user availability"
  until is-ready; do
    sleep 0.25s
    printf '.'
  done
  echo "!"
  echo "Users are ready"
}
