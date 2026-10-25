package chapter10

import (
	"errors"
	"fmt"
	"learn/internal/database"
	"log"

	"gorm.io/gorm"
)

type User struct {
	ID   int
	Name string
}

var ErrUserNotFound = errors.New("用户不存在")

func FirstOrCreate(db *gorm.DB,user User ,ID int) (*User, error) {
	err := db.First(&user, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func LearnCode() {
	db, _ := database.InitDB("10_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	user := User{ID: 1, Name: "小明"}
	db.Create(&user)

	result := db.First(&user)
	fmt.Println(result)
	fmt.Println(db.Last(&user))
	fmt.Println(db.Take(&user))
	if err := db.First(&user, 2).Error; err != nil {
		log.Fatal("查询失败",err)
	}

	user1 := []User{{ID: 1, Name: "小黄"},{ID: 2, Name: "小红"}}

	if err := db.Find(&user1, []int{2 ,3}).Error; err != nil {
		log.Fatal("Find查询失败：",err)
	}
	
}