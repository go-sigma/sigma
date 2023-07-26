package configs

import "github.com/go-sigma/sigma/pkg/types/enums"

// ConfigurationLog ...
type ConfigurationLog struct {
	Level      enums.LogLevel
	ProxyLevel enums.LogLevel
}

// ConfigurationDatabaseSqlite3 ...
type ConfigurationDatabaseSqlite3 struct {
	Path string
}

// ConfigurationDatabaseMysql ...
type ConfigurationDatabaseMysql struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// ConfigurationDatabase ...
type ConfigurationDatabasePostgresql struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SslMode  bool
}

// ConfigurationDatabase ...
type ConfigurationDatabase struct {
	Type       enums.Database
	Sqlite3    ConfigurationDatabaseSqlite3
	Mysql      ConfigurationDatabaseMysql
	Postgresql ConfigurationDatabasePostgresql
}

// ConfigurationRedis ...
type ConfigurationRedis struct {
	Type enums.RedisType
	Url  string
}

// ConfigurationCache ...
type ConfigurationCache struct {
	Type enums.CacheType
}

// ConfigurationWorkQueue ...
type ConfigurationWorkQueue struct {
	Type enums.WorkQueueType
}

// ConfigurationNamespace ...
type ConfigurationNamespace struct {
	AutoCreate bool
	Visibility enums.Visibility
}

// Configuration ...
type Configuration struct {
	Log       ConfigurationLog
	Database  ConfigurationDatabase
	Deploy    enums.Deploy
	Redis     ConfigurationRedis
	Cache     ConfigurationCache
	WorkQueue ConfigurationWorkQueue
	Namespace ConfigurationNamespace
}
