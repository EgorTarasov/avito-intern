package main

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/config"
	migration "avito-intern/internal/migrations"
	"avito-intern/pkg/db"
	"avito-intern/server"
	"avito-intern/storage"
	"context"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	db, err := db.NewDB(ctx, cfg.PG)
	if err != nil {
		panic(err)
	}

	readyFn := migration.Migrate(db)

	pg := storage.NewPgUserRepo(db)
	authService := auth.NewService(&cfg.Auth, pg)
	authHandlers := auth.NewAuthHandlers(authService)
	router := server.New(cfg.HTTP, func() bool {
		return readyFn()
	})

	router.Add(authHandlers)
	if err := router.Run(); err != nil {
		panic(err)
	}
}
