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
	serviceConf := configs.NewDefaultServiceConf()
	dbConf := configs.NewDefaultDBConfig()

	store := configs.NewDefaultSessionStore()
	rout := fiber.New()

	mware := middleware.NewMiddleware(serviceConf, store)

	db, err := model.NewDB(dbConf)
	defer func(db model.IDB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("error closing DB: %s", err)
		}
	}(db)
	if err != nil {
		log.Fatalf("error initializing DB: %s", err)
	}

	serv := services.NewService(db, store, serviceConf)

	Router := router.NewRouter(serviceConf, store, rout, mware.AuthMiddleware, serv)
	Router.Setup()
	Router.Run()
}
