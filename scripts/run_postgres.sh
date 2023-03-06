#!/bin/bash

docker run -it \
  --name ximager-postgres \
  -e POSTGRES_PASSWORD=ximager \
  -e POSTGRES_USER=ximager \
  -e POSTGRES_DB=ximager \
  -p 5432:5432 -d --rm \
  --health-cmd "pg_isready -U ximager -d ximager || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:15-alpine
