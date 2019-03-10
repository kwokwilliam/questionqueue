# Note - edit MONGO_URI and QUEUE_NAME depending on agreed environment variables
export MONGO_URI="mongodb://mongomessaging:27017/question_queue"
export REDIS_HOST="redisdb"
export REDIS_PORT=6379
export QUEUE_NAME=""
export RABBIT_HOST="rabbit"
export ADMIN_HOST="admin"

# Build the admin queue microservice and push to DockerHub
docker build -t ljandrea/api-messaging .
docker push ljandrea/api-messaging

# Deploy to AWS - need to complete
ssh ec2-user@api.uwinfotutor.me