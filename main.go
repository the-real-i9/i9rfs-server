package main

import (
	"i9rfs/initializers"
	"i9rfs/routes/appRoutes"
	"i9rfs/routes/authRoutes"
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

	app.Route("/api/app", appRoutes.Init)

	app.Listen(":8000")
}
