#!/usr/bin/env bash
# run mongodb
docker run -d \
-p 27017:27017 \
--name questionqueuemongo \
--network questionqueue \
mongo