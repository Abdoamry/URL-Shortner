package main

import (
	"log"
	"os"
    "github.com/Abdoamry/URL-Shortner/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App){
	app.Get("/:url", routes.Resolve)
	app.Post("/api/v1" , routes.Shorten)
}


func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

			
	app := fiber.New()
	setupRoutes(app)
	app.Use(logger.New())
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))

}