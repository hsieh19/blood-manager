package middleware

import (
	"net/http"
	"strconv"
	"time"

	"health-manager/internal/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthRequired 验证用户是否登录
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 核心：后端自动退出逻辑检查
		idleTimeoutStr, _ := database.GetSetting("idle_timeout")
		idleTimeout, _ := strconv.Atoi(idleTimeoutStr)

		if idleTimeout > 0 {
			lastActivity := session.Get("last_activity")
			now := time.Now().Unix()

			if lastActivity != nil {
				lastTime := lastActivity.(int64)
				if now-lastTime > int64(idleTimeout*60) {
					// 超过设定的自动退出时间
					session.Clear()
					session.Save()
					c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已超时，请重新登录"})
					c.Abort()
					return
				}
			}
			// 更新最后活跃时间，并延长 Cookie 有效期
			session.Set("last_activity", now)
			session.Save()
		}

		c.Set("user_id", userID)
		c.Set("username", session.Get("username"))
		c.Set("role", session.Get("role"))
		c.Next()
	}
}

// AdminRequired 验证用户是否为管理员
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
