#!/usr/bin/env bash
# Remove any instance of admin microservice
docker rm -f admin-micro

# Pull from DockerHub
docker pull ricowang/admin-micro:latest

# Export environment variables
export MONGO_URI="mongodb://questionqueuemongo:27017/question_queue"
export REDIS_HOST="questionqueueredis"
export REDIS_PORT=6379
export QUEUE_NAME="questionqueuerabbit"
export RABBIT_HOST="questionqueuerabbit"
export ADMIN_HOST="admin-micro"
export ADMIN_PORT=8001

# Run microservice
docker run -d \
    --name admin-micro \
    --network questionqueue \
    -e MONGO_URI=$MONGO_URI \
    -e REDIS_HOST=$REDIS_HOST \
    -e REDIS_PORT=$REDIS_PORT \
    -e QUEUE_NAME=$QUEUE_NAME \
    -e RABBIT_HOST=$RABBIT_HOST \
    -e ADMIN_HOST=$ADMIN_HOST \
    -e ADMIN_PORT=$ADMIN_PORT \
    --restart unless-stopped \
    ricowang/admin-micro:latest