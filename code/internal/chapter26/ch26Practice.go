package chapter26

import (
	"fmt"

	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primaryKey"`
	Name string
	Orders []Order
}

type Order struct {
	ID     uint `gorm:"primaryKey"`
	Amount float64
	// 用 *uint：nil 表示"还没分配给任何用户"，写进库里是 NULL。
	// 如果写成 uint，零值 0 会被当成外键值写进 user_id=0，而 0 不是任何用户的 id，
	// 会被 orders.user_id 上的外键约束拒绝（Error 1452）。
	UserID *uint
}

func PracticeCode(db *gorm.DB) {
	db.AutoMigrate(&User{}, &Order{})

	user := User{Name: "张三"}
	db.Create(&user)

	order1 := &Order{Amount: 100}
	order2 := &Order{Amount: 200}
	order3 := &Order{Amount: 300}
	db.Create(order1)
	db.Create(order2)
	db.Create(order3)

	fmt.Println("===Append===")
	db.Model(&user).Association("Orders").Append(order1, order2)
	count := db.Model(&user).Association("Orders").Count()
	fmt.Printf("Append 后订单数：%d\n", count)

	fmt.Println("\n=== Find ===")
	var orders []Order
	db.Model(&user).Association("Orders").Find(&orders)
	for _, o := range orders {
		fmt.Printf(" 订单#%d: %.2f元\n", o.ID, o.Amount)
	}

	fmt.Println("\n=== Replace ===")
	db.Model(&user).Association("Orders").Replace([]Order{*order3})
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Replace 后订单数: %d\n", count)

	fmt.Println("\n=== Delete ===")
	db.Model(&user).Association("Orders").Delete(order3)
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Delete 后订单数: %d\n", count)

	fmt.Println("\n=== Clear ===")
	db.Model(&user).Association("Orders").Append(order1, order2, order3)
	db.Model(&user).Association("Orders").Clear()
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Clear 后订单数: %d\n", count)
}