package auth

import (
	"fmt"
	"time"
	"uber-clone/config"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(userID, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.MustGetEnv("JWT_SECRET")))
}

func ValidateToken(tokenString string) (*Claims, error) {
    fmt.Println("Received Token:", tokenString) // Debug log

    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        secret := config.MustGetEnv("JWT_SECRET")
        fmt.Println("Using Secret Key:", secret) 
        return []byte(secret), nil
    })

    if err != nil {
        fmt.Println("JWT Parsing Error:", err) // Debug log
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        fmt.Println("Invalid Token or Claims Error") // Debug log
        return nil, jwt.ErrSignatureInvalid
    }

    fmt.Println("Extracted Claims:", claims) // Debug log
    return claims, nil
}
