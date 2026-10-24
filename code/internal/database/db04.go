package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB04() *gorm.DB {
	dsn := "root:root123456@tcp(127.0.0.1:3306)/04_learn?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败：", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取 sql.DB 失败：", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(25)                 // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)                 // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库 Ping 失败：", err)
	}

	return db
}
