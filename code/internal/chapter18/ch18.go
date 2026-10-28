package chapter18

import (
	"database/sql"
	"fmt"
	"learn/internal/database"
	"log"

	"gorm.io/gorm"
)

// User 练习用用户模型
type User struct {
	gorm.Model
	Name string
	Age  int
	City string
}

// Order 练习用订单模型
type Order struct {
	gorm.Model
	UserID uint
	Amount float64
}

// 练习数据概览（每次运行都会重置成这份数据）：
//
// 用户：
//   张三  21岁  北京    订单总额 600
//   李四  28岁  北京    订单总额  50
//   王五  17岁  上海    订单总额   0（没有订单）
//   赵六  35岁  上海    订单总额 1500
//   钱七  50岁  广州    订单总额 200
//   孙八  60岁  广州    订单总额 2000
//
// 提示：有订单的用户总额平均是 (600+50+1500+200+2000)/5 = 870；
//       把王五(0)也算进去则是 4350/6 = 725。练习题 1 注意自己选的口径。
//       练习题 2 的“人数 > 10”在这份小数据里不会命中，可改成 > 1 验证。

func LearnCode() {
	db, err := database.InitDB("18_learn")
	if err != nil {
		log.Fatal("连接数据库失败：", err)
	}

	if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	seed(db)

	// ===== 练习区：在下面写本章练习题 =====

	// 示例：先把数据读出来看一眼
	// var users []User
	// if err := db.Find(&users).Error; err != nil {
	// 	log.Fatal("查询用户失败：", err)
	// }
	// for _, u := range users {
	// 	fmt.Printf("id=%d name=%s age=%d city=%s\n", u.ID, u.Name, u.Age, u.City)
	// }

	var orders []Order

	subQuery1 := db.Model(&Order{}).Select("AVG(amount)")
	db.Where("amount > (?)", subQuery1).Find(&orders)

	subQuery2 := db.Model(&User{}).Select("city, COUNT(*) as total_people").Group("city")

	var results []struct {
		City string
		TotalPeople uint
	}

	db.Table("(?) as user_city", subQuery2).
		Where("total_people > ?", 10).
		Find(&results)


	var results2 []struct {
		Name string
		OrderCount float64
	}

	db.Model(&User{}).
		Select("name, (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id) as order_count").
		Scan(&results2)

	rows, err := db.Model(&User{}).Where("age > ?", 18).Rows()
	fmt.Println("rows:", rows)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		
		if err := db.ScanRows(rows, &user); err != nil {
			log.Fatal(err)
		}

		fmt.Println(user.Name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	var users []User

	db.Where("name = @name AND age > @minAge AND age < @maxAge",
		sql.Named("name", "张三"),
		sql.Named("minAge", 18),
		sql.Named("maxAge", 60),
	).Find(&users)

}

// seed 重置并写入练习数据，保证每次运行的数据一致
func seed(db *gorm.DB) {
	if err := db.Exec("TRUNCATE TABLE orders").Error; err != nil {
		log.Fatal("清空 orders 失败：", err)
	}
	if err := db.Exec("TRUNCATE TABLE users").Error; err != nil {
		log.Fatal("清空 users 失败：", err)
	}

	users := []User{
		{Name: "张三", Age: 21, City: "北京"},
		{Name: "李四", Age: 28, City: "北京"},
		{Name: "王五", Age: 17, City: "上海"},
		{Name: "赵六", Age: 35, City: "上海"},
		{Name: "钱七", Age: 50, City: "广州"},
		{Name: "孙八", Age: 60, City: "广州"},
	}
	if err := db.Create(&users).Error; err != nil {
		log.Fatal("初始化用户失败：", err)
	}

	// 依赖 Create 回填的自增 ID
	orders := []Order{
		{UserID: users[0].ID, Amount: 100},  // 张三
		{UserID: users[0].ID, Amount: 200},  // 张三
		{UserID: users[0].ID, Amount: 300},  // 张三
		{UserID: users[1].ID, Amount: 50},   // 李四
		{UserID: users[3].ID, Amount: 1000}, // 赵六
		{UserID: users[3].ID, Amount: 500},  // 赵六
		{UserID: users[4].ID, Amount: 80},   // 钱七
		{UserID: users[4].ID, Amount: 120},  // 钱七
		{UserID: users[5].ID, Amount: 2000}, // 孙八
	}
	if err := db.Create(&orders).Error; err != nil {
		log.Fatal("初始化订单失败：", err)
	}
}
