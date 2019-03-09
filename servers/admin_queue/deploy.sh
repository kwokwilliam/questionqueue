export MONGO_URI="mongodb://mongomessaging:27017/messaging"
export REDIS_HOST=""
export REDIS_PORT=""
export QUEUE_NAME=""
export RABBIT_HOST=""
export ADMIN_HOST=""

# Build the admin queue microservice and push to DockerHub
docker build -t ljandrea/api-messaging .
docker push ljandrea/api-messaging

# Deploy to AWS - need to complete
ssh ec2-user@api.uwinfotutor.me