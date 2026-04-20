package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Connect возвращает готовый *sqlx.DB (один раз создаём пул).
func Connect(ctx context.Context, dsn string) (*sqlx.DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, wrapper.Wrapf(err, "failed to parse postgres dsn")
	}

	// В serverless + PgBouncer transaction pooler prepared statements часто ломаются (42P05 stmt already exists).
	// Поэтому используем simple protocol и отключаем stmt cache.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	cfg.ConnConfig.StatementCacheCapacity = 0
	cfg.ConnConfig.DescriptionCacheCapacity = 0

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
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
