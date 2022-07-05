package configs

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
)

type SessionStoreConfig struct {
	CookieHTTPOnly bool `yaml:"cookie_http_only"`
	CookieSecure   bool `yaml:"cookie_secure"`
	Expiration     int  `yaml:"expiration_in_hours"`
}

func NewDefaultSessionStore() *session.Store {
	return session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSecure:   false, // for https
		Expiration:     time.Hour * 5,
	})
}
