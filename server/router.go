package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Router struct {
	app  *fiber.App
	cfg  Config
	root fiber.Router
}

type Module interface {
	Init(router fiber.Router)
}

func (r *Router) Add(module Module) {
	module.Init(r.root)
}

func (r *Router) Run() error {
	return r.app.Listen(fmt.Sprintf(":%d", r.cfg.Port))
}

func New(cfg Config, readyFunc func() bool) *Router {
	app := fiber.New()

	api := app.Group("/api")

	r := Router{
		app:  app,
		root: api,
		cfg:  cfg,
	}
	r.initMiddlewares(readyFunc)
	return &r
}

func (r *Router) initMiddlewares(readyFunc func() bool) {
	r.app.Use(recover.New())
	r.app.Use(logger.New())
	r.app.Use(cors.New(cors.Config{
		AllowOrigins: r.cfg.AllowHeaders,
		AllowHeaders: r.cfg.AllowHeaders,
	}))
	r.app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(_ *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",
		ReadinessProbe: func(_ *fiber.Ctx) bool {
			return readyFunc()
		},
		ReadinessEndpoint: "/ready",
	}))
}
