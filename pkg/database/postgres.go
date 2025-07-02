package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDatabase struct {
	Pool *pgxpool.Pool
}

func New(dbUrl string) *PostgresDatabase {
	slog.Info("connecting to database")
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		panic("failed to connect database")
	}
	slog.Info("connected to database")
	return &PostgresDatabase{Pool: pool}
}
