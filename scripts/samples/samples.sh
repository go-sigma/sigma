#!/bin/sh

docker login sigma.tosone.cn -u sigma -p sigma

docker pull redis:7

docker tag redis:7 sigma.tosone.cn/library/redis:7
docker push sigma.tosone.cn/library/redis:7

curl https://github.com/grafana/k6/releases/download/v0.46.0/k6-v0.46.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1
