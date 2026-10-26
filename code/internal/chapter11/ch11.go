package chapter11

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
	db, _ := database.InitDB("11_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("创建表失败: ", err)
	}

	users := []User{{Name: "张三", Age: 15},{Name: "王五", Age: 16}}
	db.Create(&users)

	result1 := db.Where("name = ?", "张三").Find(&users)
	if result1.Error != nil {
		log.Fatal("查询失败：", result1.Error)
	}
	fmt.Printf("%+v\n",users)

	db.Where("(name = ? OR name 	= ?) AND age > ?", "张三", "王五", 14).Find(&users)

	users = SearchUsers(db, "张三", 14)
	fmt.Println(users[0])

	db.Where("name = ? AND age > ?").Find(&users)
	db.Where(map[string]interface{}{"name": "张三"}).Where("age > ?", 18).Find(&users)
	db.Not("name = ?", "张三").Not("name = ?", "李四").Find(&users)
}