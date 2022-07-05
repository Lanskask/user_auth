package middleware

import (
	"configs"
	"handlers"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Middleware struct {
	config configs.ServiceConfig
	store  *session.Store
}

func NewMiddleware(config configs.ServiceConfig, store *session.Store) *Middleware {
	return &Middleware{config: config, store: store}
}

func (m *Middleware) AuthMiddleware(c *fiber.Ctx) error {
	sess, err := m.store.Get(c)

	if strings.Split(c.Path(), "/")[1] == "auth" {
		return c.Next()
	}

	if err != nil {
		return handlers.HandleUnAuth(c)
	}

	if sess.Get(m.config.AuthKey) == nil {
		return handlers.HandleUnAuth(c)
	}

	return c.Next()
}
