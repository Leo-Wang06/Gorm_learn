package chapter14

import (
	"fmt"
	"learn/internal/database"
	"log"
)

type User struct {
	ID   int
	Name string
	Age  int
}

func LearnCode() {
	db, _ := database.InitDB("14_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("建表失败: ", err)
	}

	users := []User{{Name: "张三", Age: 21},{Name: "李四", Age: 27}}
	db.Create(&users)

	db.Raw("SELECT * FROM users WHERE status = ?", "active").Scan(&users)
	fmt.Println(users[0])

}