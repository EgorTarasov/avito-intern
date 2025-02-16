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

type txKey struct{}

// Database обертка для работы с pgxpool.Pool.
type Database struct {
	cfg     Config
	cluster *pgxpool.Pool
}

func (db *Database) GetSQLConn() *sql.DB {
	dsn := createDsn(db.cfg)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open connection: %v", err)
	}
	return conn
}

type Querier interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (db *Database) getConn(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.cluster
}

func (db *Database) putTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func newDatabase(cluster *pgxpool.Pool, cfg Config) *Database {
	return &Database{cluster: cluster, cfg: cfg}
}

// GetPool возвращает пул соединений к базе данных.
func (db *Database) GetPool(_ context.Context) *pgxpool.Pool {
	return db.cluster
}

// Get возвращает одну запись.
func (db *Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.getConn(ctx), dest, query, args...)
}

// Select выполняет запрос и возвращает результат.
func (db *Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.getConn(ctx), dest, query, args...)
}

// Exec выполняет запрос.
func (db *Database) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.getConn(ctx).Exec(ctx, query, args...)
}

// ExecQueryRow выполняет запрос и возвращает строку.
func (db *Database) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.getConn(ctx).QueryRow(ctx, query, args...)
}

func (db *Database) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.cluster.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	ctx = db.putTx(ctx, tx)
	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
