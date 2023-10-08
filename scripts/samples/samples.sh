#!/bin/sh

docker login sigma.tosone.cn -u sigma -p sigma

docker pull redis:7

docker tag redis:7 sigma.tosone.cn/library/redis:7
docker push sigma.tosone.cn/library/redis:7
