package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Config struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	User  string `json:"modifyUser"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func Start() {

	// 确保 data 目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		panic("failed to create data directory: " + err.Error())
	}

	// 初始化数据库
	var err error
	DB, err = gorm.Open(sqlite.Open("data/config.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(&Config{})
	if err != nil {
		panic("failed to migrate database")
	}
}

// TableName 设置表名
func (Config) TableName() string {
	return "config"
}

func Get(name string) Config {
	var cfg Config
	if err := DB.Where("name = ?", name).First(&cfg).Error; err != nil {
		log.Printf("❌ [Config] 查询配置 %s 失败: %v", name, err)
		return cfg
	}
	return cfg
}

func Set(name, value, user string) error {
	if len(user) == 0 {
		return errors.New("❌ [Config] 用户名不能为空")
	}
	var cfg Config
	err := DB.Where("name = ?", name).First(&cfg).Error
	if err != nil {
		// 不存在则创建
		cfg = Config{
			Name:  name,
			Value: value,
			User:  user,
		}
		if err := DB.Create(&cfg).Error; err != nil {
			return fmt.Errorf("❌ [Config] 创建配置 %s 失败: %v", name, err)
		}
		return nil
	}
	// 存在则修改
	cfg.Value = value
	cfg.User = user
	if err := DB.Save(&cfg).Error; err != nil {
		return fmt.Errorf("❌ [Config] 保存配置 %s 失败: %v", name, err)
	}
	return nil
}
