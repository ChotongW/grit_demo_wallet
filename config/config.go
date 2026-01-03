package config

import (
	"time"
)

type DbConfig struct {
	Type              string        `yaml:"database_type" env:"DATABASE_TYPE" env-default:"postgres"`
	LogLevel          string        `yaml:"database_log_level" env:"DATABASE_LOG_LEVEL" env-default:"info"`
	Host              string        `yaml:"database_host" env:"DATABASE_HOST" env-default:"localhost"`
	Port              string        `yaml:"database_port" env:"DATABASE_PORT" env-default:"5432"`
	Username          string        `yaml:"database_username" env:"DATABASE_USERNAME" env-default:"postgres"`
	Password          string        `yaml:"database_password" env:"DATABASE_PASSWORD" env-default:"password"`
	DatabaseName      string        `yaml:"database_name" env:"DATABASE_NAME" env-default:"postgres_db"`
	SslMode           string        `yaml:"database_ssl_mode" env:"DATABASE_SSL_MODE" env-default:"disable"`
	DatabaseSchema    string        `yaml:"database_schema" env:"DATABASE_SCHEMA" env-default:"public"`
	MaxOpenConns      int32         `yaml:"database_max_open_conns" env:"DATABASE_MAX_OPEN_CONNS" env-default:"10"`
	MaxConnIdleTime   time.Duration `yaml:"database_max_conn_idle_time" env:"DATABASE_MAX_CONN_IDLE_TIME" env-default:"5m"`
	MaxConnLifetime   time.Duration `yaml:"database_max_conn_lifetime" env:"DATABASE_MAX_CONN_LIFETIME" env-default:"30m"`
	HealthCheckPeriod time.Duration `yaml:"database_health_check_period" env:"DATABASE_HEALTH_CHECK_PERIOD" env-default:"1m"`
}

// func LoadConfig(path string) (*DbConfig, error) {
// 	viper.AddConfigPath(path)
// 	viper.SetConfigName("config")
// 	viper.SetConfigType("yaml")

// 	viper.AutomaticEnv()
// 	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

// 	if err := viper.ReadInConfig(); err != nil {
// 		return nil, err
// 	}

// 	var cfg DbConfig
// 	if err := viper.UnmarshalKey("database", &cfg); err != nil {
// 		return nil, err
// 	}

// 	return &cfg, nil
// }
