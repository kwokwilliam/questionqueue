#!/usr/bin/env bash

docker pull ricowang/questionqueue:latest

docker run \
    --network questionqueue \
    --name rw \
    -p 8000:8000 \
    -e ADDR=":8000" \
    -e MONGOADDR="mongodb://questionqueuemongo:27017" \
    -e REDISADDR="questionqueueredis:6379" \
    -e RABBITADDR="amqp://guest:guest@questionqueuerabbit:5672" \
    ricowang/questionqueue:latest
