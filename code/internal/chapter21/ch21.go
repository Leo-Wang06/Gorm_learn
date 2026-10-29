package chapter21

import (
	"fmt"
	"learn/internal/database"
	"log"
)

type Company struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

type User struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"type:varchar(100);not null"`
	CompanyID uint    // 外键
	Company   Company // BelongsTo 关联
}

func LearnCode() {
	db, _ := database.InitDB("19_learn")

	if err := db.AutoMigrate(&Company{}, &User{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	// 1. 创建公司
	google := Company{Name: "Google"}
	apple := Company{Name: "Apple"}
	db.Create(&google)
	db.Create(&apple)

	// 2. 创建用户，分配到公司
	db.Create(&User{Name: "张三", CompanyID: google.ID})
	db.Create(&User{Name: "李四", CompanyID: apple.ID})
	db.Create(&User{Name: "王五", CompanyID: google.ID})

	var users []User
	db.Preload("Company").Find(&users)
	for _, u := range users {
		fmt.Printf("%s 在 %s 工作\n", u.Name, u.Company.Name)
	}

	var gooleUsers []User
	db.Joins("Company").Where("Company.name = ?", "Google").Find(&gooleUsers)
	fmt.Printf("Google 有 %d 名员工\n", len(gooleUsers))

	Pracice(db)
}