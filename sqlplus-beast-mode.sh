#!/usr/bin/env bash
set -e
source ./bash-common.sh
wait-healthy
docker exec -it my-oracle /bin/bash -c 'sqlplus sys/$ORACLE_PWD@//localhost:1521/$ORACLE_PDB as sysdba'
