package main

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/coin"
	"avito-intern/internal/config"
	"avito-intern/internal/merch"
	migration "avito-intern/internal/migrations"
	"avito-intern/pkg/db"
	"avito-intern/server"
	"avito-intern/storage"
	"context"
)

func main() {
	ctx := context.Background()
	cfg := config.NewConfig()
	database, err := db.NewDB(ctx, cfg.PG)
	if err != nil {
		panic(err)
	}

	readyFn := migration.Migrate(database)

	pg := storage.NewRepo(database)
	authService := auth.NewService(&cfg.Auth, pg)
	authHandlers := auth.NewAuthHandlers(authService)
	router := server.New(cfg.HTTP, func() bool {
		return readyFn()
	})

	coinService := coin.NewService(authService, pg)
	coinHandlers := coin.NewCoinHandler(coinService, authHandlers)

	merchService := merch.NewService(authService, coinService, pg)
	merchHandlers := merch.NewMerchHandler(merchService, authHandlers)

	router.Add(authHandlers)
	router.Add(coinHandlers)
	router.Add(merchHandlers)
	if err := router.Run(); err != nil {
		panic(err)
	}
}
