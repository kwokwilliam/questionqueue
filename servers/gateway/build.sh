# Set GOOS environment variable 
export GOOS="linux"

# Build go executable for linux
go build

# Build the docker container
docker build -t wkwok16/gateway .

# build the database
docker build -t wkwok16/gatewaydb ../db

# Delete go executable for linux
go clean