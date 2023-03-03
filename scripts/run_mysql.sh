#!/bin/bash

docker run -it \
  --name ximager-mysql \
  -e MYSQL_ROOT_PASSWORD=ximager \
  -e MYSQL_DATABASE=ximager \
  -e MYSQL_USER=ximager \
  -e MYSQL_PASSWORD=ximager \
  -p 3306:3306 -d --rm \
  mysql:8.0
