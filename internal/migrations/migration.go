package migration

import (
	"avito-intern/pkg/db"
	"embed"
	"log/slog"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

func Migrate(db *db.Database) func() bool {
	readyCh := make(chan bool)
	go func() {
		goose.SetBaseFS(embedMigrations)
		if err := goose.SetDialect("postgres"); err != nil {
			slog.Error("migration err", "error", err)
		}
		conn := db.GetSQLConn()
		defer conn.Close()
		if err := goose.Up(conn, "."); err != nil {
			slog.Error("migration err", "error", err)
		}
		readyCh <- true
	}()

	return func() bool {
		select {
		case <-readyCh:
			return true
		default:
			return false
		}
	}
}
