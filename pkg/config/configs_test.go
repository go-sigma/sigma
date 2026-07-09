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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func errChecker(config Configuration) error {
	return fmt.Errorf("fake error")
}

func noErrChecker(config Configuration) error {
	return nil
}

func TestCheckMiddleware(t *testing.T) {
	checkers = make([]checker, 0, 2)
	checkers = append(checkers, noErrChecker, errChecker)
	err := CheckMiddleware()
	require.Error(t, err)
}

func TestConfigurationDaemonVulnerabilityWithDefaults(t *testing.T) {
	var vulnerability ConfigurationDaemonVulnerability
	vulnerability.WithDefaults()
	require.Equal(t, 6*time.Hour, vulnerability.StaleTimeout)

	vulnerability.StaleTimeout = 3 * time.Hour
	vulnerability.WithDefaults()
	require.Equal(t, 3*time.Hour, vulnerability.StaleTimeout)
}

func TestConfigurationCacheWithDefaults(t *testing.T) {
	var cache ConfigurationCache
	cache.WithDefaults()
	require.Equal(t, "sigma-cache", cache.Prefix)
	require.Equal(t, 72*time.Hour, cache.TTL)
	require.Equal(t, 5*time.Minute, cache.NegativeTTL)
	require.Equal(t, 72*time.Hour, cache.Redis.Ttl)
	require.Equal(t, 10240, cache.Inmemory.Size)
	require.Equal(t, 1000, cache.Prewarm.Limit)
	require.Equal(t, 10*time.Second, cache.Prewarm.Timeout)
	require.True(t, cache.Prewarm.IsEnabled())

	enabled := false
	cache = ConfigurationCache{
		Prewarm: ConfigurationCachePrewarm{Enabled: &enabled},
	}
	cache.WithDefaults()
	require.False(t, cache.Prewarm.IsEnabled())
}

func TestConfigurationHTTPBodyLimitWithDefaults(t *testing.T) {
	var httpConfig ConfigurationHTTP
	httpConfig.WithDefaults()
	require.True(t, httpConfig.BodyLimit.IsEnabled())
	require.Equal(t, int64(10<<20), httpConfig.BodyLimit.DefaultLimit())
	require.Equal(t, int64(100<<20), httpConfig.BodyLimit.ManifestLimit())
	require.Equal(t, int64(10<<30), httpConfig.BodyLimit.BlobLimit())
	require.Equal(t, 30*time.Second, httpConfig.Timeout.ReadHeaderTimeout())
	require.Equal(t, time.Hour, httpConfig.Timeout.ReadTimeout())
	require.Equal(t, time.Hour, httpConfig.Timeout.WriteTimeout())
	require.Equal(t, 2*time.Minute, httpConfig.Timeout.IdleTimeout())

	zero := int64(0)
	zeroTimeout := time.Duration(0)
	enabled := false
	httpConfig = ConfigurationHTTP{
		BodyLimit: ConfigurationHTTPBodyLimit{
			Enabled: &enabled,
			Blob:    &zero,
		},
		Timeout: ConfigurationHTTPTimeout{
			Read: &zeroTimeout,
		},
	}
	httpConfig.WithDefaults()
	require.False(t, httpConfig.BodyLimit.IsEnabled())
	require.Zero(t, httpConfig.BodyLimit.BlobLimit())
	require.Zero(t, httpConfig.Timeout.ReadTimeout())
}

func TestConfigurationHTTPTimeoutUnmarshal(t *testing.T) {
	var cfg Configuration
	err := yaml.Unmarshal([]byte(`
http:
  timeout:
    readHeader: 5s
    read: 0s
    write: 2m
    idle: 30s
`), &cfg)
	require.NoError(t, err)

	cfg.WithDefaults()
	require.Equal(t, 5*time.Second, cfg.HTTP.Timeout.ReadHeaderTimeout())
	require.Zero(t, cfg.HTTP.Timeout.ReadTimeout())
	require.Equal(t, 2*time.Minute, cfg.HTTP.Timeout.WriteTimeout())
	require.Equal(t, 30*time.Second, cfg.HTTP.Timeout.IdleTimeout())
}

func TestConfigurationMCPWithDefaults(t *testing.T) {
	var mcp ConfigurationMCP
	mcp.WithDefaults()
	require.False(t, mcp.Enabled)
	require.Equal(t, "/api/v1/mcp", mcp.Path)
	require.Equal(t, "streamable_http", mcp.Transport)
	require.True(t, mcp.AuditWritesEnabled())
	require.Equal(t, 30*time.Second, mcp.ToolTimeout)
	require.Equal(t, int64(1<<20), mcp.MaxRequestBody)

	disabled := false
	mcp = ConfigurationMCP{
		Enabled:        true,
		Path:           "/custom/mcp",
		Transport:      "sse",
		AuditWrites:    &disabled,
		ToolTimeout:    time.Minute,
		MaxRequestBody: 2 << 20,
	}
	mcp.WithDefaults()
	require.True(t, mcp.Enabled)
	require.Equal(t, "/custom/mcp", mcp.Path)
	require.Equal(t, "sse", mcp.Transport)
	require.False(t, mcp.AuditWritesEnabled())
	require.Equal(t, time.Minute, mcp.ToolTimeout)
	require.Equal(t, int64(2<<20), mcp.MaxRequestBody)
}

func TestConfigurationApplyEnvOverrides(t *testing.T) {
	t.Setenv("SIGMA_DATABASE_MYSQL_PASSWORD", "mysql-env")
	t.Setenv("SIGMA_DATABASE_POSTGRESQL_PASSWORD", "postgresql-env")
	t.Setenv("SIGMA_REDIS_URL", "redis://:redis-env@example.com:6379/0")
	t.Setenv("SIGMA_REDIS_USERNAME", "redis-user-env")
	t.Setenv("SIGMA_REDIS_PASSWORD", "redis-password-env")
	t.Setenv("SIGMA_STORAGE_S3_AK", "s3-ak-env")
	t.Setenv("SIGMA_STORAGE_S3_SK", "s3-sk-env")
	t.Setenv("SIGMA_STORAGE_COS_AK", "cos-ak-env")
	t.Setenv("SIGMA_STORAGE_COS_SK", "cos-sk-env")
	t.Setenv("SIGMA_STORAGE_OSS_AK", "oss-ak-env")
	t.Setenv("SIGMA_STORAGE_OSS_SK", "oss-sk-env")

	cfg := Configuration{}
	cfg.Database.Mysql.Password = "mysql-yaml"
	cfg.Database.Postgresql.Password = "postgresql-yaml"
	cfg.Redis.URL = "redis://:redis-yaml@example.com:6379/0"
	cfg.Redis.Username = "redis-user-yaml"
	cfg.Redis.Password = "redis-password-yaml"
	cfg.Storage.S3.Ak = "s3-ak-yaml"
	cfg.Storage.S3.Sk = "s3-sk-yaml"
	cfg.Storage.Cos.Ak = "cos-ak-yaml"
	cfg.Storage.Cos.Sk = "cos-sk-yaml"
	cfg.Storage.Oss.Ak = "oss-ak-yaml"
	cfg.Storage.Oss.Sk = "oss-sk-yaml"

	err := cfg.ApplyEnvOverrides()
	require.NoError(t, err)
	require.Equal(t, "mysql-env", cfg.Database.Mysql.Password)
	require.Equal(t, "postgresql-env", cfg.Database.Postgresql.Password)
	require.Equal(t, "redis://:redis-env@example.com:6379/0", cfg.Redis.URL)
	require.Equal(t, "redis-user-env", cfg.Redis.Username)
	require.Equal(t, "redis-password-env", cfg.Redis.Password)
	require.Equal(t, "s3-ak-env", cfg.Storage.S3.Ak)
	require.Equal(t, "s3-sk-env", cfg.Storage.S3.Sk)
	require.Equal(t, "cos-ak-env", cfg.Storage.Cos.Ak)
	require.Equal(t, "cos-sk-env", cfg.Storage.Cos.Sk)
	require.Equal(t, "oss-ak-env", cfg.Storage.Oss.Ak)
	require.Equal(t, "oss-sk-env", cfg.Storage.Oss.Sk)
}

func TestConfigurationApplyEnvOverridesKeepsUnsetValues(t *testing.T) {
	cfg := Configuration{}
	cfg.Database.Mysql.Password = "mysql-yaml"
	cfg.Database.Postgresql.Password = "postgresql-yaml"
	cfg.Redis.URL = "redis://:redis-yaml@example.com:6379/0"
	cfg.Redis.Username = "redis-user-yaml"
	cfg.Redis.Password = "redis-password-yaml"
	cfg.Storage.S3.Ak = "s3-ak-yaml"
	cfg.Storage.S3.Sk = "s3-sk-yaml"
	cfg.Storage.Cos.Ak = "cos-ak-yaml"
	cfg.Storage.Cos.Sk = "cos-sk-yaml"
	cfg.Storage.Oss.Ak = "oss-ak-yaml"
	cfg.Storage.Oss.Sk = "oss-sk-yaml"

	err := cfg.ApplyEnvOverrides()
	require.NoError(t, err)
	require.Equal(t, "mysql-yaml", cfg.Database.Mysql.Password)
	require.Equal(t, "postgresql-yaml", cfg.Database.Postgresql.Password)
	require.Equal(t, "redis://:redis-yaml@example.com:6379/0", cfg.Redis.URL)
	require.Equal(t, "redis-user-yaml", cfg.Redis.Username)
	require.Equal(t, "redis-password-yaml", cfg.Redis.Password)
	require.Equal(t, "s3-ak-yaml", cfg.Storage.S3.Ak)
	require.Equal(t, "s3-sk-yaml", cfg.Storage.S3.Sk)
	require.Equal(t, "cos-ak-yaml", cfg.Storage.Cos.Ak)
	require.Equal(t, "cos-sk-yaml", cfg.Storage.Cos.Sk)
	require.Equal(t, "oss-ak-yaml", cfg.Storage.Oss.Ak)
	require.Equal(t, "oss-sk-yaml", cfg.Storage.Oss.Sk)
}
