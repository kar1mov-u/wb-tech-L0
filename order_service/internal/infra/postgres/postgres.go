package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//this file should handle connection to the postgres

type Config struct {
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port     uint16 `yaml:"port" env:"PORT" env-default:"5432"`
	Username string `yaml:"user" env:"USER" env-default:"postgres"`
	Password string `yaml:"password" env:"PASSWORD" env-default:"1234"`
	Database string `yaml:"db" env:"DB" env-default:"postgres"`
	MaxConns int32  `yaml:"max_conn" env:"MAX_CONN" env-default:"10"`
	MinConns int32  `yaml:"min_conn" env:"MIN_CONN" env-default:"5"`
}

func New(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	connstring := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_min_conns=%d&pool_max_conns=%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.MinConns,
		config.MaxConns,
	)

	conn, err := pgxpool.New(ctx, connstring)
	if err != nil {
		return conn, fmt.Errorf("unable to connect to the postgres:%v", err)
	}
	return conn, nil
}
