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

// BloodPressure 健康记录模型（包含血压和身高体重）
type BloodPressure struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Systolic   int       `json:"systolic"`   // 收缩压 (mmHg)
	Diastolic  int       `json:"diastolic"`  // 舒张压 (mmHg)
	HeartRate  int       `json:"heart_rate"` // 心率 (次/分)
	Height     float64   `json:"height"`     // 身高 (cm)
	Weight     float64   `json:"weight"`     // 体重 (kg)
	Waistline  float64   `json:"waistline"`  // 腰围 (cm)
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

// CreateBPRequest 创建健康记录请求
type CreateBPRequest struct {
	Systolic  int     `json:"systolic"`   // 收缩压（可选）
	Diastolic int     `json:"diastolic"`  // 舒张压（可选）
	HeartRate int     `json:"heart_rate"` // 心率（可选）
	Height    float64 `json:"height"`     // 身高（可选）
	Weight    float64 `json:"weight"`     // 体重（可选）
	Waistline float64 `json:"waistline"`  // 腰围（可选）
	Notes     string  `json:"notes"`      // 备注
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
