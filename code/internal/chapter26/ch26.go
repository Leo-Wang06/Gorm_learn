package chapter26

import (
	// "fmt"
	"learn/internal/database"
	// "log"
)

// type User struct {
// 	ID     uint
// 	Name   string
// 	Orders []Order // HasMany：外键在 orders 表（orders.user_id）
// }

// type Order struct {
// 	ID     uint
// 	Amount uint
// 	UserID uint
// }

func LearnCode() {
	db, _ := database.InitDB("26_learn")

	// if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
	// 	log.Fatal("建表失败:", err)
	// }

	// // 一次 Create 就能同时插入用户和它的订单，GORM 会自动回填 orders.user_id
	// user := User{
	// 	Name:   "张三",
	// 	Orders: []Order{{Amount: 250}},
	// }
	// if err := db.Create(&user).Error; err != nil {
	// 	log.Fatal("创建失败:", err)
	// }

	// var loaded User
	// if err := db.Preload("Orders").First(&loaded, user.ID).Error; err != nil {
	// 	log.Fatal("查询失败:", err)
	// }
	// log.Printf("%s 有 %d 个订单", loaded.Name, len(loaded.Orders))

	// var user1 User
	// // 不 Preload 的话 user1.Orders 是 nil，后面一取下标就越界
	// if err := db.Preload("Orders").First(&user1, 1).Error; err != nil {
	// 	log.Fatal("查询失败:", err)
	// }

	// order := Order{Amount: 100}
	// // 注意：Association 的方法直接返回 error，没有 .Error 字段
	// if err := db.Model(&user1).Association("Orders").Append(&order); err != nil {
	// 	log.Fatal("追加订单失败:", err)
	// }

	// // Append 只写数据库，不会更新内存里的 user1.Orders，需要重新加载
	// if err := db.Preload("Orders").First(&user1, user1.ID).Error; err != nil {
	// 	log.Fatal("重新查询失败:", err)
	// }
	// for _, o := range user1.Orders {
	// 	fmt.Println(o.ID, o.Amount)
	// }

	PracticeCode(db)
}
