package main

import (
	"log"
	"os"

	
	"uber-clone/db"
	"uber-clone/routes"
	"uber-clone/websockets"

	//"uber-clone/services"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	
)

func main() {
   
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
    // Set Stripe API Key
    stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
    if stripe.Key == "" {
        log.Fatal("Stripe secret key is missing. Check your .env file.")
    }
    

    db.InitMongoDB()

	
	websockets.WS_HUB = websockets.NewHub()
	go websockets.WS_HUB.Run() // Start the hub

	router := routes.SetupRouter(websockets.WS_HUB)
	router.Run(":8080")
}
