package main

import (
	"configs"
	"log"
	"middleware"
	"model"
	"router"
	"services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	conf, err := configs.GetConfigFromFile[configs.AllConfig]("config.yaml")
	if err != nil {
		log.Fatalf("error loading configs: %s", err)
	}

	store := configs.NewSessionStore(conf.SessionStoreConfig)
	rout := fiber.New()

	mware := middleware.NewMiddleware(conf.ServiceConfig, store)

	db, err := model.NewDB(conf.DBConfig)
	defer func(db model.IDB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("error closing DB: %s", err)
		}
	}(db)
	if err != nil {
		log.Fatalf("error initializing DB: %s", err)
	}

	serv := services.NewService(db, store, conf.ServiceConfig)

	Router := router.NewRouter(conf.ServiceConfig, store, rout, mware.AuthMiddleware, serv)
	Router.Setup()
	Router.Run()
}
