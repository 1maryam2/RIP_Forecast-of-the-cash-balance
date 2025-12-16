package middleware

import (
	"lab_1/internal/app/auth"
	"lab_1/internal/app/redis"
	"lab_1/internal/app/role"
	"strings"

	"github.com/gin-gonic/gin"
)

const jwtPrefix = "Bearer "

func AuthMiddleware(jwtService *auth.JWTService, redisClient *redis.Client, requiredRoles ...role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr := c.GetHeader("Authorization")
		if jwtStr == "" {
			c.AbortWithStatus(401)
			return
		}
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			c.AbortWithStatus(401)
			return
		}

		jwtStr = jwtStr[len(jwtPrefix):]

		err := redisClient.CheckJWTInBlacklist(c.Request.Context(), jwtStr)
		if err == nil {
			c.AbortWithStatus(401)
			return
		}

		claims, err := jwtService.ValidateToken(jwtStr)
		if err != nil {
			c.AbortWithStatus(401)
			return
		}

		if len(requiredRoles) > 0 {
			hasRequiredRole := false
			for _, requiredRole := range requiredRoles {
				if claims.Role == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if !hasRequiredRole {
				c.AbortWithStatus(403)
				return
			}
		}

		c.Set("userUUID", claims.UserUUID.String())
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func RoleMiddleware(requiredRoles ...role.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.AbortWithStatus(403)
			return
		}

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if userRole.(role.Role) == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.AbortWithStatus(403)
			return
		}

		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) (uint, error) {
	if id, ok := c.Get("userID"); ok {
		return id.(uint), nil
	}
	return 0, gin.Error{}
}

func GetUserRoleFromContext(c *gin.Context) (role.Role, error) {
	if roleVal, ok := c.Get("userRole"); ok {
		return roleVal.(role.Role), nil
	}
	return role.Creator, gin.Error{}
}

func IsModeratorFromContext(c *gin.Context) bool {
	userRole, err := GetUserRoleFromContext(c)
	if err != nil {
		return false
	}
	return userRole == role.Manager || userRole == role.Admin
}
