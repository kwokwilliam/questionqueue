#!/usr/bin/env bash

docker pull ricowang/rw:latest

docker rm -f rw

docker run \
    -d \
    --network questionqueue \
    --name rw \
    -p 8000:8000 \
    -e ADDR=":8000" \
    -e MONGOADDR="mongodb://questionqueuemongo:27017" \
    -e REDISADDR="questionqueueredis:6379" \
    -e RABBITADDR="amqp://questionqueuerabbit:5672" \
    ricowang/rw:latest

docker ps -a