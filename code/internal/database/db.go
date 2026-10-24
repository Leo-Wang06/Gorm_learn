package database

import (
	"fmt"
	"log"
	// "os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbUser = "root"
	dbPass = "root123456"
	dbAddr = "127.0.0.1:3306"
)

func DSN(dbName string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbAddr, dbName,
	)
}

func InitDB(dbName string) (*gorm.DB, error) {
	dsn := DSN(dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,  // 关闭默认事务，提升性能
		PrepareStmt:            false, // 开发时不缓存，方便调试
		Logger: logger.Default.LogMode(logger.Info),
		// Logger: logger.New(
		// 	log.New(os.Stdout, "\r\n", log.LstdFlags),
		// 	logger.Config{
		// 		SlowThreshold:             200 * time.Millisecond,
		// 		LogLevel:                  logger.Info,  // 开发：打印所有 SQL
		// 		IgnoreRecordNotFoundError: false,
		// 		Colorful:                  true,
		// 	},
		// ),
	})
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

	return db, nil
}
