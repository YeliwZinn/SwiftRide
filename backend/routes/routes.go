package routes

import (
    "fmt"
    "net/http"
    "strings"
    "uber-clone/auth"
    "uber-clone/websockets"
    "uber-clone/middleware"
    "uber-clone/controllers"

    "github.com/gin-gonic/gin"
)

func SetupRouter(hub *websockets.Hub) *gin.Engine {
    router := gin.Default()

    // WebSocket endpoint
    router.GET("/ws", func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
            return
        }

        claims, err := auth.ValidateToken(tokenString)
        if err != nil {
            fmt.Println("JWT validation error:", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            return
        }

        conn, err := websockets.Upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
            return
        }

        client := &websockets.Client{
            Conn:   conn,
            UserID: claims.UserID,
            Role:   claims.Role,
        }

        hub.Register <- client
        defer func() { hub.Unregister <- client }()

        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                break
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

        // Feedback route
        authGroup.POST("/feedback/:ride_id", controllers.SubmitFeedback)
    }

    return router
}
