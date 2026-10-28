package chapter19

import (
	"fmt"
	"learn/internal/database"
	"log"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
	Age int
}

type Goods struct {
	Name string
	Number uint
	Status string
}

func WithStatus(status string) func(db *gorm.DB) *gorm.DB {
	return  func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

func WithDateRange(startTime time.Time, endTime time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !startTime.IsZero() {
			db = db.Where("created_at > ?", startTime)
		}
		if !endTime.IsZero() {
			db = db.Where("created_at < ?", endTime)
		}
		return db
	}
}

func LearnCode() {
	db, _ := database.InitDB("19_learn");

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	goods := []Goods{{Name: "铅笔", Number: 100, Status: "充足"},{Name: "钢笔", Number: 10, Status: "紧缺"}}

	if err := db.Create(&goods); err != nil {
		log.Fatal("插入数据失败：", err)
	}
	
	db.Scopes(WithStatus("充足")).Find(&goods)

	fmt.Println(goods[0].Name)

	var users []User

	start := time.Date(2026, 10, 28, 0, 0, 0, 0, time.Local)
	end   := time.Date(2026, 10, 28, 23, 59, 59, 0, time.Local)

	db.Scopes(WithDateRange(start, end)).Find(&users)

}