#!/usr/bin/env bash
# Build the admin queue microservice and push to DockerHub
docker build -t ricowang/admin-micro .
docker push ricowang/admin-micro:latest

# Deploy to AWS 
ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < run.sh