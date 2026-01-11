package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"blood-manager/internal/database"
	"blood-manager/internal/handlers"
	"blood-manager/internal/middleware"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置日志输出到文件
	f, _ := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	r := gin.Default()

	// 配置Session
	store := cookie.NewStore([]byte("blood-manager-secret-key-2024"))
	r.Use(sessions.Sessions("session", store))

	// 静态文件服务
	r.Static("/static", "./web/static")

	// 根路径重定向到登录页
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/static/pages/login.html")
	})

	// 公开API
	r.POST("/api/login", handlers.Login)
	r.POST("/api/logout", handlers.Logout)
	r.GET("/api/me", handlers.GetCurrentUser)

	// 用户API (需要登录)
	userAPI := r.Group("/api")
	userAPI.Use(middleware.AuthRequired())
	{
		userAPI.GET("/bp", handlers.GetBPRecords)
		userAPI.POST("/bp", handlers.CreateBP)
		userAPI.DELETE("/bp/:id", handlers.DeleteBP)
	}

	// 管理员API (需要管理员权限)
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		adminAPI.GET("/users", handlers.GetUsers)
		adminAPI.POST("/users", handlers.CreateUser)
		adminAPI.DELETE("/users/:id", handlers.DeleteUser)
		adminAPI.PUT("/users/:id/password", handlers.ChangeUserPassword)
		adminAPI.PUT("/users/:id/role", handlers.ToggleAdminRole)
		adminAPI.GET("/db-config", handlers.GetDBConfig)
		adminAPI.POST("/db-config", handlers.SaveDBConfig)
		adminAPI.POST("/db-config/test", handlers.TestDBConfig)
		adminAPI.POST("/db/backup", handlers.BackupDatabase)
		adminAPI.POST("/db/restore", handlers.RestoreDatabase)
		adminAPI.POST("/settings/idle-timeout", handlers.SetIdleTimeout)
	}

	// 通用设置API (需要登录)
	userAPI.GET("/settings/idle-timeout", handlers.GetIdleTimeout)

	log.Println("血压管理系统启动在 http://localhost:8080")
	r.Run(":8080")
}
