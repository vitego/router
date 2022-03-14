package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vitego/router/manager"
)

type middlewareHandler interface {
	Defaults() []Middleware
	Middleware() map[string]Middleware
}

type Middleware interface {
	Run(c *fiber.Ctx, m *manager.Manager) error
}
