package db

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryDatabase оборачивает операции с базой данных с телеметрией.
type TelemetryDatabase struct {
	db     *Database
	tracer trace.Tracer
}

// NewTelemetryDatabase возвращает новый экземпляр TelemetryDatabase.
func NewTelemetryDatabase(pool *pgxpool.Pool, cfg Config, tracer trace.Tracer) *TelemetryDatabase {
	return &TelemetryDatabase{
		db:     newDatabase(pool, cfg),
		tracer: tracer,
	}
}

// Get оборачивает Database.Get с телеметрией.
func (td *TelemetryDatabase) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, span := td.tracer.Start(ctx, "Database.Get", trace.WithAttributes(
		attribute.String("db.statement", query),
	))
	defer span.End()

	err := pgxscan.Get(ctx, td.db.cluster, dest, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "Success")
	}
	return err
}

// Select оборачивает Database.Select с телеметрией.
func (td *TelemetryDatabase) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, span := td.tracer.Start(ctx, "Database.Select", trace.WithAttributes(
		attribute.String("db.statement", query),
	))
	defer span.End()

	err := pgxscan.Select(ctx, td.db.cluster, dest, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "Success")
	}
	return err
}

// Exec оборачивает Database.Exec с телеметрией.
func (td *TelemetryDatabase) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, span := td.tracer.Start(ctx, "Database.Exec", trace.WithAttributes(
		attribute.String("db.statement", query),
	))
	defer span.End()

	tag, err := td.db.cluster.Exec(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "Success")
	}
	return tag, err
}

// ExecQueryRow оборачивает Database.ExecQueryRow с телеметрией.
func (td *TelemetryDatabase) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	ctx, span := td.tracer.Start(ctx, "Database.ExecQueryRow", trace.WithAttributes(
		attribute.String("db.statement", query),
	))
	// Примечание: для строк мы вызываем defer span.End() в вызывающем коде после сканирования,
	// иначе вы можете завершить здесь, если данные не нужны для дополнительной телеметрии.

	row := td.db.cluster.QueryRow(ctx, query, args...)
	span.End()
	return row
}
