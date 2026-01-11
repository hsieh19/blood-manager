package config

import (
	"encoding/json"
	"os"
	"sync"
)

// DBConfig 数据库配置
type DBConfig struct {
	Type     string `json:"type"`     // sqlite 或 mysql
	Host     string `json:"host"`     // MySQL主机
	Port     string `json:"port"`     // MySQL端口
	User     string `json:"user"`     // MySQL用户名
	Password string `json:"password"` // MySQL密码
	DBName   string `json:"dbname"`   // 数据库名
}

var (
	currentConfig *DBConfig
	configMutex   sync.RWMutex
	configFile    = "db_config.json"
)

// GetConfig 获取当前数据库配置
func GetConfig() *DBConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if currentConfig == nil {
		return &DBConfig{Type: "sqlite"}
	}
	return currentConfig
}

// SetConfig 设置数据库配置
func SetConfig(cfg *DBConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()
	currentConfig = cfg
	return saveConfig(cfg)
}

// LoadConfig 从文件加载配置
func LoadConfig() (*DBConfig, error) {
	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			currentConfig = &DBConfig{Type: "sqlite"}
			return currentConfig, nil
		}
		return nil, err
	}

	var cfg DBConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	currentConfig = &cfg
	return currentConfig, nil
}

// saveConfig 保存配置到文件
func saveConfig(cfg *DBConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}
