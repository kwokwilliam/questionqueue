#!/usr/bin/env bash

./clean.sh

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build -t ricowang/questionqueue:latest .
go clean

docker push ricowang/questionqueue:latest

ssh -i ~/.ssh/id_rsa ec2-user@apif.uwinfotutor.me 'bash -s' < ./run_remote.sh