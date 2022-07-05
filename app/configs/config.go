package configs

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
)

type Config struct {
	AuthKey    string
	UserIDKey  string
	BcryptCost int
	Port       string
}

func NewDefaultSessionStore() *session.Store {
	return session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSecure:   false, // for https
		Expiration:     time.Hour * 5,
	})
}

func NewDefaultServiceConf() Config {
	return Config{
		AuthKey:    "authenticated",
		UserIDKey:  "user_id",
		BcryptCost: 14,
		Port:       ":5000",
	}
}
