package gorm

import (
	"fmt"
	"hxzl/config"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Run() {
	dbType := config.Get("db_type").Value
	if dbType == "" {
		log.Fatal("❌ [GROM] 未找到 db_type 配置")
		return
	}

	var dialector gorm.Dialector

	switch dbType {
	case "sqlite", "sqlite3":
		path := config.Get("db_path").Value
		dialector = sqlite.Open(path)

	case "pgsql", "postgres", "postgresql":
		host := config.Get("db_host").Value
		port := config.Get("db_port").Value
		user := config.Get("db_user").Value
		password := config.Get("db_password").Value
		name := config.Get("db_name").Value
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
		dialector = postgres.Open(dsn)

	case "mysql", "mariadb":
		host := config.Get("db_host").Value
		port := config.Get("db_port").Value
		user := config.Get("db_user").Value
		password := config.Get("db_password").Value
		name := config.Get("db_name").Value
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)
		dialector = mysql.Open(dsn)

	default:
		log.Fatalf("⚠️ [GROM] 数据库模块 配置错误！无法识别的数据库类型: %s", dbType)
		return
	}

	// 初始化 GORM 实例
	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Printf("❌ [GORM] 初始化数据库 %s 失败: %v", dbType, err)
	}

	log.Printf("✅ [GORM] 已成功初始化数据库 %s", dbType)
}
