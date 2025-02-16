package config

import (
	"avito-intern/internal/auth"
	"avito-intern/pkg/db"
	"avito-intern/server"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PG   db.Config
	HTTP server.Config
	Auth auth.Config
}

func NewConfig() Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
