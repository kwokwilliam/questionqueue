#!/usr/bin/env bash
# run rabbit
docker run -d \
--name questionqueuerabbit \
--network questionqueue \
-p 5672:5672 \
rabbitmq:3-management