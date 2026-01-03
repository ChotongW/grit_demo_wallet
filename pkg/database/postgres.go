package database

import (
	"context"
	"fmt"
	"os"

	"github.com/ChotongW/grit_demo_wallet/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type PostgresDb struct {
	Pool   *pgxpool.Pool
	logger *logrus.Entry
}

func New(config *config.DbConfig, logger *logrus.Logger) (pgdb *PostgresDb, err error) {
	pgdb = &PostgresDb{
		logger: logger.WithFields(logrus.Fields{
			"package": "db/postgres",
		}),
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.DatabaseName,
		config.SslMode,
		config.DatabaseSchema,
	)

	connectConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	connectConf.MaxConns = config.MaxOpenConns
	connectConf.MaxConnIdleTime = config.MaxConnIdleTime
	connectConf.MaxConnLifetime = config.MaxConnLifetime
	connectConf.HealthCheckPeriod = config.HealthCheckPeriod

	if s := os.Getenv("TZ"); s != "" {
		connectConf.ConnConfig.RuntimeParams["timezone"] = s
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), connectConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping the database: %w", err)
	}

	pgdb.Pool = pool
	pgdb.logger.Info("Successfully connected to PostgreSQL")

	return pgdb, nil
}

func (p *PostgresDb) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
