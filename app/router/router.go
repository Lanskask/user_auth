package router

import (
	"configs"
	"log"
	"services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type Router struct {
	conf       configs.ServiceConfig
	store      *session.Store
	router     *fiber.App
	middleware fiber.Handler
	service    services.IService
}

func NewRouter(
	conf configs.ServiceConfig,
	store *session.Store,
	router *fiber.App,
	middleware fiber.Handler,
	service services.IService,
) *Router {
	return &Router{conf: conf, store: store, router: router, middleware: middleware, service: service}
}

func (r *Router) Setup() {
	r.router.Use(r.middleware, cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "*", // to access from localhost
		AllowHeaders:     "Access-Control-Allow-Origin, Content-Type, Origin, Accept",
	}))

	r.router.Post("/auth/register", r.service.Register)
	r.router.Post("/auth/login", r.service.Login)
	r.router.Post("/auth/logout", r.service.Logout)
	r.router.Post("/auth/healthcheck", r.service.HealthCheck)

	r.router.Get("/user", r.service.GetUser)
}

func (r *Router) Run() {
	if err := r.router.Listen(r.conf.Port); err != nil {
		log.Fatalf("Error listenning port %s", r.conf.Port)
		return
	}
}
