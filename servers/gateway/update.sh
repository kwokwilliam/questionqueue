# remove any instance of gateway
docker rm -f questionqueue-gateway

# pull new docker container from docker hub
docker pull ricowang/gateway

# make sure TLSCERT and TLSKEY exports are set
export TLSCERT=/etc/letsencrypt/live/apif.uwinfotutor.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/apif.uwinfotutor.me/privkey.pem

# Set addresses for microservices
export REDISADDR="questionqueueredis:6379"
export RABBITADDR="amqp://questionqueuerabbit:5672"
export RABBITQUEUENAME="rabbitqueue"
export REDISQUEUENAME="queue"
export RWADDRS="http://rw:8000"
export AJADDRS="http://admin-micro:8001"

# create docker network if not already existing
docker network create questionqueue

# run gateway
docker run -d \
--name questionqueue-gateway \
--network questionqueue \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e REDISADDR=$REDISADDR \
-e RABBITADDR=$RABBITADDR \
-e RABBITQUEUENAME=$RABBITQUEUENAME \
-e REDISQUEUENAME=$REDISQUEUENAME \
-e RWADDRS=$RWADDRS \
-e AJADDRS=$AJADDRS \
--restart unless-stopped \
ricowang/gateway

exit