#!/bin/bash

docker run -it \
  --name ximager-postgres \
  -e POSTGRES_PASSWORD=ximager \
  -e POSTGRES_USER=ximager \
  -e POSTGRES_DB=ximager \
  -p 5432:5432 -d --rm \
  postgres:15-alpine
