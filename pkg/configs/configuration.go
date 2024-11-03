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

package configs

import (
	"time"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

var configuration = &Configuration{}

// GetConfiguration ...
func GetConfiguration() *Configuration {
	return configuration
}

// SetConfiguration ...
func SetConfiguration(c *Configuration) {
	configuration = c
}

// Configuration ...
type Configuration struct {
	Log       ConfigurationLog       `yaml:"log"`
	Database  ConfigurationDatabase  `yaml:"database"`
	Deploy    enums.Deploy           `yaml:"deploy"`
	Redis     ConfigurationRedis     `yaml:"redis"`
	Badger    ConfigurationBadger    `yaml:"badger"`
	Cache     ConfigurationCache     `yaml:"cache"`
	WorkQueue ConfigurationWorkQueue `yaml:"workqueue"`
	Locker    ConfigurationLocker    `yaml:"locker"`
	Namespace ConfigurationNamespace `yaml:"namespace"`
	HTTP      ConfigurationHTTP      `yaml:"http"`
	Storage   ConfigurationStorage   `yaml:"storage"`
	Proxy     ConfigurationProxy     `yaml:"proxy"`
	Daemon    ConfigurationDaemon    `yaml:"daemon"`
	Auth      ConfigurationAuth      `yaml:"auth"`
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
	Type       enums.Database                  `yaml:"type"`
	Sqlite3    ConfigurationDatabaseSqlite3    `yaml:"sqlite3"`
	Mysql      ConfigurationDatabaseMysql      `yaml:"mysql"`
	Postgresql ConfigurationDatabasePostgresql `yaml:"postgresql"`
}

// ConfigurationRedis ...
type ConfigurationRedis struct {
	Type enums.RedisType `yaml:"type"`
	URL  string          `yaml:"url"`
}

// ConfigurationBadger ...
type ConfigurationBadger struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// ConfigurationCacheRedis ...
type ConfigurationCacheRedis struct {
	Ttl time.Duration `yaml:"ttl"`
}

// ConfigurationCacheBadger ...
type ConfigurationCacheBadger struct {
	Ttl time.Duration `yaml:"ttl"`
}

// ConfigurationCacheInmemory ...
type ConfigurationCacheInmemory struct {
	Size int `yaml:"size"`
}

// ConfigurationCache ...
type ConfigurationCache struct {
	Type     enums.CacherType           `yaml:"type"`
	Prefix   string                     `yaml:"prefix"`
	Redis    ConfigurationCacheRedis    `yaml:"redis"`
	Inmemory ConfigurationCacheInmemory `yaml:"inmemory"`
	Badger   ConfigurationCacheBadger   `yaml:"badger"`
}

type ConfigurationWorkQueueRedis struct {
	Concurrency int `yaml:"concurrency"`
}

type ConfigurationWorkQueueDatabase struct {
}

type ConfigurationWorkQueueKafka struct {
}

type ConfigurationWorkQueueInmemmory struct {
	Concurrency int `yaml:"concurrency"`
}

// ConfigurationWorkQueue ...
type ConfigurationWorkQueue struct {
	Type     enums.WorkQueueType             `yaml:"type"`
	Redis    ConfigurationWorkQueueRedis     `yaml:"redis"`
	Database ConfigurationWorkQueueDatabase  `yaml:"database"`
	Kafka    ConfigurationWorkQueueKafka     `yaml:"kafka"`
	Inmemory ConfigurationWorkQueueInmemmory `yaml:"inmemory"`
}

// ConfigurationLockerBadger ...
type ConfigurationLockerBadger struct{}

// ConfigurationLockerRedis ...
type ConfigurationLockerRedis struct{}

// ConfigurationLocker ...
type ConfigurationLocker struct {
	Type   enums.LockerType          `yaml:"type"`
	Badger ConfigurationLockerBadger `yaml:"badger"`
	Redis  ConfigurationLockerRedis  `yaml:"redis"`
	Prefix string                    `yaml:"prefix"`
}

// ConfigurationNamespace ...
type ConfigurationNamespace struct {
	AutoCreate bool             `yaml:"autoCreate"`
	Visibility enums.Visibility `yaml:"visibility"`
}

// ConfigurationHttpTLS ...
type ConfigurationHttpTLS struct {
	Enabled     bool   `yaml:"enabled"`
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

// ConfigurationHTTP ...
type ConfigurationHTTP struct {
	Endpoint                     string               `yaml:"endpoint"`
	InternalEndpoint             string               `yaml:"internalEndpoint"`
	InternalDistributionEndpoint string               `yaml:"internalDistributionEndpoint"`
	TLS                          ConfigurationHttpTLS `yaml:"tls"`
}

// ConfigurationStorageFilesystem ...
type ConfigurationStorageFilesystem struct {
	Path string `yaml:"path"`
}

// ConfigurationStorageS3 ...
type ConfigurationStorageS3 struct {
	Ak             string `yaml:"ak"`
	Sk             string `yaml:"sk"`
	Endpoint       string `yaml:"endpoint"`
	Region         string `yaml:"region"`
	Bucket         string `yaml:"bucket"`
	ForcePathStyle bool   `yaml:"forcePathStyle"`
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
	Retention time.Duration `yaml:"retention"`
	Cron      string        `yaml:"cron"`
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

// ConfigurationDaemonPodman ...
type ConfigurationDaemonPodman struct {
	URI string `yaml:"uri"`
}

// ConfigurationDaemonBuilder ...
type ConfigurationDaemonBuilder struct {
	Enabled    bool                          `yaml:"enabled"`
	Type       enums.BuilderType             `yaml:"type"`
	Image      string                        `yaml:"image"`
	Docker     ConfigurationDaemonDocker     `yaml:"docker"`
	Kubernetes ConfigurationDaemonKubernetes `yaml:"kubernetes"`
	Podman     ConfigurationDaemonPodman     `yaml:"podman"`
}

// ConfigurationDaemon ...
type ConfigurationDaemon struct {
	Builder ConfigurationDaemonBuilder `yaml:"builder"`
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
	RefreshTtl time.Duration `yaml:"refreshTtl"`
	PrivateKey string        `yaml:"privateKey"`
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
