package chapter15

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
	db, _ := database.InitDB("15_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	users := []User{{Name: "张三", Age: 21},{Name: "李四", Age: 28}}
	if result := db.Create(users); result.Error != nil {
		log.Fatal("插入表数据失败：", result.Error)
	}

	db.Order("age desc").Find(&users)
	fmt.Println(users[0].Name)

	db.Order("age desc, name asc").Find(&users)
	fmt.Println(users[0].ID)

	var total int64

	page := 2
	pageSize := 10
	baseQuery := db.Model(&User{}).Where("age > ?", 18)

	if err := baseQuery.Count(&total).Error; err != nil {
		log.Fatal("查询总数失败：", err)
	}

	if err := baseQuery.Scopes(Paginate(page, pageSize)).Order("id DESC").Find(&users).Error; err != nil {
		log.Fatal("分页查询失败：", err)
	}

	fmt.Printf("第 %d 页 / 共 %d 条, 每页 %d 条\n", page, total, pageSize)

	db.Order("age desc, name asc").Find(&users)

	var lastID int

	db.Where("id > ?", lastID).Order("id desc").Limit(10)
}