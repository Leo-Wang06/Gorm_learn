package chapter07

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type User3 struct {
	Name     string
	Settings map[string]interface{} `gorm:"serializer:json"`
	Status UserStatus `gorm:"type:varchar(20);default:active;commen:用户状态"`
}

type Timestamp struct {
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdateAt time.Time `gorm:"autoUpdateTime"`
}

type Product struct {
	Name string `gorm:"type:varchar(200)"`
	Timestamp
}

type UserStatus string

const (
	StatusActive UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned UserStatus = "banned"
)

func Practice(db *gorm.DB) {
	if err := db.AutoMigrate(&User3{}); err != nil {
		log.Fatal(err)
	}

	user3 := User3{
		Name: "小明",
		Settings: map[string]interface{}{"type": 23},
	}

	if err := db.Create(&user3).Error; err != nil {
		log.Fatal(err)
	}
}