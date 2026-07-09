---
title: 配置
---

# 配置

下面是 sigma 的完整配置项示例。生产环境中，数据库、Redis、对象存储和 JWT 私钥等敏感配置应通过 Secret 或环境变量注入。

```yaml
# this is a full config file, just show all available config items.

log:
  level: debug
  proxyLevel: info

database:
  # The database type to use. Supported types are: sqlite3, turso, mysql, postgresql
  type: sqlite3
  sqlite3:
    path: /var/lib/sigma/sigma.db
  turso:
    dsn: /var/lib/sigma/sigma-turso.db
  mysql:
    host: localhost
    port: 3306
    username: sigma
    password: sigma
    database: sigma
  postgresql:
    host: localhost
    port: 5432
    username: sigma
    password: sigma
    database: sigma
    sslMode: disable

# deploy available: single, replica
# replica should use external redis
deploy: single

redis:
  # redis type available: none, external. Following all of redis config just use reference here.
  # none: means never use redis
  # external: means use the specific redis instance
  type: none
  url: redis://:sigma@localhost:6379/0

cache:
  # the cache type available is: redis, inmemory
  # please attention in multi-node mode, you should use redis
  type: inmemory
  # cache key prefix
  prefix: sigma-cache
  inmemory:
    size: 10240
  redis:
    ttl: 72h

workqueue:
  # the workqueue type available: redis, database, inmemory
  type: redis
  redis:
    concurrency: 10
    # reject new messages when redis queue backlog reaches this value; 0 means unlimited
    maxBacklog: 0
  database: {}
  inmemory:
    concurrency: 1024
    maxBacklog: 1000

locker:
  # the locker type available: redis, inmemory
  type: redis
  prefix: sigma-locker
  inmemory: {}
  redis: {}

namespace:
  # push image to registry, if namespace not exist, it will be created automatically
  autoCreate: false
  # the automatic created namespace visibility, available: public, private
  visibility: public

http:
  # endpoint can be a domain or domain with port, eg: http://sigma.test.io, https://sigma.test.io:30080, http://127.0.0.1:3000
  # this endpoint will be used to generate the token service url in auth middleware,
  # you can leave it blank and it will use http://127.0.0.1:3000 as internal domain by default,
  # because the front page need show this endpoint.
  endpoint:
  # in some cases, daemon may pull image and scan it, but we don't want to pull image from public registry domain,
  # so use this internal domain to pull image from registry.
  # you can leave it blank and it will use http://127.0.0.1:3000 as internal domain by default.
  # in k8s cluster, it will be set to the distribution service which is used to pull image from registry, eg: http://registry.default.svc.cluster.local:3000
  # in docker-compose, it will be set to the registry service which is used to pull image from registry, eg: http://registry:3000
  # if http.tls.enabled is true, internalEndpoint should start with https://
  # eg: http://sigma.test.io, http://sigma.test.io:3000, https://sigma.test.io:30080
  internalEndpoint:
  # eg: http://sigma-distribution:3000
  internalDistributionEndpoint:
  timeout:
    readHeader: 30s
    # read and write cover full push/pull transfers; increase them for very large images or slow links
    read: 1h
    write: 1h
    idle: 2m
  bodyLimit:
    # request body size limits in bytes; 0 disables the limit for that route class
    enabled: true
    # default applies to regular API requests
    default: 10485760
    # manifest applies to PUT /v2/<name>/manifests/<reference>
    manifest: 104857600
    # blob applies to /v2/<name>/blobs/uploads requests; increase for large image pushes
    blob: 10737418240
  tls:
    enabled: false
    certificate: ./conf/sigma.test.io.crt
    key: ./conf/sigma.test.io.key

storage:
  rootdirectory: ./storage
  redirect: false
  type: filesystem
  filesystem:
    path: /var/lib/sigma/oci/
  s3:
    ak: sigma
    sk: sigma-sigma
    endpoint: http://127.0.0.1:9000
    region: cn-north-1
    bucket: sigma
    forcePathStyle: true
  cos:
    ak: sigma
    sk: sigma-sigma
    endpoint: https://hack-1251887554.cos.na-toronto.myqcloud.com
  oss:
    ak: sigma
    sk: sigma-sigma
    endpoint: http://127.0.0.1:9000
    forcePathStyle: true

# Notice: the tag never update after the first pulled from remote registry, unless you delete the image and pull again.
proxy:
  enabled: false
  endpoint: https://registry-1.docker.io
  tlsVerify: true
  username: ""
  password: ""

# daemon task config
daemon:
  builder:
    enabled: false
    image: ghcr.io/go-sigma/sigma-builder:nightly
    type: docker
    docker:
      sock:
      network: sigma
    kubernetes:
      kubeconfig:
      namespace: sigma-builder
    podman:
      uri: unix:///run/podman/podman.sock
  vulnerability:
    staleTimeout: 6h
    grype:
      dbCacheDir: ""

auth:
  anonymous:
    # anonymous will disabled if auth.anonymous.enabled set false
    enabled: true
  admin:
    username: sigma
    password: Admin@123
  token:
    realm: ""
    service: ""
  jwt:
    ttl: 6h
    refreshTTL: 72h
    # generate the key with: openssl genpkey -algorithm ED25519 | base64
    privateKey: "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1DNENBUUF3QlFZREsyVndCQ0lFSUFZY0NvYldvT2V2dnNjOWRlcStab1JFS2lyWW52MW1MRUFpZzlySDNmOHMKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="
  oauth2:
    github:
      # github login will disable if auth.oauth.github.enabled set false
      enabled: false
      clientId: "e5f9fa9e372dfac66aed"
      clientSecret: "49ab83f4d0665f8579516f7a3f2f753a6a57189b"
    gitlab:
      # gitlab login will disable if auth.oauth.gitlab.enabled set false
      enabled: false
      clientId: "4df6efcf8c319efb73e8116c72d881c559ccaf822096220a13cee3047b05ed70"
      clientSecret: "94ceddf22fc1560f33caec6be32c9c61a91719bd2df3b5127ccd43187192f95b"

mcp:
  enabled: false
  path: /api/v1/mcp
  transport: streamable_http
  auditWrites: true
  toolTimeout: 30s
  maxRequestBody: 1048576

audit:
  # record completed authenticated business HTTP requests
  enabled: true
  # record list-style GET requests, such as namespace, tag, and catalog lists
  recordListGet: false
  cleanup:
    # soft delete audit records older than this many months
    softDeleteAfterMonths: 3
    # hard delete audit records older than this many months
    hardDeleteAfterMonths: 6
    # run audit cleanup once per week by default
    interval: 168h
    # maximum audit records changed per cleanup iteration
    batchSize: 1000
```
