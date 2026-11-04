// Copyright 2026 sigma
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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestEnvCatalogNames(t *testing.T) {
	catalog := catalogOnce()

	paths := make(map[string]string, len(catalog.supported))
	for _, field := range catalog.supported {
		require.Equal(t, envPrefix, field.name[:len(envPrefix)])
		require.NotEmpty(t, field.path)
		require.NotContains(t, paths, field.name, "duplicated env name")
		paths[field.name] = field.path
	}

	// Spot check every naming shape: enums, lists, pointers, durations, camelCase
	// segments and snake_case segments.
	expected := map[string]string{
		"SIGMA_DAEMON_BUILDER_ENABLED":        "daemon.builder.enabled",
		"SIGMA_DAEMON_BUILDER_TYPE":           "daemon.builder.type",
		"SIGMA_DAEMON_BUILDER_IMAGE":          "daemon.builder.image",
		"SIGMA_DAEMON_BUILDER_TLSVERIFY":      "daemon.builder.tlsVerify",
		"SIGMA_DAEMON_BUILDER_EXTRA_HOSTS":    "daemon.builder.extra_hosts",
		"SIGMA_DAEMON_SIZERECONCILE_INTERVAL": "daemon.sizeReconcile.interval",
		"SIGMA_AUTH_LOGINRATELIMIT_WINDOW":    "auth.loginRateLimit.window",
		"SIGMA_STORAGE_S3_FORCEPATHSTYLE":     "storage.s3.forcePathStyle",
		"SIGMA_HTTP_TIMEOUT_READHEADER":       "http.timeout.readHeader",
		"SIGMA_TRACE_SAMPLERATIO":             "trace.sampleRatio",
		"SIGMA_LOG_LEVEL":                     "log.level",
	}
	for name, path := range expected {
		require.Equal(t, path, paths[name], "unexpected path for %s", name)
	}

	// The variables that existed before the generic override must keep working.
	for _, name := range []string{
		"SIGMA_DATABASE_MYSQL_PASSWORD",
		"SIGMA_DATABASE_POSTGRESQL_PASSWORD",
		"SIGMA_REDIS_URL",
		"SIGMA_REDIS_USERNAME",
		"SIGMA_REDIS_PASSWORD",
		"SIGMA_STORAGE_S3_AK",
		"SIGMA_STORAGE_S3_SK",
		"SIGMA_STORAGE_COS_AK",
		"SIGMA_STORAGE_COS_SK",
		"SIGMA_STORAGE_OSS_AK",
		"SIGMA_STORAGE_OSS_SK",
	} {
		require.Contains(t, paths, name)
	}

	// Structured values cannot be expressed by one variable.
	skipped := make([]string, 0, len(catalog.skipped))
	for _, field := range catalog.skipped {
		skipped = append(skipped, field.path)
	}
	require.ElementsMatch(t, []string{"namespace.initialize", "scm.providers"}, skipped)
}

func TestApplyEnvOverridesEnablesEverySection(t *testing.T) {
	var cfg Configuration
	err := cfg.applyEnvOverrides([]string{
		"SIGMA_LOG_LEVEL=trace",
		"SIGMA_LOG_PROXYLEVEL=warn",
		"SIGMA_DEPLOY=replica",
		"SIGMA_DATABASE_TYPE=postgresql",
		"SIGMA_DATABASE_MAXOPENCONNS=42",
		"SIGMA_DATABASE_CONNMAXLIFETIME=3m",
		"SIGMA_DATABASE_SQLITE3_PATH=/data/sigma.db",
		"SIGMA_REDIS_ADDRS=redis-0:6379, redis-1:6379",
		"SIGMA_REDIS_DB=3",
		"SIGMA_CACHE_TYPE=redis",
		"SIGMA_CACHE_TTL=24h",
		"SIGMA_CACHE_PREWARM_ENABLED=false",
		"SIGMA_WORKQUEUE_TYPE=database",
		"SIGMA_WORKQUEUE_INMEMORY_CONCURRENCY=8",
		"SIGMA_LOCKER_TYPE=redis",
		"SIGMA_NAMESPACE_AUTOCREATE=true",
		"SIGMA_NAMESPACE_VISIBILITY=public",
		"SIGMA_HTTP_ENDPOINT=https://registry.example.com",
		"SIGMA_HTTP_TIMEOUT_READ=90s",
		"SIGMA_HTTP_BODYLIMIT_BLOB=2048",
		"SIGMA_HTTP_TLS_ENABLED=true",
		"SIGMA_STORAGE_TYPE=s3",
		"SIGMA_STORAGE_S3_BUCKET=bucket",
		"SIGMA_STORAGE_S3_FORCEPATHSTYLE=true",
		"SIGMA_PROXY_TLSVERIFY=false",
		"SIGMA_DAEMON_BUILDER_ENABLED=true",
		"SIGMA_DAEMON_BUILDER_TYPE=docker",
		"SIGMA_DAEMON_BUILDER_IMAGE=ghcr.io/go-sigma/sigma-builder:nightly-alpine",
		"SIGMA_DAEMON_BUILDER_TLSVERIFY=false",
		"SIGMA_DAEMON_BUILDER_DOCKER_SOCK=unix:///var/run/docker.sock",
		"SIGMA_DAEMON_BUILDER_EXTRA_HOSTS=host.docker.internal:host-gateway,a:1.2.3.4",
		"SIGMA_DAEMON_GC_RETENTION=48h",
		"SIGMA_DAEMON_SIZERECONCILE_INTERVAL=30m",
		"SIGMA_DAEMON_VULNERABILITY_GRYPE_DBCACHEDIR=/data/grype",
		"SIGMA_AUTH_ADMIN_USERNAME=sigma",
		"SIGMA_AUTH_ADMIN_PASSWORD=secret",
		"SIGMA_AUTH_JWT_PRIVATEKEY=private-key",
		"SIGMA_AUTH_LOGINRATELIMIT_WINDOW=10m",
		"SIGMA_AUTH_OAUTH2_GITHUB_CLIENTSECRET=github-secret",
		"SIGMA_AUTH_CREDENTIALKEY=credential-key",
		"SIGMA_TRACE_ENABLED=true",
		"SIGMA_TRACE_SAMPLERATIO=0.25",
		"SIGMA_ANALYTICS_ENABLED=true",
		"SIGMA_AUDIT_ENABLED=false",
	})
	require.NoError(t, err)

	require.Equal(t, enums.LogLevelTrace, cfg.Log.Level)
	require.Equal(t, enums.LogLevelWarn, cfg.Log.ProxyLevel)
	require.Equal(t, enums.DeployReplica, cfg.Deploy)
	require.Equal(t, enums.DatabasePostgresql, cfg.Database.Type)
	require.Equal(t, 42, cfg.Database.MaxOpenConns)
	require.Equal(t, 3*time.Minute, cfg.Database.ConnMaxLifetime)
	require.Equal(t, "/data/sigma.db", cfg.Database.Sqlite3.Path)
	require.Equal(t, []string{"redis-0:6379", "redis-1:6379"}, cfg.Redis.Addrs)
	require.Equal(t, 3, cfg.Redis.DB)
	require.Equal(t, enums.CacherTypeRedis, cfg.Cache.Type)
	require.Equal(t, 24*time.Hour, cfg.Cache.TTL)
	require.False(t, cfg.Cache.Prewarm.IsEnabled())
	require.Equal(t, enums.WorkQueueTypeDatabase, cfg.WorkQueue.Type)
	require.Equal(t, 8, cfg.WorkQueue.Inmemory.Concurrency)
	require.Equal(t, enums.LockerTypeRedis, cfg.Locker.Type)
	require.True(t, cfg.Namespace.AutoCreate)
	require.Equal(t, enums.VisibilityPublic, cfg.Namespace.Visibility)
	require.Equal(t, "https://registry.example.com", cfg.HTTP.Endpoint)
	require.Equal(t, 90*time.Second, cfg.HTTP.Timeout.ReadTimeout())
	require.Equal(t, int64(2048), cfg.HTTP.BodyLimit.BlobLimit())
	require.True(t, cfg.HTTP.TLS.Enabled)
	require.Equal(t, enums.StorageTypeS3, cfg.Storage.Type)
	require.Equal(t, "bucket", cfg.Storage.S3.Bucket)
	require.True(t, cfg.Storage.S3.ForcePathStyle)
	require.False(t, cfg.Proxy.TlsVerify)
	require.True(t, cfg.Daemon.Builder.Enabled)
	require.Equal(t, enums.BuilderTypeDocker, cfg.Daemon.Builder.Type)
	require.Equal(t, "ghcr.io/go-sigma/sigma-builder:nightly-alpine", cfg.Daemon.Builder.Image)
	require.False(t, cfg.Daemon.Builder.TLSVerifyEnabled())
	require.NotNil(t, cfg.Daemon.Builder.Docker.Sock)
	require.Equal(t, "unix:///var/run/docker.sock", *cfg.Daemon.Builder.Docker.Sock)
	require.Equal(t, []string{"host.docker.internal:host-gateway", "a:1.2.3.4"}, cfg.Daemon.Builder.ExtraHosts)
	require.Equal(t, 48*time.Hour, cfg.Daemon.GC.Retention)
	require.Equal(t, 30*time.Minute, cfg.Daemon.SizeReconcile.Interval)
	require.Equal(t, "/data/grype", cfg.Daemon.Vulnerability.Grype.DBCacheDir)
	require.Equal(t, "sigma", cfg.Auth.Admin.Username)
	require.Equal(t, "secret", cfg.Auth.Admin.Password)
	require.Equal(t, "private-key", cfg.Auth.Jwt.PrivateKey)
	require.Equal(t, 10*time.Minute, cfg.Auth.LoginRateLimit.Window)
	require.Equal(t, "github-secret", cfg.Auth.Oauth2.Github.ClientSecret)
	require.Equal(t, "credential-key", cfg.Auth.CredentialKey)
	require.True(t, cfg.Trace.Enabled)
	require.InDelta(t, 0.25, cfg.Trace.SampleRatio, 1e-9)
	require.True(t, cfg.Analytics.Enabled)
	require.False(t, cfg.Audit.IsEnabled())
}

func TestApplyEnvOverridesAcceptsAlternateSpellings(t *testing.T) {
	var cfg Configuration
	err := cfg.applyEnvOverrides([]string{
		"SIGMA_DAEMON_BUILDER_TLS_VERIFY=false",
		"sigma_daemon_builder_enabled=true",
		"SIGMA_AUTH_LOGIN_RATE_LIMIT_MAX_FAILURES=9",
	})
	require.NoError(t, err)
	require.False(t, cfg.Daemon.Builder.TLSVerifyEnabled())
	require.True(t, cfg.Daemon.Builder.Enabled)
	require.Equal(t, 9, cfg.Auth.LoginRateLimit.MaxFailures)
}

func TestApplyEnvOverridesKeepsFileValues(t *testing.T) {
	cfg := Configuration{
		Daemon: ConfigurationDaemon{Builder: ConfigurationDaemonBuilder{Image: "from-file", Enabled: false}},
	}
	err := cfg.applyEnvOverrides([]string{"SIGMA_DAEMON_BUILDER_ENABLED=true"})
	require.NoError(t, err)
	require.True(t, cfg.Daemon.Builder.Enabled)
	require.Equal(t, "from-file", cfg.Daemon.Builder.Image)
}

func TestApplyEnvOverridesIgnoresUnrelatedVariables(t *testing.T) {
	var cfg Configuration
	err := cfg.applyEnvOverrides([]string{
		"PATH=/usr/local/bin",
		"SIGMA_SERVER=http://localhost:3000",
		"SIGMA_USERNAME=sigma",
		"SIGMA_PASSWORD=sigma",
		"SIGMA_CREDENTIALS_FILE=/tmp/sigma.json",
		"SIGMA_UNKNOWN_SETTING=1",
	})
	require.NoError(t, err)
	require.Equal(t, Configuration{}, cfg)
}

func TestApplyEnvOverridesSkipsStructuredValues(t *testing.T) {
	cfg := Configuration{
		Scm: ConfigurationScm{Providers: []ConfigurationScmProvider{{Provider: enums.ScmProviderGithub}}},
	}
	err := cfg.applyEnvOverrides([]string{
		"SIGMA_SCM_PROVIDERS=github",
		"SIGMA_NAMESPACE_INITIALIZE=library",
	})
	require.NoError(t, err)
	require.Len(t, cfg.Scm.Providers, 1)
	require.Equal(t, enums.ScmProviderGithub, cfg.Scm.Providers[0].Provider)
	require.Empty(t, cfg.Namespace.Initialize)
}

func TestApplyEnvOverridesReportsInvalidValues(t *testing.T) {
	var cfg Configuration
	err := cfg.applyEnvOverrides([]string{
		"SIGMA_LOG_LEVEL=verbose",
		"SIGMA_DAEMON_GC_BATCHSIZE=abc",
		"SIGMA_TRACE_SAMPLERATIO=abc",
		"SIGMA_DATABASE_CONNMAXLIFETIME=forever",
		"SIGMA_DAEMON_BUILDER_ENABLED=maybe",
		// A valid override in the same batch must still be applied.
		"SIGMA_DAEMON_BUILDER_IMAGE=ghcr.io/go-sigma/sigma-builder:nightly-alpine",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "SIGMA_LOG_LEVEL (log.level)")
	require.Contains(t, err.Error(), "SIGMA_DAEMON_GC_BATCHSIZE (daemon.gc.batchSize)")
	require.Contains(t, err.Error(), "SIGMA_TRACE_SAMPLERATIO (trace.sampleRatio)")
	require.Contains(t, err.Error(), "SIGMA_DATABASE_CONNMAXLIFETIME (database.connMaxLifetime)")
	require.Contains(t, err.Error(), "SIGMA_DAEMON_BUILDER_ENABLED (daemon.builder.enabled)")
	require.Equal(t, "ghcr.io/go-sigma/sigma-builder:nightly-alpine", cfg.Daemon.Builder.Image)
}

func TestApplyEnvOverridesEmptyList(t *testing.T) {
	cfg := Configuration{Redis: ConfigurationRedis{Addrs: []string{"redis:6379"}}}
	err := cfg.applyEnvOverrides([]string{"SIGMA_REDIS_ADDRS="})
	require.NoError(t, err)
	require.Empty(t, cfg.Redis.Addrs)
}
