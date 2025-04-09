package routes

import (
	"log"
	"net/http"
	"uber-clone/auth"
	"uber-clone/controllers"
	"uber-clone/middleware"
	"uber-clone/websockets"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter(hub *websockets.Hub) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // All
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// WebSocket endpoint
	router.GET("/ws", func(c *gin.Context) {
		tokenString := c.Query("token") // Changed from header to query param
		if tokenString == "" {
			// Abort before attempting upgrade
			c.Writer.WriteHeader(http.StatusUnauthorized)
			c.Writer.Write([]byte("Token required"))
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		conn, err := websockets.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}

		client := &websockets.Client{
			Conn:   conn,
			UserID: claims.UserID,
			Role:   claims.Role,
		}

		websockets.WS_HUB.Register <- client
		defer func() {
			websockets.WS_HUB.Unregister <- client
		}()

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}

			log.Println("Message received:", string(msg))

			// You can respond back to client to confirm it's alive
			if string(msg) == `{"type":"ping"}` {
				err = conn.WriteMessage(mt, []byte(`{"type":"pong"}`))
				if err != nil {
					log.Println("Write error:", err)
					break
				}
			}
		}

	})

	// Auth routes
	router.POST("/signup", controllers.Signup)
	router.POST("/login", controllers.Login)

	// Protected routes
	authGroup := router.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{

		// Ride-related routes
		rideGroup := authGroup.Group("/rides")
		{
			rideGroup.POST("/", controllers.RequestRide)
			rideGroup.GET("/:ride_id", controllers.GetRideDetails)
			rideGroup.POST("/:ride_id/verifyOTP", controllers.VerifyOTP)
			rideGroup.POST("/:ride_id/respond", controllers.HandleDriverResponse)
			rideGroup.POST("/:ride_id/complete", controllers.CompleteRide)
			rideGroup.POST("/:ride_id/cancel", controllers.CancelRide)
			rideGroup.POST("/:ride_id/pay", controllers.HandlePayment)
			rideGroup.POST("/:ride_id/confirm-payment", controllers.ConfirmPayment)
		}

		authGroup.GET("/profile", controllers.Profile)

		// Feedback route
		authGroup.POST("/feedback/:ride_id", controllers.SubmitFeedback)
	}

	return router
}
