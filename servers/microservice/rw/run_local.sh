#!/usr/bin/env bash

docker network rm api
docker network create api

./infra_run_local.sh
sleep 5
./service_run_local.sh
