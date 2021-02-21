package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type claims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	Timestamp int    `json:"timestamp"`
	ExpiredAt int    `json:"expired_at"`
}

func VerifyToken(c *gin.Context) {
	tokenRaw := c.Request.Header.Get("Authorization")
	bearerToken := strings.Split(tokenRaw, " ")

	if len(bearerToken) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "bearer not exist", "status": false})
		c.Abort()
		return
	}

	if bearerToken[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "error: unauthorized", "status": false})
		c.Abort()
		return
	}

	tokenString := bearerToken[1]
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod("HS512") != t.Method {
			return nil, errors.New("signing method is not matched")
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "status": false})
		c.Abort()
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "error: token not valid", "status": false})
		c.Abort()
		return
	}

	claimsByte, _ := json.Marshal(token.Claims)

	var claim claims
	json.Unmarshal(claimsByte, &claim)

	if claim.ExpiredAt < int(time.Now().Unix()) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "error: token has expired", "status": false})
		c.Abort()
		return
	}

	c.Next()
}
