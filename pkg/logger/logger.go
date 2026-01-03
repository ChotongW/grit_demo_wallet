package logger

import (
	"github.com/sirupsen/logrus"
)

type LogConfig struct {
	Level          string
	FormatJson     bool
	Color          bool
	LogLineDetails bool
}

func NewLogger(cfg LogConfig) *logrus.Logger {
	logger := logrus.New()
	if cfg.LogLineDetails {
		logger.SetReportCaller(true)
	}
	if cfg.FormatJson {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   cfg.Color,
			FullTimestamp: true,
		})
	}
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	return logger
}
