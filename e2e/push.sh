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
  echo "sigma cannot be available within one minute."
  exit 1
fi

set -e

ADMIN_PASSWORD=$(docker logs sigma 2>&1 | sed -n 's/.*auto-generated password.*password=\([A-Za-z0-9]*\).*/\1/p' | tail -n 1)
if [ -z "$ADMIN_PASSWORD" ]; then
  echo "cannot find the auto-generated admin password in the sigma logs."
  docker logs sigma
  exit 1
fi
if [ -n "$GITHUB_ENV" ]; then
  echo "SIGMA_ADMIN_PASSWORD=$ADMIN_PASSWORD" >> "$GITHUB_ENV"
fi

docker pull hello-world:latest
docker tag hello-world:latest 127.0.0.1:3000/library/hello-world:latest
docker pull mysql:8
docker tag mysql:8 127.0.0.1:3000/library/mysql:8

docker login 127.0.0.1:3000 -u sigma -p "$ADMIN_PASSWORD"

docker push 127.0.0.1:3000/library/hello-world:latest
docker pull 127.0.0.1:3000/library/hello-world:latest
docker push 127.0.0.1:3000/library/mysql:8
docker pull 127.0.0.1:3000/library/mysql:8

docker logs sigma
