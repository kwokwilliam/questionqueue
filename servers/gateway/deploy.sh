# Call build script
./build.sh

# Push API docker container to docker hub
docker push questionqueue/gateway

# cat setpasswords.sh update.sh > scriptToSend.sh
# SSH into `api.uwinfotutor.me` and run update.sh
ssh ec2-user@api.uwinfotutor.me 'bash -s' < update.sh