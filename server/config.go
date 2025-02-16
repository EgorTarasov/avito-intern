package server

type Config struct {
	Port         int    `env:"HTTP_PORT"`
	AllowOrigins string `env:"HTTP_ORIGINS"`
	AllowHeaders string `env:"HTTP_HEADERS"`
}
