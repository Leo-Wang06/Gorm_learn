package chapter25

import (
	"fmt"
	"learn/internal/database"
	"log"

	"gorm.io/gorm"
)

// Company 公司
type Company struct {
	gorm.Model
	Name  string
	City  string
	Users []User // 一对多：一个公司有多个用户
}

// User 用户
type User struct {
	gorm.Model
	Name      string
	CompanyID uint
	Company   Company // 属于一个公司
	Orders    []Order // 一对多：一个用户有多个订单
}

// Order 订单
type Order struct {
	gorm.Model
	UserID uint
	Status string // paid / pending
	Amount float64
}

// 练习数据（每次运行都会重置）：
//
// 公司：云图科技(北京) / 海联科技(上海) / 南风科技(广州)
// 用户：张三、李四(云图)  王五、赵六(海联)  钱七、孙八(南风)
// 订单：
//   张三：paid 150、paid 90、pending 200
//   李四：paid 300
//   王五：paid 120、pending 50
//   赵六：无订单
//   钱七：paid 80
//   孙八：paid 500、paid 150
//
// 练习提示：
//   1. N+1：先 Find 所有用户，再循环查每个用户的订单；然后改成 Preload("Orders")
//   2. 三层嵌套：Preload("Users.Orders").Find(&companies)
//   3. 条件预加载：Preload("Orders", "status = ? AND amount > ?", "paid", 100)
//   4. Joins：db.Joins("Company").Where("Company.city = ?", "上海").Find(&users)

func LearnCode() {
	db, err := database.InitDB("25_learn")
	if err != nil {
		log.Fatal("连接数据库失败：", err)
	}

	if err := db.AutoMigrate(&Company{}, &User{}, &Order{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	seed(db)

	// ===== 练习区：在下面写本章练习题 =====
	var users []User
	db.Find(&users)

	for _, u := range users {
		var orders []Order
		db.Where("user_id = ?", u.ID).Find(&orders)
		fmt.Printf("%s 有 %d 个订单\n", u.Name, len(orders))
	}

	db.Preload("Orders").Find(&users)

	db.Preload("Company.Users.Orders").Find(&users)

	db.Preload("Orders", "amount > ? AND status = ?", 100, "paid").Find(&users)
	fmt.Println(users[0].Name)

	db.Joins("Company").Where("Company.city = ?", "上海").Find(&users)
	fmt.Println(users[0].Name)

	// 示例：预加载公司 + 订单
	// var users []User
	// if err := db.Preload("Company").Preload("Orders").Find(&users).Error; err != nil {
	// 	log.Fatal("查询失败：", err)
	// }
	// for _, u := range users {
	// 	fmt.Printf("%s @ %s，订单数 %d\n", u.Name, u.Company.Name, len(u.Orders))
	// }
}

// seed 重置并写入练习数据
func seed(db *gorm.DB) {
	// 有外键约束，按「子表 → 父表」的顺序硬删除
	for _, model := range []interface{}{&Order{}, &User{}, &Company{}} {
		if err := db.Unscoped().Where("1 = 1").Delete(model).Error; err != nil {
			log.Fatal("清理旧数据失败：", err)
		}
	}

	companies := []Company{
		{Name: "云图科技", City: "北京"},
		{Name: "海联科技", City: "上海"},
		{Name: "南风科技", City: "广州"},
	}
	if err := db.Create(&companies).Error; err != nil {
		log.Fatal("初始化公司失败：", err)
	}

	users := []User{
		{Name: "张三", CompanyID: companies[0].ID},
		{Name: "李四", CompanyID: companies[0].ID},
		{Name: "王五", CompanyID: companies[1].ID},
		{Name: "赵六", CompanyID: companies[1].ID},
		{Name: "钱七", CompanyID: companies[2].ID},
		{Name: "孙八", CompanyID: companies[2].ID},
	}
	if err := db.Create(&users).Error; err != nil {
		log.Fatal("初始化用户失败：", err)
	}

	orders := []Order{
		{UserID: users[0].ID, Status: "paid", Amount: 150},    // 张三
		{UserID: users[0].ID, Status: "paid", Amount: 90},     // 张三
		{UserID: users[0].ID, Status: "pending", Amount: 200}, // 张三
		{UserID: users[1].ID, Status: "paid", Amount: 300},    // 李四
		{UserID: users[2].ID, Status: "paid", Amount: 120},    // 王五
		{UserID: users[2].ID, Status: "pending", Amount: 50},  // 王五
		{UserID: users[4].ID, Status: "paid", Amount: 80},     // 钱七
		{UserID: users[5].ID, Status: "paid", Amount: 500},    // 孙八
		{UserID: users[5].ID, Status: "paid", Amount: 150},    // 孙八
	}
	if err := db.Create(&orders).Error; err != nil {
		log.Fatal("初始化订单失败：", err)
	}
}
