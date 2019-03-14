#!/usr/bin/env bash

docker run \
    --network questionqueue \
    --name rw \
    -e ADDR=":8000" \
    -e MONGOADDR="mongodb://questionqueuemongo:27017" \
    -e REDISADDR="questionqueueredis:6379" \
    -e RABBITADDR="amqp://guest:guest@questionqueuerabbit:5672" \
    ricowang/questionqueue:latest
