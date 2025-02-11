package main

import (
	"i9rfs/server/initializers"
	"i9rfs/server/middlewares"
	"i9rfs/server/routes/appRoutes"
	"i9rfs/server/routes/authRoutes"
	"log"

	"github.com/gofiber/fiber/v2"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	defer initializers.CleanUp()

	app := fiber.New()

	app.Route("/api/auth", authRoutes.Init)

	app.Use("/api/app", middlewares.Auth)

	app.Route("/api/app", appRoutes.Init)

	app.Listen(":8000")
}
