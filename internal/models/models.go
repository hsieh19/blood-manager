package models

import "time"

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // 不在JSON中暴露密码
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// BloodPressure 血压记录模型
type BloodPressure struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Systolic   int       `json:"systolic"`   // 收缩压
	Diastolic  int       `json:"diastolic"`  // 舒张压
	HeartRate  int       `json:"heart_rate"` // 心率
	RecordTime time.Time `json:"record_time"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// CreateBPRequest 创建血压记录请求
type CreateBPRequest struct {
	Systolic  int    `json:"systolic" binding:"required"`
	Diastolic int    `json:"diastolic" binding:"required"`
	HeartRate int    `json:"heart_rate"`
	Notes     string `json:"notes"`
}

// BPQueryRequest 血压查询请求
type BPQueryRequest struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// DBConfigRequest 数据库配置请求
type DBConfigRequest struct {
	Type     string `json:"type" binding:"required"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}
