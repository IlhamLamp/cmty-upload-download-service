package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

type AuthMiddlewareResponse struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}

func AuthMiddleware(jwtSecretKey string) gin.HandlerFunc {
    return func (c *gin.Context) {

        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, AuthMiddlewareResponse{
                Status: http.StatusUnauthorized,
                Message: "Authorization header required",
                Error: "Authorization header required",
            })
            c.Abort()
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenStr == authHeader {
            c.JSON(http.StatusUnauthorized, AuthMiddlewareResponse{
                Status: http.StatusUnauthorized,
                Message: "Bearer token required",
                Error: "Bearer token required",
            })
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(jwtSecretKey), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, AuthMiddlewareResponse{
                Status: http.StatusUnauthorized,
                Message: "Invalid or expired token",
                Error: err.Error(),
            })
            c.Abort()
            return
        }

        c.Next()
    }
}