# create docker network if not already existing
docker network create questionqueue

# run rabbit
docker run -d \
--name questionqueuerabbit \
--network questionqueue \
-p 5672:5672 \
rabbitmq:3-management