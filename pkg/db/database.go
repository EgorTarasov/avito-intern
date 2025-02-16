package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // required by goose sql.DB as driver
)

// Database обертка для работы с pgxpool.Pool.
type Database struct {
	cfg     Config
	cluster *pgxpool.Pool
}

func (db Database) GetSQLConn() *sql.DB {
	dsn := createDsn(db.cfg)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open connection: %v", err)
	}
	return conn
}

func newDatabase(cluster *pgxpool.Pool, cfg Config) *Database {
	return &Database{cluster: cluster, cfg: cfg}
}

// GetPool возвращает пул соединений к базе данных.
func (db Database) GetPool(_ context.Context) *pgxpool.Pool {
	return db.cluster
}

// Get возвращает одну запись.
func (db Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

// Select выполняет запрос и возвращает результат.
func (db Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

// Exec выполняет запрос.
func (db Database) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

// ExecQueryRow выполняет запрос и возвращает строку.
func (db Database) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}
