package auth

import "time"

type Config struct {
	JWTSecret           string        `env:"JWT_SECRET" envDefault:"superSecret"`
	TokenExpireDuration time.Duration `env:"TOKEN_EXPIRE_DURATION" envDefault:"24h"`
}
