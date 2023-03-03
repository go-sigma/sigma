#!/bin/bash

docker run -it \
  --name ximager-redis \
  -p 6379:6379 -d --rm \
  redis:7.0-alpine --requirepass ximager
