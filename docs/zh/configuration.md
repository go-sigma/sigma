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
  # namespaces created at startup; each entry is created when missing,
  # and each entry may set its own visibility (public or private)
  initialize:
    - name: library
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
  # the endpoint the builder container uses to push the image and call the sigma API; it defaults
  # to internalEndpoint. Set it when the builder resolves the server differently from the daemons
  # on the host, eg. http://sigma-distribution:3000 in docker-compose or
  # http://host.docker.internal:3000 when the server runs on the host and the builder in docker.
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
  # sigma 可使用的存储总量（字节），管理员首页用它展示存储使用率；0 表示不限制。
  capacity: 0
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
    tlsVerify: true
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
    # 留空表示首次启动时随机生成密码。生成的密码只会打印一次到服务端日志，例如：
    #   level=WARN msg="admin user created with an auto-generated password, please record it and
    #   change it after login" username=sigma password=62AWdMKiKb
    # 请记录该密码并用它登录，登录后立即修改：仍在使用自动生成密码的账号，登录页会先要求设置
    # 新密码，然后才允许进入系统。
    # 如果在这里显式填写密码，则不会生成随机密码，也不会打印该日志。
    password: ""
    email: sigma@sigma.io
  token:
    realm: ""
    service: ""
  jwt:
    ttl: 6h
    refreshTTL: 72h
    # base64 编码的 Ed25519 私钥，可用以下命令生成：
    # openssl genpkey -algorithm ED25519 | base64
    # 留空则在首次启动时生成并写入 settings 表，后续启动直接复用；
    # 配置了值则始终以配置为准。
    privateKey: ""
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
  # throttle failed logins by username: once maxFailures is reached, each
  # authentication attempt is delayed by delay instead of being rejected
  loginRateLimit:
    enabled: true
    maxFailures: 5
    # failure counter window; it resets after this with no failures
    window: 5m
    # fixed delay applied to every authentication attempt after maxFailures
    delay: 1s
  # master key used to encrypt source code credentials (access tokens, ssh keys,
  # passwords) at rest. when empty, sigma generates one on first start and stores it in
  # the settings table. rotating it makes the stored credentials unreadable, so users have
  # to enter them again.
  # generate the key with: openssl rand -base64 32
  credentialKey: ""

scm:
  # source code management providers available for code repository sync and image builds.
  # this section is independent of auth.oauth2, which only decides whether a provider can be
  # used to sign in. when empty, the providers are derived from the enabled oauth2 providers,
  # so an existing configuration keeps working.
  providers: []
  # providers:
  #   - provider: github
  #     baseUrl: https://github.com
  #   - provider: gitlab
  #     # self-hosted gitlab
  #     baseUrl: https://gitlab.example.com
  #   - provider: gitea
  #     # gitea has no default base url and is skipped when it is empty
  #     baseUrl: https://gitea.example.com

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

## 环境变量覆盖

上面每一项都可以用环境变量覆盖，容器化部署时不必再对配置文件做模板化。覆盖发生在
读取 YAML 之后：设置了的变量一定优先于配置文件，没有设置的变量一定保留配置文件里的值。

### 命名规则

变量名是 `SIGMA_` 加上配置在文件中的路径，路径用 `_` 连接并全部大写：

| 配置文件路径 | 环境变量 |
| --- | --- |
| `daemon.builder.enabled` | `SIGMA_DAEMON_BUILDER_ENABLED` |
| `daemon.builder.type` | `SIGMA_DAEMON_BUILDER_TYPE` |
| `daemon.builder.image` | `SIGMA_DAEMON_BUILDER_IMAGE` |
| `daemon.builder.tlsVerify` | `SIGMA_DAEMON_BUILDER_TLSVERIFY` |
| `daemon.builder.extra_hosts` | `SIGMA_DAEMON_BUILDER_EXTRA_HOSTS` |
| `auth.jwt.privateKey` | `SIGMA_AUTH_JWT_PRIVATEKEY` |

变量名不区分大小写，同一路径段内部的下划线可以省略，所以
`SIGMA_DAEMON_BUILDER_TLS_VERIFY`、`sigma_daemon_builder_tlsverify` 与
`SIGMA_DAEMON_BUILDER_TLSVERIFY` 是同一个配置项。

### 取值写法

| 类型 | 可用写法 | 示例 |
| --- | --- | --- |
| 字符串 | 原样使用 | `SIGMA_HTTP_ENDPOINT=https://registry.example.com` |
| 布尔 | `true`、`false`、`1`、`0` | `SIGMA_DAEMON_BUILDER_ENABLED=true` |
| 整数 | 十进制 | `SIGMA_DATABASE_MYSQL_PORT=3306` |
| 时长 | Go duration 语法 | `SIGMA_DAEMON_GC_RETENTION=48h` |
| 枚举 | 文档中列出的取值，启动时校验 | `SIGMA_DAEMON_BUILDER_TYPE=docker` |
| 浮点 | 十进制，用 `.` 作小数点 | `SIGMA_TRACE_SAMPLERATIO=0.25` |
| 列表 | 逗号分隔，自动去掉首尾空格 | `SIGMA_REDIS_ADDRS=redis-0:6379,redis-1:6379` |

无法解析的值、未知的枚举值、以及不是 `true/false/1/0` 的布尔值都会让启动直接失败，
报错信息同时包含变量名和配置路径，例如
`SIGMA_DAEMON_GC_BATCHSIZE (daemon.gc.batchSize): ... invalid syntax`。
同一次启动里的其它合法覆盖仍会先生效，然后才返回错误。

### 完整对照表

按字母顺序排列，覆盖配置文件的每一个叶子节点。

#### analytics

| config path | environment variable |
| --- | --- |
| `analytics.backend` | `SIGMA_ANALYTICS_BACKEND` |
| `analytics.counterBackend` | `SIGMA_ANALYTICS_COUNTERBACKEND` |
| `analytics.enabled` | `SIGMA_ANALYTICS_ENABLED` |
| `analytics.flushInterval` | `SIGMA_ANALYTICS_FLUSHINTERVAL` |
| `analytics.reconcileCron` | `SIGMA_ANALYTICS_RECONCILECRON` |
| `analytics.retentionDays` | `SIGMA_ANALYTICS_RETENTIONDAYS` |

#### audit

| config path | environment variable |
| --- | --- |
| `audit.cleanup.batchSize` | `SIGMA_AUDIT_CLEANUP_BATCHSIZE` |
| `audit.cleanup.hardDeleteAfterMonths` | `SIGMA_AUDIT_CLEANUP_HARDDELETEAFTERMONTHS` |
| `audit.cleanup.interval` | `SIGMA_AUDIT_CLEANUP_INTERVAL` |
| `audit.cleanup.softDeleteAfterMonths` | `SIGMA_AUDIT_CLEANUP_SOFTDELETEAFTERMONTHS` |
| `audit.enabled` | `SIGMA_AUDIT_ENABLED` |
| `audit.recordListGet` | `SIGMA_AUDIT_RECORDLISTGET` |

#### auth

| config path | environment variable |
| --- | --- |
| `auth.admin.email` | `SIGMA_AUTH_ADMIN_EMAIL` |
| `auth.admin.password` | `SIGMA_AUTH_ADMIN_PASSWORD` |
| `auth.admin.username` | `SIGMA_AUTH_ADMIN_USERNAME` |
| `auth.anonymous.enabled` | `SIGMA_AUTH_ANONYMOUS_ENABLED` |
| `auth.credentialKey` | `SIGMA_AUTH_CREDENTIALKEY` |
| `auth.jwt.privateKey` | `SIGMA_AUTH_JWT_PRIVATEKEY` |
| `auth.jwt.refreshTTL` | `SIGMA_AUTH_JWT_REFRESHTTL` |
| `auth.jwt.ttl` | `SIGMA_AUTH_JWT_TTL` |
| `auth.loginRateLimit.delay` | `SIGMA_AUTH_LOGINRATELIMIT_DELAY` |
| `auth.loginRateLimit.enabled` | `SIGMA_AUTH_LOGINRATELIMIT_ENABLED` |
| `auth.loginRateLimit.maxFailures` | `SIGMA_AUTH_LOGINRATELIMIT_MAXFAILURES` |
| `auth.loginRateLimit.window` | `SIGMA_AUTH_LOGINRATELIMIT_WINDOW` |
| `auth.oauth2.gitea.clientId` | `SIGMA_AUTH_OAUTH2_GITEA_CLIENTID` |
| `auth.oauth2.gitea.clientSecret` | `SIGMA_AUTH_OAUTH2_GITEA_CLIENTSECRET` |
| `auth.oauth2.gitea.enabled` | `SIGMA_AUTH_OAUTH2_GITEA_ENABLED` |
| `auth.oauth2.github.clientId` | `SIGMA_AUTH_OAUTH2_GITHUB_CLIENTID` |
| `auth.oauth2.github.clientSecret` | `SIGMA_AUTH_OAUTH2_GITHUB_CLIENTSECRET` |
| `auth.oauth2.github.enabled` | `SIGMA_AUTH_OAUTH2_GITHUB_ENABLED` |
| `auth.oauth2.gitlab.clientId` | `SIGMA_AUTH_OAUTH2_GITLAB_CLIENTID` |
| `auth.oauth2.gitlab.clientSecret` | `SIGMA_AUTH_OAUTH2_GITLAB_CLIENTSECRET` |
| `auth.oauth2.gitlab.enabled` | `SIGMA_AUTH_OAUTH2_GITLAB_ENABLED` |
| `auth.token.realm` | `SIGMA_AUTH_TOKEN_REALM` |
| `auth.token.service` | `SIGMA_AUTH_TOKEN_SERVICE` |

#### cache

| config path | environment variable |
| --- | --- |
| `cache.inmemory.size` | `SIGMA_CACHE_INMEMORY_SIZE` |
| `cache.negativeTtl` | `SIGMA_CACHE_NEGATIVETTL` |
| `cache.prefix` | `SIGMA_CACHE_PREFIX` |
| `cache.prewarm.enabled` | `SIGMA_CACHE_PREWARM_ENABLED` |
| `cache.prewarm.limit` | `SIGMA_CACHE_PREWARM_LIMIT` |
| `cache.prewarm.timeout` | `SIGMA_CACHE_PREWARM_TIMEOUT` |
| `cache.redis.ttl` | `SIGMA_CACHE_REDIS_TTL` |
| `cache.ttl` | `SIGMA_CACHE_TTL` |
| `cache.type` | `SIGMA_CACHE_TYPE` |

#### daemon

| config path | environment variable |
| --- | --- |
| `daemon.builder.docker.network` | `SIGMA_DAEMON_BUILDER_DOCKER_NETWORK` |
| `daemon.builder.docker.sock` | `SIGMA_DAEMON_BUILDER_DOCKER_SOCK` |
| `daemon.builder.enabled` | `SIGMA_DAEMON_BUILDER_ENABLED` |
| `daemon.builder.extra_hosts` | `SIGMA_DAEMON_BUILDER_EXTRA_HOSTS` |
| `daemon.builder.image` | `SIGMA_DAEMON_BUILDER_IMAGE` |
| `daemon.builder.kubernetes.kubeconfig` | `SIGMA_DAEMON_BUILDER_KUBERNETES_KUBECONFIG` |
| `daemon.builder.kubernetes.namespace` | `SIGMA_DAEMON_BUILDER_KUBERNETES_NAMESPACE` |
| `daemon.builder.podman.uri` | `SIGMA_DAEMON_BUILDER_PODMAN_URI` |
| `daemon.builder.tlsVerify` | `SIGMA_DAEMON_BUILDER_TLSVERIFY` |
| `daemon.builder.type` | `SIGMA_DAEMON_BUILDER_TYPE` |
| `daemon.gc.batchSize` | `SIGMA_DAEMON_GC_BATCHSIZE` |
| `daemon.gc.cron` | `SIGMA_DAEMON_GC_CRON` |
| `daemon.gc.lockExpire` | `SIGMA_DAEMON_GC_LOCKEXPIRE` |
| `daemon.gc.lockWaitTimeout` | `SIGMA_DAEMON_GC_LOCKWAITTIMEOUT` |
| `daemon.gc.recordBatchSize` | `SIGMA_DAEMON_GC_RECORDBATCHSIZE` |
| `daemon.gc.retention` | `SIGMA_DAEMON_GC_RETENTION` |
| `daemon.gc.workerCount` | `SIGMA_DAEMON_GC_WORKERCOUNT` |
| `daemon.sizeReconcile.batchSize` | `SIGMA_DAEMON_SIZERECONCILE_BATCHSIZE` |
| `daemon.sizeReconcile.interval` | `SIGMA_DAEMON_SIZERECONCILE_INTERVAL` |
| `daemon.sizeReconcile.lockExpire` | `SIGMA_DAEMON_SIZERECONCILE_LOCKEXPIRE` |
| `daemon.sizeReconcile.lockWaitTimeout` | `SIGMA_DAEMON_SIZERECONCILE_LOCKWAITTIMEOUT` |
| `daemon.vulnerability.grype.dbCacheDir` | `SIGMA_DAEMON_VULNERABILITY_GRYPE_DBCACHEDIR` |
| `daemon.vulnerability.staleTimeout` | `SIGMA_DAEMON_VULNERABILITY_STALETIMEOUT` |

#### database

| config path | environment variable |
| --- | --- |
| `database.connMaxIdleTime` | `SIGMA_DATABASE_CONNMAXIDLETIME` |
| `database.connMaxLifetime` | `SIGMA_DATABASE_CONNMAXLIFETIME` |
| `database.maxIdleConns` | `SIGMA_DATABASE_MAXIDLECONNS` |
| `database.maxOpenConns` | `SIGMA_DATABASE_MAXOPENCONNS` |
| `database.mysql.database` | `SIGMA_DATABASE_MYSQL_DATABASE` |
| `database.mysql.host` | `SIGMA_DATABASE_MYSQL_HOST` |
| `database.mysql.password` | `SIGMA_DATABASE_MYSQL_PASSWORD` |
| `database.mysql.port` | `SIGMA_DATABASE_MYSQL_PORT` |
| `database.mysql.username` | `SIGMA_DATABASE_MYSQL_USERNAME` |
| `database.postgresql.database` | `SIGMA_DATABASE_POSTGRESQL_DATABASE` |
| `database.postgresql.host` | `SIGMA_DATABASE_POSTGRESQL_HOST` |
| `database.postgresql.password` | `SIGMA_DATABASE_POSTGRESQL_PASSWORD` |
| `database.postgresql.port` | `SIGMA_DATABASE_POSTGRESQL_PORT` |
| `database.postgresql.sslMode` | `SIGMA_DATABASE_POSTGRESQL_SSLMODE` |
| `database.postgresql.username` | `SIGMA_DATABASE_POSTGRESQL_USERNAME` |
| `database.sqlite3.path` | `SIGMA_DATABASE_SQLITE3_PATH` |
| `database.turso.dsn` | `SIGMA_DATABASE_TURSO_DSN` |
| `database.type` | `SIGMA_DATABASE_TYPE` |

#### deploy

| config path | environment variable |
| --- | --- |
| `deploy` | `SIGMA_DEPLOY` |

#### http

| config path | environment variable |
| --- | --- |
| `http.bodyLimit.blob` | `SIGMA_HTTP_BODYLIMIT_BLOB` |
| `http.bodyLimit.default` | `SIGMA_HTTP_BODYLIMIT_DEFAULT` |
| `http.bodyLimit.enabled` | `SIGMA_HTTP_BODYLIMIT_ENABLED` |
| `http.bodyLimit.manifest` | `SIGMA_HTTP_BODYLIMIT_MANIFEST` |
| `http.endpoint` | `SIGMA_HTTP_ENDPOINT` |
| `http.internalDistributionEndpoint` | `SIGMA_HTTP_INTERNALDISTRIBUTIONENDPOINT` |
| `http.internalEndpoint` | `SIGMA_HTTP_INTERNALENDPOINT` |
| `http.timeout.idle` | `SIGMA_HTTP_TIMEOUT_IDLE` |
| `http.timeout.read` | `SIGMA_HTTP_TIMEOUT_READ` |
| `http.timeout.readHeader` | `SIGMA_HTTP_TIMEOUT_READHEADER` |
| `http.timeout.write` | `SIGMA_HTTP_TIMEOUT_WRITE` |
| `http.tls.certificate` | `SIGMA_HTTP_TLS_CERTIFICATE` |
| `http.tls.enabled` | `SIGMA_HTTP_TLS_ENABLED` |
| `http.tls.key` | `SIGMA_HTTP_TLS_KEY` |

#### locker

| config path | environment variable |
| --- | --- |
| `locker.prefix` | `SIGMA_LOCKER_PREFIX` |
| `locker.type` | `SIGMA_LOCKER_TYPE` |

#### log

| config path | environment variable |
| --- | --- |
| `log.level` | `SIGMA_LOG_LEVEL` |
| `log.proxyLevel` | `SIGMA_LOG_PROXYLEVEL` |

#### namespace

| config path | environment variable |
| --- | --- |
| `namespace.autoCreate` | `SIGMA_NAMESPACE_AUTOCREATE` |
| `namespace.visibility` | `SIGMA_NAMESPACE_VISIBILITY` |

#### proxy

| config path | environment variable |
| --- | --- |
| `proxy.enabled` | `SIGMA_PROXY_ENABLED` |
| `proxy.endpoint` | `SIGMA_PROXY_ENDPOINT` |
| `proxy.password` | `SIGMA_PROXY_PASSWORD` |
| `proxy.tlsVerify` | `SIGMA_PROXY_TLSVERIFY` |
| `proxy.token` | `SIGMA_PROXY_TOKEN` |
| `proxy.username` | `SIGMA_PROXY_USERNAME` |

#### redis

| config path | environment variable |
| --- | --- |
| `redis.addrs` | `SIGMA_REDIS_ADDRS` |
| `redis.db` | `SIGMA_REDIS_DB` |
| `redis.masterName` | `SIGMA_REDIS_MASTERNAME` |
| `redis.maxRetries` | `SIGMA_REDIS_MAXRETRIES` |
| `redis.minIdleConns` | `SIGMA_REDIS_MINIDLECONNS` |
| `redis.password` | `SIGMA_REDIS_PASSWORD` |
| `redis.poolSize` | `SIGMA_REDIS_POOLSIZE` |
| `redis.url` | `SIGMA_REDIS_URL` |
| `redis.username` | `SIGMA_REDIS_USERNAME` |

#### storage

| config path | environment variable |
| --- | --- |
| `storage.capacity` | `SIGMA_STORAGE_CAPACITY` |
| `storage.cos.ak` | `SIGMA_STORAGE_COS_AK` |
| `storage.cos.endpoint` | `SIGMA_STORAGE_COS_ENDPOINT` |
| `storage.cos.forcePathStyle` | `SIGMA_STORAGE_COS_FORCEPATHSTYLE` |
| `storage.cos.sk` | `SIGMA_STORAGE_COS_SK` |
| `storage.filesystem.path` | `SIGMA_STORAGE_FILESYSTEM_PATH` |
| `storage.oss.ak` | `SIGMA_STORAGE_OSS_AK` |
| `storage.oss.bucket` | `SIGMA_STORAGE_OSS_BUCKET` |
| `storage.oss.endpoint` | `SIGMA_STORAGE_OSS_ENDPOINT` |
| `storage.oss.forcePathStyle` | `SIGMA_STORAGE_OSS_FORCEPATHSTYLE` |
| `storage.oss.sk` | `SIGMA_STORAGE_OSS_SK` |
| `storage.redirect` | `SIGMA_STORAGE_REDIRECT` |
| `storage.rootDirectory` | `SIGMA_STORAGE_ROOTDIRECTORY` |
| `storage.s3.ak` | `SIGMA_STORAGE_S3_AK` |
| `storage.s3.bucket` | `SIGMA_STORAGE_S3_BUCKET` |
| `storage.s3.checksumValidation` | `SIGMA_STORAGE_S3_CHECKSUMVALIDATION` |
| `storage.s3.endpoint` | `SIGMA_STORAGE_S3_ENDPOINT` |
| `storage.s3.forcePathStyle` | `SIGMA_STORAGE_S3_FORCEPATHSTYLE` |
| `storage.s3.operateObjectWithoutMD5` | `SIGMA_STORAGE_S3_OPERATEOBJECTWITHOUTMD5` |
| `storage.s3.region` | `SIGMA_STORAGE_S3_REGION` |
| `storage.s3.sk` | `SIGMA_STORAGE_S3_SK` |
| `storage.type` | `SIGMA_STORAGE_TYPE` |

#### trace

| config path | environment variable |
| --- | --- |
| `trace.enabled` | `SIGMA_TRACE_ENABLED` |
| `trace.endpoint` | `SIGMA_TRACE_ENDPOINT` |
| `trace.insecure` | `SIGMA_TRACE_INSECURE` |
| `trace.protocol` | `SIGMA_TRACE_PROTOCOL` |
| `trace.sampleRatio` | `SIGMA_TRACE_SAMPLERATIO` |
| `trace.serviceName` | `SIGMA_TRACE_SERVICENAME` |

#### workqueue

| config path | environment variable |
| --- | --- |
| `workqueue.inmemory.concurrency` | `SIGMA_WORKQUEUE_INMEMORY_CONCURRENCY` |
| `workqueue.inmemory.maxBacklog` | `SIGMA_WORKQUEUE_INMEMORY_MAXBACKLOG` |
| `workqueue.redis.concurrency` | `SIGMA_WORKQUEUE_REDIS_CONCURRENCY` |
| `workqueue.redis.maxBacklog` | `SIGMA_WORKQUEUE_REDIS_MAXBACKLOG` |
| `workqueue.type` | `SIGMA_WORKQUEUE_TYPE` |

### 示例

```bash
# 构建 daemon
SIGMA_DAEMON_BUILDER_ENABLED=true
SIGMA_DAEMON_BUILDER_TYPE=docker
SIGMA_DAEMON_BUILDER_IMAGE=ghcr.io/go-sigma/sigma-builder:nightly-alpine
SIGMA_DAEMON_BUILDER_TLSVERIFY=true
SIGMA_DAEMON_BUILDER_EXTRA_HOSTS=host.docker.internal:host-gateway

# 数据库与存储
SIGMA_DATABASE_TYPE=postgresql
SIGMA_DATABASE_POSTGRESQL_PASSWORD=change-me
SIGMA_STORAGE_TYPE=s3
SIGMA_STORAGE_S3_BUCKET=sigma
SIGMA_STORAGE_S3_AK=sigma
SIGMA_STORAGE_S3_SK=change-me

# 列表与时长
SIGMA_REDIS_ADDRS=redis-0:6379,redis-1:6379
SIGMA_DAEMON_GC_RETENTION=48h
```

### 只能用配置文件设置的项

`scm.providers` 与 `namespace.initialize` 是对象列表，单个变量无法表达，只能写在
配置文件里；设置 `SIGMA_SCM_PROVIDERS` 或 `SIGMA_NAMESPACE_INITIALIZE` 会在日志中
输出一条警告并被忽略。

### 与配置无关的 `SIGMA_` 变量

`SIGMA_SERVER`、`SIGMA_USERNAME`、`SIGMA_PASSWORD`、`SIGMA_CREDENTIALS_FILE` 属于
`sigma` CLI（参见 [CLI](./cli.md)），不是服务端配置项，可以和上面的变量同时使用。
