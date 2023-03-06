#!/bin/bash

docker run -it \
  --name ximager-mysql \
  -e MYSQL_ROOT_PASSWORD=ximager \
  -e MYSQL_DATABASE=ximager \
  -e MYSQL_USER=ximager \
  -e MYSQL_PASSWORD=ximager \
  -p 3306:3306 -d --rm \
  --health-cmd "mysqladmin -uximager -pximager ping || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  mysql:8.0
