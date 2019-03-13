# docker remove network if existing
docker network disconnect -f slack gatewaydb
docker network disconnect -f slack gateway
docker network disconnect -f slack redis
docker network rm slack
# Remove any instance of wkwok16/gateway
docker rm -f gateway
docker rm -f gatewaydb
docker rm -f redisdb

# clean
docker image prune -f
docker container prune -f
docker volume prune -f
#

# Pull docker container from docker hub
docker pull wkwok16/gateway
docker pull wkwok16/gatewaydb

# Make sure TLSCERT and TLSKEY exports are set
export TLSCERT=/etc/letsencrypt/live/api.uwinfotutor.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/api.uwinfotutor.me/privkey.pem
export REDISADDR="redisdb:6379"




# create docker network
docker network create slack

# Run redis
# -p 6379:6379 \
docker run -d \
--name redisdb \
--network slack \
redis

# Run mysqlstore
# -p 3306:3306 \
docker run -d \
--name gatewaydb \
--network slack \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=gatewaydb \
wkwok16/gatewaydb

# Run the new docker container
docker run -d \
--name gateway \
--network slack \
-p 443:443 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e SESSIONKEY=$SESSIONKEY \
-e REDISADDR=$REDISADDR \
-e DSN=$DSN \
-e AWSACCESS=$AWSACCESS \
-e AWSSECRET=$AWSSECRET \
--restart unless-stopped \
wkwok16/gateway

exit