#!/usr/bin/env bash
# Build the admin queue microservice and push to DockerHub
docker build -t questionqueue/admin-micro .
docker push questionqueue/admin-micro

# Deploy to AWS 
ssh ec2-user@apif.uwinfotutor.me 'bash -s' < run.sh