#!/usr/bin/env bash

docker rm -f questionqueue
docker rmi ricowang/questionqueue

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build -t ricowang/questionqueue:latest .
go clean
#docker push ricowang/questionqueue:latest

# linux
#docker run \
#    --network host \
#    --name questionqueue \
#    -e ADDR=":80" \
#    -e MONGOADDR="mongodb://127.0.0.1:27017" \
#    -e REDISADDR="127.0.0.1:6379" \
#    -e RABBITADDR="amqp://guest:guest@127.0.0.1:5672" \
#    ricowang/questionqueue:latest

# mac
docker run \
    -p 8080:80 \
    --network api \
    --name questionqueue \
    -e ADDR=":80" \
    -e MONGOADDR="mongodb://mongo:27017" \
    -e REDISADDR="redis:6379" \
    -e RABBITADDR="amqp://guest:guest@rabbitmq:5672" \
    ricowang/questionqueue:latest
