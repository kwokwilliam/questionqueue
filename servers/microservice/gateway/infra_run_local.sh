#!/usr/bin/env bash

docker run -d --network host --name mongo mongo:latest
docker run -d --network host --name redis redis
docker run -d --network host --name rabbitmq rabbitmq:3