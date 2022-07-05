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

func NewSessionStore(conf SessionStoreConfig) *session.Store {
	return session.New(session.Config{
		CookieHTTPOnly: conf.CookieHTTPOnly,
		CookieSecure:   conf.CookieSecure, // for https
		Expiration:     time.Hour * time.Duration(conf.Expiration),
	})
}
