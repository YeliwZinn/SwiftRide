package middleware

import (
	"net/http"
	"strings"
	"uber-clone/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Validate Bearer Token Format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Empty token provided"})
			c.Abort()
			return
		}

		// Validate Token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
                var errMsg string
                if ve.Errors&jwt.ValidationErrorExpired != 0 {
                    errMsg = "Token expired"
                } else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
                    errMsg = "Invalid token signature"
                } else {

                    errMsg = "Invalid token"
                }
                c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
                c.Abort()
                return
            }
    }
            

		// Store user details in context
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		// Proceed to the next handler
		c.Next()
	}
}
