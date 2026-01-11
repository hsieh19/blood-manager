package handlers

import (
	"fmt"
	"net/http"

	"blood-manager/config"
	"blood-manager/database"
	"blood-manager/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	users, err := database.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整的用户信息"})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码处理失败"})
		return
	}

	if err := database.CreateUser(req.Username, string(hashedPwd), "user"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户创建成功"})
}

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	role := database.GetUserRole(id)
	if role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员账号"})
		return
	}

	if err := database.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ChangeUserPassword 修改用户密码
func ChangeUserPassword(c *gin.Context) {
	userID := c.Param("id")
	var id int64
	fmt.Sscanf(userID, "%d", &id)

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入新密码"})
		return
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码处理失败"})
		return
	}

	if err := database.UpdateUserPassword(id, string(hashedPwd)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// GetDBConfig 获取数据库配置
func GetDBConfig(c *gin.Context) {
	cfg := config.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"type":   cfg.Type,
		"host":   cfg.Host,
		"port":   cfg.Port,
		"user":   cfg.User,
		"dbname": cfg.DBName,
	})
}

// SaveDBConfig 保存数据库配置
func SaveDBConfig(c *gin.Context) {
	var req models.DBConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置信息不完整"})
		return
	}

	cfg := &config.DBConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		DBName:   req.DBName,
	}

	if err := database.SwitchDB(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据库连接失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "数据库配置已保存并切换成功"})
}

// TestDBConfig 测试数据库连接
func TestDBConfig(c *gin.Context) {
	var req models.DBConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置信息不完整"})
		return
	}

	cfg := &config.DBConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		DBName:   req.DBName,
	}

	if err := database.TestConnection(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "连接失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "连接成功"})
}

// BackupDatabase 备份数据库
func BackupDatabase(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定备份路径"})
		return
	}

	if err := database.BackupDB(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "备份失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "备份成功", "path": req.Path})
}

// RestoreDatabase 还原数据库
func RestoreDatabase(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请指定还原文件路径"})
		return
	}

	if err := database.RestoreDB(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "还原失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "还原成功，请重新登录"})
}

// GetIdleTimeout 获取全局自动退出时间
func GetIdleTimeout(c *gin.Context) {
	value, err := database.GetSetting("idle_timeout")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}
	if value == "" {
		value = "0"
	}
	c.JSON(http.StatusOK, gin.H{"idle_timeout": value})
}

// SetIdleTimeout 设置全局自动退出时间
func SetIdleTimeout(c *gin.Context) {
	var req struct {
		Timeout int `json:"timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的参数"})
		return
	}

	if req.Timeout < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "时间必须大于等于0"})
		return
	}

	if err := database.SetSetting("idle_timeout", fmt.Sprintf("%d", req.Timeout)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置已保存"})
}
