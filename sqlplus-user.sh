#!/usr/bin/env bash
source ./common-functions.sh

function main() {
    echo "Waiting for container to come up"
    until container-is-healthy; do
        printf '.'
        sleep 2s
    done
    docker-sql-plus-user
}

main
