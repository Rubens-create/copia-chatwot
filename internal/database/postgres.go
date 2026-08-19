package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database url: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	var pool *pgxpool.Pool
	var pingErr error

	// Retry connecting up to 10 times with backoff
	for attempt := 1; attempt <= 10; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			ctxPing, cancel := context.WithTimeout(ctx, 3*time.Second)
			pingErr = pool.Ping(ctxPing)
			cancel()
			if pingErr == nil {
				log.Printf("[Database] Connected successfully to PostgreSQL")
				return &DB{Pool: pool}, nil
			}
			pool.Close()
		} else {
			pingErr = err
		}

		log.Printf("[Database] Connection attempt %d/10 failed: %v. Retrying in 2s...", attempt, pingErr)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return nil, fmt.Errorf("failed to connect to database after 10 attempts: %w", pingErr)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) Ping(ctx context.Context) error {
	if db.Pool == nil {
		return fmt.Errorf("database pool is nil")
	}
	return db.Pool.Ping(ctx)
}

func (db *DB) RunMigration(ctx context.Context, sqlScript string) error {
	_, err := db.Pool.Exec(ctx, sqlScript)
	if err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}
	log.Printf("[Database] Migration script executed successfully")
	return nil
}
