#!/usr/bin/env bash
# Set GOOS environment variable
export GOOS="linux"

# Build go executable for linux
go build

# Build the docker container
docker build -t ricowang/gateway .

# Delete go executable for linux
go clean