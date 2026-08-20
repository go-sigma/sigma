// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"strings"
	"sync"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

const (
	defaultHTTPBodyLimitDefault  int64 = 10 << 20  // 10Mi
	defaultHTTPBodyLimitManifest int64 = 100 << 20 // 100Mi
	defaultHTTPBodyLimitBlob     int64 = 10 << 30  // 10Gi

	defaultHTTPTimeoutReadHeader = 30 * time.Second
	defaultHTTPTimeoutRead       = time.Hour
	defaultHTTPTimeoutWrite      = time.Hour
	defaultHTTPTimeoutIdle       = 2 * time.Minute

	defaultMCPPath           = "/api/v1/mcp"
	defaultMCPTransport      = "streamable_http"
	defaultMCPToolTimeout    = 30 * time.Second
	defaultMCPMaxRequestBody = int64(1 << 20) // 1Mi
)

var configuration = &Configuration{}

var configOnce sync.Once

// GetConfig returns the singleton configuration after applying defaults.
func GetConfig() *Configuration {
	configOnce.Do(func() {
		configuration.WithDefaults()
	})
	return configuration
}

// Configuration ...
type Configuration struct {
	Log       ConfigurationLog       `yaml:"log"`
	Database  ConfigurationDatabase  `yaml:"database" mapstructure:"DATABASE"`
	Deploy    enums.Deploy           `yaml:"deploy"`
	Redis     ConfigurationRedis     `yaml:"redis"`
	Cache     ConfigurationCache     `yaml:"cache"`
	WorkQueue ConfigurationWorkQueue `yaml:"workqueue"`
	Locker    ConfigurationLocker    `yaml:"locker"`
	Namespace ConfigurationNamespace `yaml:"namespace"`
	HTTP      ConfigurationHTTP      `yaml:"http"`
	Storage   ConfigurationStorage   `yaml:"storage"`
	Proxy     ConfigurationProxy     `yaml:"proxy"`
	Daemon    ConfigurationDaemon    `yaml:"daemon"`
	Auth      ConfigurationAuth      `yaml:"auth"`
	Audit     ConfigurationAudit     `yaml:"audit"`
	MCP       ConfigurationMCP       `yaml:"mcp"`
	Trace     ConfigurationTrace     `yaml:"trace"`
	Analytics ConfigurationAnalytics `yaml:"analytics"`
}

// WithDefaults applies defaults to all top-level configuration sections.
func (c *Configuration) WithDefaults() {
	c.HTTP.WithDefaults()
	c.Auth.Jwt.WithDefaults()
	c.Namespace.WithDefaults()
	c.Daemon.WithDefaults()
	c.Audit.Cleanup.WithDefaults()
	c.MCP.WithDefaults()
	c.WorkQueue.Inmemory.WithDefaults()
	c.Cache.WithDefaults()
	c.Locker.WithDefaults()
	c.Database.WithDefaults()
}

// ConfigurationMCP configures the embedded MCP server.
type ConfigurationMCP struct {
	Enabled        bool          `yaml:"enabled"`
	Path           string        `yaml:"path"`
	Transport      string        `yaml:"transport"`
	AuditWrites    *bool         `yaml:"auditWrites"`
	ToolTimeout    time.Duration `yaml:"toolTimeout"`
	MaxRequestBody int64         `yaml:"maxRequestBody"`
}

// WithDefaults applies default values for MCP.
func (c *ConfigurationMCP) WithDefaults() {
	if c.Path == "" {
		c.Path = defaultMCPPath
	}
	if c.Transport == "" {
		c.Transport = defaultMCPTransport
	}
	if c.AuditWrites == nil {
		enabled := true
		c.AuditWrites = &enabled
	}
	if c.ToolTimeout == 0 {
		c.ToolTimeout = defaultMCPToolTimeout
	}
	if c.MaxRequestBody == 0 {
		c.MaxRequestBody = defaultMCPMaxRequestBody
	}
}

// AuditWritesEnabled reports whether write MCP tool calls should be audited.
func (c ConfigurationMCP) AuditWritesEnabled() bool {
	return c.AuditWrites == nil || *c.AuditWrites
}

// ConfigurationAudit configures HTTP audit recording.
type ConfigurationAudit struct {
	Enabled       *bool                     `yaml:"enabled"`
	RecordListGet bool                      `yaml:"recordListGet"`
	Cleanup       ConfigurationAuditCleanup `yaml:"cleanup"`
}

// ConfigurationAuditCleanup configures audit retention cleanup.
type ConfigurationAuditCleanup struct {
	SoftDeleteAfterMonths int           `yaml:"softDeleteAfterMonths"`
	HardDeleteAfterMonths int           `yaml:"hardDeleteAfterMonths"`
	Interval              time.Duration `yaml:"interval"`
	BatchSize             int           `yaml:"batchSize"`
}

// WithDefaults applies default values for audit cleanup.
func (c *ConfigurationAuditCleanup) WithDefaults() {
	if c.SoftDeleteAfterMonths == 0 {
		c.SoftDeleteAfterMonths = 3
	}
	if c.HardDeleteAfterMonths == 0 {
		c.HardDeleteAfterMonths = 6
	}
	if c.Interval == 0 {
		c.Interval = 7 * 24 * time.Hour
	}
	if c.BatchSize == 0 {
		c.BatchSize = 1000
	}
}

// IsEnabled reports whether HTTP audit recording is enabled.
func (c ConfigurationAudit) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// ConfigurationAnalytics configures backend analytics aggregation.
type ConfigurationAnalytics struct {
	Enabled        bool   `yaml:"enabled"`
	Backend        string `yaml:"backend"`
	CounterBackend string `yaml:"counterBackend"`
	RetentionDays  int    `yaml:"retentionDays"`
	FlushInterval  string `yaml:"flushInterval"`
	ReconcileCron  string `yaml:"reconcileCron"`
}

// ConfigurationTrace configures OpenTelemetry tracing.
type ConfigurationTrace struct {
	Enabled     bool    `yaml:"enabled"`
	Endpoint    string  `yaml:"endpoint"`
	Protocol    string  `yaml:"protocol"`
	Insecure    bool    `yaml:"insecure"`
	SampleRatio float64 `yaml:"sampleRatio"`
	ServiceName string  `yaml:"serviceName"`
}

type ConfigurationBuilderK8s struct {
	Kubeconfig string `yaml:"kubeconfig"`
	Namespace  string `yaml:"namespace"`
}

type ConfigurationBuilderDocker struct {
}

// ConfigurationLog ...
type ConfigurationLog struct {
	Level      enums.LogLevel `yaml:"level"`
	ProxyLevel enums.LogLevel `yaml:"proxyLevel"`
}

// ConfigurationDatabaseSqlite3 ...
type ConfigurationDatabaseSqlite3 struct {
	Path string `yaml:"path"`
}

// ConfigurationDatabaseTurso ...
type ConfigurationDatabaseTurso struct {
	DSN string `yaml:"dsn"`
}

// ConfigurationDatabaseMysql ...
type ConfigurationDatabaseMysql struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// ConfigurationDatabase ...
type ConfigurationDatabasePostgresql struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SslMode  string `yaml:"sslMode"`
}

// ConfigurationDatabase ...
type ConfigurationDatabase struct {
	Type            enums.Database                  `yaml:"type" mapstructure:"TYPE"`
	MaxOpenConns    int                             `yaml:"maxOpenConns"`
	MaxIdleConns    int                             `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration                   `yaml:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration                   `yaml:"connMaxIdleTime"`
	Sqlite3         ConfigurationDatabaseSqlite3    `yaml:"sqlite3"`
	Turso           ConfigurationDatabaseTurso      `yaml:"turso"`
	Mysql           ConfigurationDatabaseMysql      `yaml:"mysql"`
	Postgresql      ConfigurationDatabasePostgresql `yaml:"postgresql"`
}

// WithDefaults applies default values for database connection pool.
func (c *ConfigurationDatabase) WithDefaults() {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 50
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 10 * time.Minute
	}
}

// ConfigurationRedis ...
//
// 启用条件：URL 或 Addrs 任一非空，即视为启用 Redis；两者皆空则视为未启用。
type ConfigurationRedis struct {
	URL          string   `yaml:"url"`          // URL 形式的连接字符串；非空时优先使用，覆盖以下字段
	Addrs        []string `yaml:"addrs"`        // Redis 节点地址列表（host:port）。单节点写一个；集群写多个；哨兵写所有 Sentinel 地址
	Username     string   `yaml:"username"`     // Redis 6+ ACL 用户名
	Password     string   `yaml:"password"`     // 密码
	DB           int      `yaml:"db"`           // 单节点模式下使用的数据库编号
	MasterName   string   `yaml:"masterName"`   // 哨兵模式下的 master 名；非空时 Addrs 视为 Sentinel 列表，走 failover 客户端
	PoolSize     int      `yaml:"poolSize"`     // 连接池大小；0 表示使用 go-redis 默认值（每 CPU 10 个）
	MinIdleConns int      `yaml:"minIdleConns"` // 最小空闲连接数
	MaxRetries   int      `yaml:"maxRetries"`   // 命令重试次数
}

// Enabled reports whether Redis is configured. Returns true when either URL
// or Addrs is non-empty (or MasterName is set for sentinel mode).
func (r ConfigurationRedis) Enabled() bool {
	return r.URL != "" || len(r.Addrs) > 0 || r.MasterName != ""
}

// ConfigurationCacheRedis ...
type ConfigurationCacheRedis struct {
	Ttl time.Duration `yaml:"ttl"`
}

// ConfigurationCachePrewarm configures metadata cache prewarming.
type ConfigurationCachePrewarm struct {
	Enabled *bool         `yaml:"enabled"`
	Limit   int           `yaml:"limit"`
	Timeout time.Duration `yaml:"timeout"`
}

// IsEnabled reports whether metadata cache prewarming is enabled.
func (c ConfigurationCachePrewarm) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

// ConfigurationCacheInmemory ...
type ConfigurationCacheInmemory struct {
	Size int `yaml:"size"`
}

// ConfigurationCache ...
type ConfigurationCache struct {
	Type        enums.CacherType           `yaml:"type"`
	Prefix      string                     `yaml:"prefix"`
	TTL         time.Duration              `yaml:"ttl"`
	NegativeTTL time.Duration              `yaml:"negativeTtl"`
	Redis       ConfigurationCacheRedis    `yaml:"redis"`
	Inmemory    ConfigurationCacheInmemory `yaml:"inmemory"`
	Prewarm     ConfigurationCachePrewarm  `yaml:"prewarm"`
}

// WithDefaults applies default values for cache.
func (c *ConfigurationCache) WithDefaults() {
	if c.Inmemory.Size == 0 {
		c.Inmemory.Size = 10240
	}
	if strings.TrimSpace(c.Prefix) == "" {
		c.Prefix = "sigma-cache"
	}
	if c.TTL == 0 {
		c.TTL = 72 * time.Hour
	}
	if c.NegativeTTL == 0 {
		c.NegativeTTL = 5 * time.Minute
	}
	if c.Redis.Ttl == 0 {
		c.Redis.Ttl = c.TTL
	}
	if c.Prewarm.Limit == 0 {
		c.Prewarm.Limit = 1000
	}
	if c.Prewarm.Timeout == 0 {
		c.Prewarm.Timeout = 10 * time.Second
	}
}

type ConfigurationWorkQueueRedis struct {
	Concurrency int `yaml:"concurrency"`
	MaxBacklog  int `yaml:"maxBacklog"`
}

type ConfigurationWorkQueueDatabase struct {
}

type ConfigurationWorkQueueKafka struct {
}

type ConfigurationWorkQueueInmemmory struct {
	Concurrency int `yaml:"concurrency"`
	MaxBacklog  int `yaml:"maxBacklog"`
}

// WithDefaults applies default values for inmemory workqueue.
func (c *ConfigurationWorkQueueInmemmory) WithDefaults() {
	if c.Concurrency == 0 {
		c.Concurrency = 1024
	}
	if c.MaxBacklog == 0 {
		c.MaxBacklog = 1000
	}
}

// ConfigurationWorkQueue ...
type ConfigurationWorkQueue struct {
	Type     enums.WorkQueueType             `yaml:"type"`
	Redis    ConfigurationWorkQueueRedis     `yaml:"redis"`
	Database ConfigurationWorkQueueDatabase  `yaml:"database"`
	Kafka    ConfigurationWorkQueueKafka     `yaml:"kafka"`
	Inmemory ConfigurationWorkQueueInmemmory `yaml:"inmemory"`
}

// ConfigurationLockerRedis ...
type ConfigurationLockerRedis struct{}

// ConfigurationLocker ...
type ConfigurationLocker struct {
	Type   enums.LockerType         `yaml:"type"`
	Redis  ConfigurationLockerRedis `yaml:"redis"`
	Prefix string                   `yaml:"prefix"`
}

// WithDefaults applies default values for locker.
func (c *ConfigurationLocker) WithDefaults() {
	if strings.TrimSpace(c.Prefix) == "" {
		c.Prefix = "sigma-locker"
	}
}

// ConfigurationNamespace ...
type ConfigurationNamespace struct {
	AutoCreate bool             `yaml:"autoCreate"`
	Visibility enums.Visibility `yaml:"visibility"`
}

// WithDefaults applies default values for namespace.
func (c *ConfigurationNamespace) WithDefaults() {
	if c.Visibility.String() == "" {
		c.Visibility = enums.VisibilityPrivate
	}
}

// ConfigurationHttpTLS ...
type ConfigurationHttpTLS struct {
	Enabled     bool   `yaml:"enabled"`
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

// ConfigurationHTTPBodyLimit configures request body size limits in bytes.
type ConfigurationHTTPBodyLimit struct {
	Enabled  *bool  `yaml:"enabled"`
	Default  *int64 `yaml:"default"`
	Manifest *int64 `yaml:"manifest"`
	Blob     *int64 `yaml:"blob"`
}

// IsEnabled reports whether request body size limiting is enabled.
func (c ConfigurationHTTPBodyLimit) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

// DefaultLimit returns the default request body size limit in bytes.
func (c ConfigurationHTTPBodyLimit) DefaultLimit() int64 {
	return valueOrZero(c.Default)
}

// ManifestLimit returns the distribution manifest request body size limit in bytes.
func (c ConfigurationHTTPBodyLimit) ManifestLimit() int64 {
	return valueOrZero(c.Manifest)
}

// BlobLimit returns the distribution blob upload request body size limit in bytes.
func (c ConfigurationHTTPBodyLimit) BlobLimit() int64 {
	return valueOrZero(c.Blob)
}

// WithDefaults applies default values for HTTP body size limits.
func (c *ConfigurationHTTPBodyLimit) WithDefaults() {
	if c.Default == nil {
		limit := defaultHTTPBodyLimitDefault
		c.Default = &limit
	}
	if c.Manifest == nil {
		limit := defaultHTTPBodyLimitManifest
		c.Manifest = &limit
	}
	if c.Blob == nil {
		limit := defaultHTTPBodyLimitBlob
		c.Blob = &limit
	}
}

// ConfigurationHTTPTimeout configures HTTP server timeouts.
type ConfigurationHTTPTimeout struct {
	ReadHeader *time.Duration `yaml:"readHeader"`
	Read       *time.Duration `yaml:"read"`
	Write      *time.Duration `yaml:"write"`
	Idle       *time.Duration `yaml:"idle"`
}

// ReadHeaderTimeout returns the maximum duration for reading request headers.
func (c ConfigurationHTTPTimeout) ReadHeaderTimeout() time.Duration {
	return durationOrZero(c.ReadHeader)
}

// ReadTimeout returns the maximum duration for reading the full request.
func (c ConfigurationHTTPTimeout) ReadTimeout() time.Duration {
	return durationOrZero(c.Read)
}

// WriteTimeout returns the maximum duration before timing out response writes.
func (c ConfigurationHTTPTimeout) WriteTimeout() time.Duration {
	return durationOrZero(c.Write)
}

// IdleTimeout returns the maximum duration to wait for the next request.
func (c ConfigurationHTTPTimeout) IdleTimeout() time.Duration {
	return durationOrZero(c.Idle)
}

// WithDefaults applies default values for HTTP server timeouts.
func (c *ConfigurationHTTPTimeout) WithDefaults() {
	if c.ReadHeader == nil {
		timeout := defaultHTTPTimeoutReadHeader
		c.ReadHeader = &timeout
	}
	if c.Read == nil {
		timeout := defaultHTTPTimeoutRead
		c.Read = &timeout
	}
	if c.Write == nil {
		timeout := defaultHTTPTimeoutWrite
		c.Write = &timeout
	}
	if c.Idle == nil {
		timeout := defaultHTTPTimeoutIdle
		c.Idle = &timeout
	}
}

// ConfigurationHTTP ...
type ConfigurationHTTP struct {
	Endpoint                     string                     `yaml:"endpoint"`
	InternalEndpoint             string                     `yaml:"internalEndpoint"`
	InternalDistributionEndpoint string                     `yaml:"internalDistributionEndpoint"`
	TLS                          ConfigurationHttpTLS       `yaml:"tls"`
	BodyLimit                    ConfigurationHTTPBodyLimit `yaml:"bodyLimit"`
	Timeout                      ConfigurationHTTPTimeout   `yaml:"timeout"`
}

// WithDefaults applies default values for HTTP.
func (c *ConfigurationHTTP) WithDefaults() {
	if c.Endpoint == "" {
		c.Endpoint = "http://127.0.0.1:3000"
	}
	if c.InternalEndpoint == "" {
		c.InternalEndpoint = "http://127.0.0.1:3000"
	}
	c.BodyLimit.WithDefaults()
	c.Timeout.WithDefaults()
}

func valueOrZero(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func durationOrZero(v *time.Duration) time.Duration {
	if v == nil {
		return 0
	}
	return *v
}

// ConfigurationStorageFilesystem ...
type ConfigurationStorageFilesystem struct {
	Path string `yaml:"path"`
}

// ConfigurationStorageS3 ...
type ConfigurationStorageS3 struct {
	Ak                      string `yaml:"ak"`
	Sk                      string `yaml:"sk"`
	Endpoint                string `yaml:"endpoint"`
	Region                  string `yaml:"region"`
	Bucket                  string `yaml:"bucket"`
	ForcePathStyle          bool   `yaml:"forcePathStyle"`
	OperateObjectWithoutMD5 bool   `yaml:"operateObjectWithoutMD5"`
	ChecksumValidation      string `yaml:"checksumValidation"`
}

// ConfigurationStorageCos ...
type ConfigurationStorageCos struct {
	Ak             string `yaml:"ak"`
	Sk             string `yaml:"sk"`
	Endpoint       string `yaml:"endpoint"`
	ForcePathStyle bool   `yaml:"forcePathStyle"`
}

// ConfigurationStorageQiniu ...
type ConfigurationStorageQiniu struct {
	Ak       string `yaml:"ak"`
	Sk       string `yaml:"sk"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
	UseHTTPS bool   `yaml:"useHttps"`
}

// ConfigurationStorageOss ...
type ConfigurationStorageOss struct {
	Ak             string `yaml:"ak"`
	Sk             string `yaml:"sk"`
	Bucket         string `yaml:"bucket"`
	Endpoint       string `yaml:"endpoint"`
	ForcePathStyle bool   `yaml:"forcePathStyle"`
}

// ConfigurationStorage ...
type ConfigurationStorage struct {
	RootDirectory string                         `yaml:"rootDirectory"`
	Redirect      bool                           `yaml:"redirect"`
	Type          enums.StorageType              `yaml:"type"`
	Filesystem    ConfigurationStorageFilesystem `yaml:"filesystem"`
	S3            ConfigurationStorageS3         `yaml:"s3"`
	Cos           ConfigurationStorageCos        `yaml:"cos"`
	Oss           ConfigurationStorageOss        `yaml:"oss"`
}

// ConfigurationProxy ...
type ConfigurationProxy struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	TlsVerify bool   `yaml:"tlsVerify"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Token     string `yaml:"token"`
}

// ConfigurationDaemonGc ...
type ConfigurationDaemonGc struct {
	Retention       time.Duration `yaml:"retention"`
	Cron            string        `yaml:"cron"`
	BatchSize       int           `yaml:"batchSize"`
	WorkerCount     int           `yaml:"workerCount"`
	RecordBatchSize int           `yaml:"recordBatchSize"`
	LockExpire      time.Duration `yaml:"lockExpire"`
	LockWaitTimeout time.Duration `yaml:"lockWaitTimeout"`
}

// WithDefaults applies default values for GC daemon.
func (c *ConfigurationDaemonGc) WithDefaults() {
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	if c.WorkerCount == 0 {
		c.WorkerCount = 2
	}
	if c.RecordBatchSize == 0 {
		c.RecordBatchSize = 100
	}
	if c.LockExpire == 0 {
		c.LockExpire = 30 * time.Second
	}
	if c.LockWaitTimeout == 0 {
		c.LockWaitTimeout = 5 * time.Second
	}
}

// ConfigurationDaemonDocker ...
type ConfigurationDaemonDocker struct {
	Sock    *string `yaml:"sock"`
	Network string  `yaml:"network"`
}

// ConfigurationDaemonKubernetes ...
type ConfigurationDaemonKubernetes struct {
	Kubeconfig *string `yaml:"kubeconfig"`
	Namespace  string  `yaml:"namespace"`
}

// WithDefaults applies default values for kubernetes daemon config.
func (c *ConfigurationDaemonKubernetes) WithDefaults() {
	if c.Namespace == "" {
		c.Namespace = "default"
	}
}

// ConfigurationDaemonPodman ...
type ConfigurationDaemonPodman struct {
	URI string `yaml:"uri"`
}

// WithDefaults applies default values for podman daemon.
func (c *ConfigurationDaemonPodman) WithDefaults() {
	if c.URI == "" {
		c.URI = "unix:///run/podman/podman.sock"
	}
}

// ConfigurationDaemonBuilder ...
type ConfigurationDaemonBuilder struct {
	Enabled    bool                          `yaml:"enabled"`
	Type       enums.BuilderType             `yaml:"type"`
	Image      string                        `yaml:"image"`
	TlsVerify  *bool                         `yaml:"tlsVerify"`
	Docker     ConfigurationDaemonDocker     `yaml:"docker"`
	Kubernetes ConfigurationDaemonKubernetes `yaml:"kubernetes"`
	Podman     ConfigurationDaemonPodman     `yaml:"podman"`
	ExtraHosts []string                      `yaml:"extra_hosts"` // 构建容器的额外 hosts，格式: "hostname:ip"
}

// TLSVerifyEnabled reports whether builder cache API TLS certificates should be verified.
func (c ConfigurationDaemonBuilder) TLSVerifyEnabled() bool {
	return c.TlsVerify == nil || *c.TlsVerify
}

// ConfigurationDaemonVulnerabilityGrype configures grype vulnerability scanning.
type ConfigurationDaemonVulnerabilityGrype struct {
	DBCacheDir string `yaml:"dbCacheDir"`
}

// ConfigurationDaemonVulnerability configures vulnerability scanning.
type ConfigurationDaemonVulnerability struct {
	StaleTimeout time.Duration                         `yaml:"staleTimeout"`
	Grype        ConfigurationDaemonVulnerabilityGrype `yaml:"grype"`
}

// WithDefaults applies default values for vulnerability scanning.
func (c *ConfigurationDaemonVulnerability) WithDefaults() {
	if c.StaleTimeout <= 0 {
		c.StaleTimeout = 6 * time.Hour
	}
}

// ConfigurationDaemonSizeReconcile configures size reconciliation for namespaces and repositories.
type ConfigurationDaemonSizeReconcile struct {
	Interval        time.Duration `yaml:"interval"`
	BatchSize       int           `yaml:"batchSize"`
	LockExpire      time.Duration `yaml:"lockExpire"`
	LockWaitTimeout time.Duration `yaml:"lockWaitTimeout"`
}

// WithDefaults applies default values for size reconcile daemon.
func (c *ConfigurationDaemonSizeReconcile) WithDefaults() {
	if c.Interval == 0 {
		c.Interval = 6 * time.Hour
	}
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	if c.LockExpire == 0 {
		c.LockExpire = 30 * time.Second
	}
	if c.LockWaitTimeout == 0 {
		c.LockWaitTimeout = 100 * time.Millisecond
	}
}

// ConfigurationDaemon ...
type ConfigurationDaemon struct {
	Builder       ConfigurationDaemonBuilder       `yaml:"builder"`
	GC            ConfigurationDaemonGc            `yaml:"gc"`
	Vulnerability ConfigurationDaemonVulnerability `yaml:"vulnerability"`
	SizeReconcile ConfigurationDaemonSizeReconcile `yaml:"sizeReconcile"`
}

// WithDefaults applies default values for daemon.
func (c *ConfigurationDaemon) WithDefaults() {
	c.Builder.Kubernetes.WithDefaults()
	c.Builder.Podman.WithDefaults()
	c.GC.WithDefaults()
	c.Vulnerability.WithDefaults()
	c.SizeReconcile.WithDefaults()
}

// ConfigurationAuthInternalUser ...
type ConfigurationAuthInternalUser struct {
	Username string `yaml:"username"`
}

// ConfigurationAuthAdmin ...
type ConfigurationAuthAdmin struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

// ConfigurationAuthToken ...
type ConfigurationAuthToken struct {
	Realm   string `yaml:"realm"`
	Service string `yaml:"service"`
}

// ConfigurationAuthJwt ...
type ConfigurationAuthJwt struct {
	Ttl        time.Duration `yaml:"ttl"`
	RefreshTTL time.Duration `yaml:"refreshTTL"`
	PrivateKey string        `yaml:"privateKey"`
}

// WithDefaults applies default values for JWT auth.
func (c *ConfigurationAuthJwt) WithDefaults() {
	if c.Ttl == 0 {
		c.Ttl = time.Hour
	}
	if c.RefreshTTL == 0 {
		c.RefreshTTL = time.Hour * 24
	}
}

// ConfigurationAuthOauth2Github ...
type ConfigurationAuthOauth2Github struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

// ConfigurationAuthOauth2Gitlab ...
type ConfigurationAuthOauth2Gitlab struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

// ConfigurationAuthOauth2Gitea ...
type ConfigurationAuthOauth2Gitea struct {
	Enabled      bool   `yaml:"enabled"`
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

// ConfigurationAuthOauth2 ...
type ConfigurationAuthOauth2 struct {
	Github ConfigurationAuthOauth2Github `yaml:"github"`
	Gitlab ConfigurationAuthOauth2Gitlab `yaml:"gitlab"`
	Gitea  ConfigurationAuthOauth2Gitea  `yaml:"gitea"`
}

// ConfigurationAuthAnonymous ...
type ConfigurationAuthAnonymous struct {
	Enabled bool `yaml:"enabled"`
}

// ConfigurationAuth ...
type ConfigurationAuth struct {
	Anonymous ConfigurationAuthAnonymous `yaml:"anonymous"`
	Admin     ConfigurationAuthAdmin     `yaml:"admin"`
	Token     ConfigurationAuthToken     `yaml:"token"`
	Oauth2    ConfigurationAuthOauth2    `yaml:"oauth2"`
	Jwt       ConfigurationAuthJwt       `yaml:"jwt"`
}
