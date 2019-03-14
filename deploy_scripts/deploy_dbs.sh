#!/usr/bin/env bash

ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < ./clean.sh

ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'docker network rm api; docker network create questionqueue; exit;'

ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < ./deploy_mongo.sh
ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < ./deploy_redis.sh
ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < ./deploy_rabbit.sh

