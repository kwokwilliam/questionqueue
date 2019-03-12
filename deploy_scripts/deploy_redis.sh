# create docker network if not already existing
docker network create questionqueue

# Run redis
docker run -d \
--name questionqueueredis \
--network questionqueue \
redis