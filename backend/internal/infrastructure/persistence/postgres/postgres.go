package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Connect возвращает готовый *sqlx.DB (один раз создаём пул).
func Connect(ctx context.Context, dsn string) (*sqlx.DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, wrapper.Wrapf(err, "failed to create pgxpool")
	}

	pgxDB := stdlib.OpenDBFromPool(pool)
	db := sqlx.NewDb(pgxDB, "pgx")

	err = db.PingContext(ctx)
	if err != nil {
		return nil, wrapper.Wrapf(err, "failed to ping database")
	}

	return db, nil
}
