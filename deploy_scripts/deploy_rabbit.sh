#!/usr/bin/env bash
# run rabbit
docker run -d \
-p 5672:5672 \
--name questionqueuerabbit \
--network questionqueue \
rabbitmq:3-management