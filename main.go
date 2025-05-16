package main

import (
	"i9rfs/src/initializers"
	"i9rfs/src/routes/appRoutes"
	"i9rfs/src/routes/authRoutes"
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

	app.Route("/api/auth", authRoutes.Route)

	app.Route("/api/app", appRoutes.Route)

	app.Listen(":8000")
}
