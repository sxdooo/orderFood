package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orderfood/server/internal/model"
	"github.com/orderfood/server/internal/pkg/response"
	"github.com/orderfood/server/internal/service"
)

const ContextUserKey = "authUser"

func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		user, err := auth.ParseToken(c.Request.Context(), token)
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func RequireRole(role model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ContextUserKey)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}
		user := raw.(*model.User)
		if user.Role != role {
			response.Forbidden(c, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func GetUser(c *gin.Context) *model.User {
	raw, _ := c.Get(ContextUserKey)
	if raw == nil {
		return nil
	}
	return raw.(*model.User)
}
