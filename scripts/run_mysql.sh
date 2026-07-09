#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" run -it \
  --name sigma-mysql \
  -e MYSQL_ROOT_PASSWORD=sigma \
  -e MYSQL_DATABASE=sigma \
  -e MYSQL_USER=sigma \
  -e MYSQL_PASSWORD=sigma \
  -p 3306:3306 -d --rm \
  --health-cmd "mysqladmin -usigma -psigma ping || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  mysql:8.0
