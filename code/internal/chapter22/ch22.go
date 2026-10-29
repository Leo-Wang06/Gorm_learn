package chapter22

import (
	"fmt"
	"learn/internal/database"
	"log"
)

type User struct {
	ID uint `gorm:"primaryKey"`
	Name string
	CreditCard CreditCard
}

type CreditCard struct {
	ID uint `gorm:"primaryKey"`
	Number string
	UserID uint
}

func LearnCode() {
	db, _ := database.InitDB("22_learn")

	if err := db.AutoMigrate(&User{}, &CreditCard{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	user := User{
		Name: "张三",
		CreditCard: CreditCard{
			Number: "4111-1111-1111-1111",
		},
	}

	// Create 默认就会连同 HasOne 关联一起插入；不要写 Select("CreditCard")，
	// 那会把 users 表的写入列也限制成"只选 CreditCard"，导致 users.name 根本没被写入（NULL）
	db.Create(&user)

	// 注意：user 已经被 Create 填上了主键 ID=3，直接拿它当查询目标会多带一个 id = 3 条件
	var found User
	db.Preload("CreditCard").First(&found, 1)
	fmt.Println(found.CreditCard.Number)

	// db.Joins("CreditCard").First(&user, 1)
}