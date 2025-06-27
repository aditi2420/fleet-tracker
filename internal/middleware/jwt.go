package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// NewJWT ...
func NewJWT(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearer = "Bearer "

		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, bearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tok := strings.TrimPrefix(auth, bearer)

		var claims jwt.RegisteredClaims
		parsed, err := jwt.ParseWithClaims(tok, &claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || !parsed.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// ✅ store values in Gin’s context
		c.Set("user", claims.Subject)
		c.Set("claims", claims)
		c.Next()
	}
}

// GenerateDevToken ...
func GenerateDevToken(sub string, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
