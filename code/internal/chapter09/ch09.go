package chapter09

import (
	"fmt"
	"learn/internal/database"
	"log"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
	Age int
}

func LearnCode() {
	db, _ := database.InitDB("09_Learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("创建表失败：", err)
	}
	
	user := User{Name: "张三", Age: 25}
	fmt.Println("创建前ID: ", user.ID)
	result := db.Create(&user)

	fmt.Println("插入ID: ", user.ID)
	fmt.Println("影响行数：", result.RowsAffected)
	fmt.Println("影响：", result.Error)

	users := []User{
		{Name: "张三", Age: 25},
		{Name: "李四", Age: 30},
		{Name: "王五", Age: 22},
	}
	db.Create(&users)

}