#!/usr/bin/env bash

docker rm -f mongo redis rabbitmq

# linux
#docker run -d --network host --name mongo mongo:latest
#docker run -d --network host --name redis redis
#docker run -d --network host --name rabbitmq rabbitmq:3

# mac
#docker run -d -p 27017:27017 --name mongo mongo:latest
#docker run -d -p 6379:6379 --name redis redis
#docker run -d -p 5672:5672 --name rabbitmq rabbitmq:3

docker run -d --network api --name mongo --hostname mongo mongo:latest
docker run -d --network api --name redis --hostname redis redis
docker run -d --network api --name rabbitmq --hostname rabbitmq rabbitmq:3
