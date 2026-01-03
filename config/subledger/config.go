package subledger

import (
	"github.com/ChotongW/grit_demo_wallet/config"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServiceConfig struct {
	LogLevel       string          `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	LogFormatJson  bool            `yaml:"log_format_json" env:"LOG_FORMAT_JSON" env-default:"false"`
	LogColor       bool            `yaml:"log_color" env:"LOG_COLOR" env-default:"false"`
	LogLineDetails bool            `yaml:"log_line_details" env:"LOG_LINE_DETAILS" env-default:"false"`
	Port           int             `yaml:"grpc_port" env:"GRPC_PORT" env-default:"50051"`
	DbConfig       config.DbConfig `yaml:"database"`
}

func LoadConfig(path string) (*ServiceConfig, error) {
	var cfg ServiceConfig

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
