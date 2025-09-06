package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStoragePostgres struct {
	pool *pgxpool.Pool
}

func NewOrderStoragePostgres(pool *pgxpool.Pool) *OrderStoragePostgres {
	return &OrderStoragePostgres{pool: pool}
}
