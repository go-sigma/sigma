#!/bin/sh

set +e

TIMES=12
while [ $TIMES -gt 0 ]; do
  STATUS=$(curl --insecure -s -o /dev/null -w '%{http_code}' http://127.0.0.1:3000/healthz)
  if [ "$STATUS" -eq 200 ]; then
    break
  fi
  TIMES=$((TIMES - 1))
  sleep 5
done

if [ $TIMES -eq 0 ]; then
  echo "XImager cannot be available within one minute."
  exit 1
fi

set -e

docker pull hello-world:latest
docker tag hello-world:latest 127.0.0.1:3000/library/hello-world:latest
docker pull mysql:8
docker tag mysql:8 127.0.0.1:3000/library/mysql:8

docker login 127.0.0.1:3000 -u ximager -p ximager

docker push 127.0.0.1:3000/library/hello-world:latest
docker pull 127.0.0.1:3000/library/hello-world:latest
docker push 127.0.0.1:3000/library/mysql:8
docker pull 127.0.0.1:3000/library/mysql:8
