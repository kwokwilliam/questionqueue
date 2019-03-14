#!/usr/bin/env bash
# create docker network if not already existing
docker network create questionqueue

# Run redis
docker run -d \
-p 6379:6379 \
--name questionqueueredis \
--network questionqueue \
redis