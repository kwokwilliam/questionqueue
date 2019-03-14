#!/usr/bin/env bash
# create docker network if not already existing
docker network create questionqueue

# run mongodb
docker run -d \
-p 27017:27017 \
--name questionqueuemongo \
--network questionqueue \
mongo