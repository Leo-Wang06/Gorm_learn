package chapter05

import (
	"learn/internal/database"
	"log"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name  string
	Age   int
	Email string
}

type User1 struct {
	gorm.Model

	Name     string    `gorm:"column:name;type:varchar(100);not null;comment:用户名"`
	Age      int       `gorm:"column:age;type:int;default:18;comment:年龄"`
	Email    string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null;comment:邮箱"`
	Phone    string    `gorm:"column:phone;type:varchar(20);index;comment:手机号"`
	Birthday *time.Time `gorm:"column:birthday;type:date;comment:生日"`
	IsActive bool      `gorm:"column:is_active;type:tinyint(1);default:1;comment:是否激活"`
}

func (User1) TableName() string {
	return "users"
}

func LearnCode() {
	db, _ := database.InitDB("05_learn")

	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}

	err2 := db.AutoMigrate(&User1{})
	if err2 != nil {
		log.Fatal(err2)
	}

	log.Println("表结构创建成功！")
}