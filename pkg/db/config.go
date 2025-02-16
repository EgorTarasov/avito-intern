package db

import (
	"fmt"
)

// Config содержит конфигурацию подключения к базе данных.
type Config struct {
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT"`
	DB       string `yaml:"db" env:"POSTGRES_DB"`
}

func createDsn(cfg Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
}
