package gateway

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServiceConfig struct {
	LogLevel          string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	LogFormatJson     bool          `yaml:"log_format_json" env:"LOG_FORMAT_JSON" env-default:"false"`
	LogColor          bool          `yaml:"log_color" env:"LOG_COLOR" env-default:"false"`
	LogLineDetails    bool          `yaml:"log_line_details" env:"LOG_LINE_DETAILS" env-default:"false"`
	HttpPort          int           `yaml:"http_port" env:"HTTP_PORT" env-default:"8080"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env:"READ_HEADER_TIMEOUT" env-default:"5s"`
	ReadTimeout       time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"10s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60s"`
	SubledgerService  string        `yaml:"subledger_service" env:"SUBLEDGER_SERVICE" env-default:"localhost:50051"`
	AccountsService   string        `yaml:"accounts_service" env:"ACCOUNTS_SERVICE" env-default:"localhost:50052"`
	GrpcTimeout       time.Duration `yaml:"grpc_timeout" env:"GRPC_TIMEOUT" env-default:"5s"`
	ApiKey            string        `yaml:"api_key" env:"API_KEY" env-default:"secret"`
}

func LoadConfig(path string) (*ServiceConfig, error) {
	var cfg ServiceConfig
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}

	fmt.Printf("Loaded config: %+v\n", cfg)

	return &cfg, nil
}
