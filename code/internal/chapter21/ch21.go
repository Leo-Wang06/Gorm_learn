package chapter21

import (
	"learn/internal/database"
	"log"
)

type Company struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type User struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CompanyID uint
	Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
}

func LearnCode() {
	db, _ := database.InitDB("19_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	company := Company{Name: "Google"}
	db.Create(&company)

	user := User{Name: "张三", CompanyID: company.ID}
	db.Create(&user)

}