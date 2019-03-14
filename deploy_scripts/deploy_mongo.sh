#!/usr/bin/env bash
# create docker network if not already existing
docker network create questionqueue

# run mongodb
docker run -d \
--name questionqueuemongo \
--network questionqueue \
mongo