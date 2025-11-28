package middleware

import (
	"fmt"
	"gintugas/modules/components/Auth/blacklist"
	utils "gintugas/modules/components/Auth/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Dapatkan blacklist instance
		bl := blacklist.GetInstance()

		// Debug log (bisa dihapus di production)
		fmt.Printf("[MIDDLEWARE] Checking token. Total in blacklist: %d\n", bl.GetCount())
		fmt.Printf("[MIDDLEWARE] Token: %s\n", tokenString[:20]+"...")

		// CEK APAKAH TOKEN ADA DI BLACKLIST (SUDAH LOGOUT)
		if bl.IsBlacklisted(tokenString) {
			fmt.Printf("[MIDDLEWARE] Token IS BLACKLISTED - REJECTING\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		fmt.Printf("[MIDDLEWARE] Token NOT in blacklist - ALLOWING\n")

		// Cek validitas token dengan utils
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// User ID get claim
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role in token"})
			c.Abort()
			return
		}

		// Set info ke context
		c.Set("user_id", userID)
		c.Set("username", claims["username"])
		c.Set("user_role", role)

		c.Next()
	}
}
