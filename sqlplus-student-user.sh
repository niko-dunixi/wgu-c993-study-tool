#!/usr/bin/env bash
set -e
source ./bash-common.sh
wait-is-ready
docker exec -it my-oracle /bin/bash -c 'sqlplus student/$USER_PASSWORD@//localhost:1521/$ORACLE_PDB'
