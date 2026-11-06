---
title: Configuration
---

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
  # the default namespace visibility, available: public, private
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
  # total storage sigma may use, in bytes; used by the admin home dashboard to show
  # the usage ratio. 0 means unlimited.
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
    # Leave empty to generate a random password on the first start. The generated password is
    # written to the server log exactly once, for example:
    #   level=WARN msg="admin user created with an auto-generated password, please record it and
    #   change it after login" username=sigma password=62AWdMKiKb
    # Record that password, sign in with it, and change it right away: the login page asks for a
    # new password before it lets an account that still uses a generated password into the app.
    # Set a value here to pin the initial password instead (then no password is generated and no
    # log line is printed).
    password: ""
    email: sigma@sigma.io
  token:
    realm: ""
    service: ""
  jwt:
    ttl: 6h
    refreshTTL: 72h
    # base64 encoded Ed25519 private key, generate one with:
    # openssl genpkey -algorithm ED25519 | base64
    # leave it empty to generate a key on first start and store it in the
    # settings table; later starts reuse it. A configured key always wins.
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

## Environment variable overrides

Every item above can be overridden from the environment, so a container
deployment does not need to template the config file. Variables are resolved
after the YAML file is loaded: a set variable always wins over the file, and an
unset variable always keeps the value from the file.

### Naming rule

The variable name is `SIGMA_` followed by the YAML path, upper cased and joined
with `_`:

| config path | environment variable |
| --- | --- |
| `daemon.builder.enabled` | `SIGMA_DAEMON_BUILDER_ENABLED` |
| `daemon.builder.type` | `SIGMA_DAEMON_BUILDER_TYPE` |
| `daemon.builder.image` | `SIGMA_DAEMON_BUILDER_IMAGE` |
| `daemon.builder.tlsVerify` | `SIGMA_DAEMON_BUILDER_TLSVERIFY` |
| `daemon.builder.extra_hosts` | `SIGMA_DAEMON_BUILDER_EXTRA_HOSTS` |
| `auth.jwt.privateKey` | `SIGMA_AUTH_JWT_PRIVATEKEY` |

Names are case insensitive, and the underscores inside one path segment are
optional, so `SIGMA_DAEMON_BUILDER_TLS_VERIFY` and
`sigma_daemon_builder_tlsverify` address the same item as
`SIGMA_DAEMON_BUILDER_TLSVERIFY`.

### Value syntax

| type | accepted syntax | example |
| --- | --- | --- |
| string | taken verbatim | `SIGMA_HTTP_ENDPOINT=https://registry.example.com` |
| bool | `true`, `false`, `1`, `0` | `SIGMA_DAEMON_BUILDER_ENABLED=true` |
| integer | decimal | `SIGMA_DATABASE_MYSQL_PORT=3306` |
| duration | Go duration syntax | `SIGMA_DAEMON_GC_RETENTION=48h` |
| enum | one of the documented values, validated on startup | `SIGMA_DAEMON_BUILDER_TYPE=docker` |
| float | decimal with `.` as separator | `SIGMA_TRACE_SAMPLERATIO=0.25` |
| list | comma separated, whitespace trimmed | `SIGMA_REDIS_ADDRS=redis-0:6379,redis-1:6379` |

An unparsable value, an unknown enum value, or a boolean that is not
`true/false/1/0` aborts the startup with an error naming both the variable and
the config path, for example
`SIGMA_DAEMON_GC_BATCHSIZE (daemon.gc.batchSize): ... invalid syntax`. Valid
overrides in the same process are still applied before the error is returned.

### Complete mapping

Sections and items are sorted alphabetically; every leaf of the config file is
listed.

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

### Sample environment

```bash
# build daemons
SIGMA_DAEMON_BUILDER_ENABLED=true
SIGMA_DAEMON_BUILDER_TYPE=docker
SIGMA_DAEMON_BUILDER_IMAGE=ghcr.io/go-sigma/sigma-builder:nightly-alpine
SIGMA_DAEMON_BUILDER_TLSVERIFY=true
SIGMA_DAEMON_BUILDER_EXTRA_HOSTS=host.docker.internal:host-gateway

# storage and database
SIGMA_DATABASE_TYPE=postgresql
SIGMA_DATABASE_POSTGRESQL_PASSWORD=change-me
SIGMA_STORAGE_TYPE=s3
SIGMA_STORAGE_S3_BUCKET=sigma
SIGMA_STORAGE_S3_AK=sigma
SIGMA_STORAGE_S3_SK=change-me

# list and duration values
SIGMA_REDIS_ADDRS=redis-0:6379,redis-1:6379
SIGMA_DAEMON_GC_RETENTION=48h
```

### Items that must stay in the config file

`scm.providers` and `namespace.initialize` are lists of objects, which a single
variable cannot express. They are only read from the config file; setting
`SIGMA_SCM_PROVIDERS` or `SIGMA_NAMESPACE_INITIALIZE` logs a warning and is
ignored.

### Unrelated `SIGMA_` variables

`SIGMA_SERVER`, `SIGMA_USERNAME`, `SIGMA_PASSWORD` and `SIGMA_CREDENTIALS_FILE`
belong to the `sigma` CLI (see [CLI](./cli.md)), not to the server
configuration. They are independent of the variables above and can be set side
by side with them.
