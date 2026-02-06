package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"blood-manager/internal/config"

	_ "github.com/go-sql-driver/mysql"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

var (
	sqlDB    *sql.DB
	boltDB   *bolt.DB
	dbMux    sync.Mutex
	usingSQL bool
)

// User 用户结构
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// BloodPressure 血压记录结构
type BloodPressure struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Systolic   int       `json:"systolic"`
	Diastolic  int       `json:"diastolic"`
	HeartRate  int       `json:"heart_rate"`
	RecordTime time.Time `json:"record_time"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	usersBucket = []byte("users")
	bpBucket    = []byte("blood_pressure")
	metaBucket  = []byte("meta")
)

// InitDB 初始化数据库
func InitDB() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if cfg.Type == "mysql" {
		return connectMySQL(cfg)
	}
	return connectBolt()
}

// connectBolt 连接Bolt数据库
func connectBolt() error {
	var err error
	boltDB, err = bolt.Open("data/blood_manager.db", 0600, nil)
	if err != nil {
		return err
	}
	usingSQL = false

	// 创建buckets
	err = boltDB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(usersBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(bpBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(metaBucket); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 创建默认管理员
	return createDefaultAdmin()
}

func createDefaultAdmin() error {
	users, _ := GetAllUsers()
	for _, u := range users {
		if u.Role == "admin" {
			return nil
		}
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	return CreateUser("admin", string(hashedPwd), "admin")
}

// connectMySQL 连接MySQL
func connectMySQL(cfg *config.DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	sqlDB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = sqlDB.Ping(); err != nil {
		return err
	}

	usingSQL = true
	return createMySQLTables()
}

func createMySQLTables() error {
	userTable := `CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := sqlDB.Exec(userTable); err != nil {
		return err
	}

	bpTable := `CREATE TABLE IF NOT EXISTS blood_pressure (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL,
		systolic INT NOT NULL,
		diastolic INT NOT NULL,
		heart_rate INT,
		record_time DATETIME NOT NULL,
		notes TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := sqlDB.Exec(bpTable); err != nil {
		return err
	}

	// 创建默认管理员
	var count int
	sqlDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if count == 0 {
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		sqlDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			"admin", string(hashedPwd), "admin")
	}

	settingsTable := `CREATE TABLE IF NOT EXISTS settings (
		description VARCHAR(100),
		setting_key VARCHAR(50) PRIMARY KEY,
		setting_value TEXT
	)`
	if _, err := sqlDB.Exec(settingsTable); err != nil {
		return err
	}

	return nil
}

// TestConnection 测试MySQL连接
func TestConnection(cfg *config.DBConfig) error {
	if cfg.Type != "mysql" {
		return nil
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	testDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer testDB.Close()
	return testDB.Ping()
}

// SwitchDB 切换数据库
func SwitchDB(cfg *config.DBConfig) error {
	dbMux.Lock()
	defer dbMux.Unlock()

	// 关闭旧连接
	if boltDB != nil {
		boltDB.Close()
		boltDB = nil
	}
	if sqlDB != nil {
		sqlDB.Close()
		sqlDB = nil
	}

	var err error
	if cfg.Type == "mysql" {
		err = connectMySQL(cfg)
	} else {
		err = connectBolt()
	}

	if err != nil {
		return err
	}

	return config.SetConfig(cfg)
}

// ========== 用户操作 ==========

func getNextID(tx *bolt.Tx, bucket []byte) int64 {
	b := tx.Bucket(metaBucket)
	key := append(bucket, []byte("_seq")...)
	val := b.Get(key)
	var id int64 = 1
	if val != nil {
		json.Unmarshal(val, &id)
		id++
	}
	data, _ := json.Marshal(id)
	b.Put(key, data)
	return id
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	if usingSQL {
		var user User
		err := sqlDB.QueryRow("SELECT id, username, password, role, created_at FROM users WHERE username = ?",
			username).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	var user *User
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			json.Unmarshal(v, &u)
			if u.Username == username {
				user = &u
				break
			}
		}
		return nil
	})

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetAllUsers 获取所有用户
func GetAllUsers() ([]User, error) {
	if usingSQL {
		rows, err := sqlDB.Query("SELECT id, username, role, created_at FROM users ORDER BY id")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var u User
			rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt)
			users = append(users, u)
		}
		return users, nil
	}

	var users []User
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			json.Unmarshal(v, &u)
			users = append(users, u)
		}
		return nil
	})
	return users, nil
}

// CreateUser 创建用户
func CreateUser(username, hashedPassword, role string) error {
	if usingSQL {
		_, err := sqlDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			username, hashedPassword, role)
		return err
	}

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)

		// 检查用户名是否已存在
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			json.Unmarshal(v, &u)
			if u.Username == username {
				return fmt.Errorf("username already exists")
			}
		}

		id := getNextID(tx, usersBucket)
		user := User{
			ID:        id,
			Username:  username,
			Password:  hashedPassword,
			Role:      role,
			CreatedAt: time.Now(),
		}

		data, _ := json.Marshal(user)
		key := fmt.Sprintf("%d", id)
		return b.Put([]byte(key), data)
	})
}

// DeleteUser 删除用户
func DeleteUser(id int64) error {
	if usingSQL {
		sqlDB.Exec("DELETE FROM blood_pressure WHERE user_id = ?", id)
		_, err := sqlDB.Exec("DELETE FROM users WHERE id = ?", id)
		return err
	}

	// 删除用户的血压记录
	boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bpBucket)
		c := b.Cursor()
		var toDelete [][]byte
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var bp BloodPressure
			json.Unmarshal(v, &bp)
			if bp.UserID == id {
				toDelete = append(toDelete, k)
			}
		}
		for _, k := range toDelete {
			b.Delete(k)
		}
		return nil
	})

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		key := fmt.Sprintf("%d", id)
		return b.Delete([]byte(key))
	})
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(id int64, hashedPassword string) error {
	if usingSQL {
		_, err := sqlDB.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, id)
		return err
	}

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		key := fmt.Sprintf("%d", id)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("user not found")
		}

		var user User
		json.Unmarshal(data, &user)
		user.Password = hashedPassword

		newData, _ := json.Marshal(user)
		return b.Put([]byte(key), newData)
	})
}

// GetUserRole 获取用户角色
func GetUserRole(id int64) string {
	if usingSQL {
		var role string
		sqlDB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
		return role
	}

	var role string
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		key := fmt.Sprintf("%d", id)
		data := b.Get([]byte(key))
		if data != nil {
			var u User
			json.Unmarshal(data, &u)
			role = u.Role
		}
		return nil
	})
	return role
}

// CountAdmins 统计管理员数量
func CountAdmins() int {
	if usingSQL {
		var count int
		sqlDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
		return count
	}

	count := 0
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			json.Unmarshal(v, &u)
			if u.Role == "admin" {
				count++
			}
		}
		return nil
	})
	return count
}

// UpdateUserRole 更新用户角色
func UpdateUserRole(id int64, role string) error {
	if usingSQL {
		_, err := sqlDB.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
		return err
	}

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		key := fmt.Sprintf("%d", id)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("user not found")
		}

		var user User
		json.Unmarshal(data, &user)
		user.Role = role

		newData, _ := json.Marshal(user)
		return b.Put([]byte(key), newData)
	})
}

// ========== 血压记录操作 ==========

// CreateBPRecord 创建血压记录
func CreateBPRecord(userID int64, systolic, diastolic, heartRate int, recordTime time.Time, notes string) (int64, error) {
	if usingSQL {
		result, err := sqlDB.Exec(`INSERT INTO blood_pressure (user_id, systolic, diastolic, heart_rate, record_time, notes) 
			VALUES (?, ?, ?, ?, ?, ?)`, userID, systolic, diastolic, heartRate, recordTime, notes)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	var id int64
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bpBucket)
		id = getNextID(tx, bpBucket)

		bp := BloodPressure{
			ID:         id,
			UserID:     userID,
			Systolic:   systolic,
			Diastolic:  diastolic,
			HeartRate:  heartRate,
			RecordTime: recordTime,
			Notes:      notes,
			CreatedAt:  time.Now(),
		}

		data, _ := json.Marshal(bp)
		key := fmt.Sprintf("%d", id)
		return b.Put([]byte(key), data)
	})

	return id, err
}

// GetBPRecords 获取血压记录
func GetBPRecords(userID int64, startDate, endDate string) ([]BloodPressure, error) {
	if usingSQL {
		query := "SELECT id, systolic, diastolic, heart_rate, record_time, notes FROM blood_pressure WHERE user_id = ?"
		args := []interface{}{userID}

		if startDate != "" {
			query += " AND DATE(record_time) >= ?"
			args = append(args, startDate)
		}
		if endDate != "" {
			query += " AND DATE(record_time) <= ?"
			args = append(args, endDate)
		}
		query += " ORDER BY record_time DESC"

		rows, err := sqlDB.Query(query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var records []BloodPressure
		for rows.Next() {
			var bp BloodPressure
			rows.Scan(&bp.ID, &bp.Systolic, &bp.Diastolic, &bp.HeartRate, &bp.RecordTime, &bp.Notes)
			bp.UserID = userID
			records = append(records, bp)
		}
		return records, nil
	}

	var records []BloodPressure
	boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bpBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var bp BloodPressure
			json.Unmarshal(v, &bp)
			if bp.UserID != userID {
				continue
			}

			dateStr := bp.RecordTime.Format("2006-01-02")
			if startDate != "" && dateStr < startDate {
				continue
			}
			if endDate != "" && dateStr > endDate {
				continue
			}

			records = append(records, bp)
		}
		return nil
	})

	// 按时间倒序
	sort.Slice(records, func(i, j int) bool {
		return records[i].RecordTime.After(records[j].RecordTime)
	})

	return records, nil
}

// DeleteBPRecord 删除血压记录
func DeleteBPRecord(id, userID int64) error {
	if usingSQL {
		result, err := sqlDB.Exec("DELETE FROM blood_pressure WHERE id = ? AND user_id = ?", id, userID)
		if err != nil {
			return err
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			return fmt.Errorf("record not found")
		}
		return nil
	}

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bpBucket)
		key := fmt.Sprintf("%d", id)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("record not found")
		}

		var bp BloodPressure
		json.Unmarshal(data, &bp)
		if bp.UserID != userID {
			return fmt.Errorf("record not found")
		}

		return b.Delete([]byte(key))
	})
}

// ========== 数据库备份还原 ==========

// BackupDB 备份数据库到指定路径
func BackupDB(destPath string) error {
	if usingSQL {
		return fmt.Errorf("MySQL数据库请使用mysqldump进行备份")
	}

	if boltDB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("无法创建备份文件: %v", err)
	}
	defer destFile.Close()

	return boltDB.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(destFile)
		return err
	})
}

// RestoreDB 从指定路径还原数据库
func RestoreDB(srcPath string) error {
	if usingSQL {
		return fmt.Errorf("MySQL数据库请使用mysql命令进行还原")
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", srcPath)
	}

	dbMux.Lock()
	defer dbMux.Unlock()

	if boltDB != nil {
		boltDB.Close()
		boltDB = nil
	}

	dbPath := "data/blood_manager.db"

	srcFile, err := os.Open(srcPath)
	if err != nil {
		connectBolt()
		return fmt.Errorf("无法打开备份文件: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dbPath)
	if err != nil {
		connectBolt()
		return fmt.Errorf("无法创建数据库文件: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		connectBolt()
		return fmt.Errorf("复制文件失败: %v", err)
	}

	destFile.Close()
	return connectBolt()
}

// GetSetting 获取全局设置
func GetSetting(key string) (string, error) {
	dbMux.Lock()
	defer dbMux.Unlock()

	if usingSQL {
		var value string
		err := sqlDB.QueryRow("SELECT setting_value FROM settings WHERE setting_key = ?", key).Scan(&value)
		if err == sql.ErrNoRows {
			return "", nil
		}
		return value, err
	}

	var value string
	err := boltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucket)
		v := b.Get([]byte(key))
		if v != nil {
			value = string(v)
		}
		return nil
	})
	return value, err
}

// SetSetting 保存全局设置
func SetSetting(key, value string) error {
	dbMux.Lock()
	defer dbMux.Unlock()

	if usingSQL {
		_, err := sqlDB.Exec("INSERT INTO settings (setting_key, setting_value) VALUES (?, ?) ON DUPLICATE KEY UPDATE setting_value = ?", key, value, value)
		return err
	}

	return boltDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucket)
		return b.Put([]byte(key), []byte(value))
	})
}
