package chapter12

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
	Email string
}

func LearnCode() {
	db, _ := database.InitDB("12_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	users := []User{{Name: "张三", Age: 21, Email: "zhangsan@qq.com"},{Name: "李四", Age: 24, Email: "lisiqq.com"}}
	db.Exec("TRUNCATE TABLE users")
	db.Create(&users)

	// db.First(&users, 1)
	// users[0].Name = "张三三"

	// db.Save(&users)

	// db.Model(&users).Where("id = ?", 1).Update("name", "张三三")
	
	// db.Model(&users).Update("age", gorm.Expr("age + ?", 1))

	db.Model(&users).Where("id = ?", 1).Update("name", "张三三")
	db.Model(&users).Where("id = ?", 2).Updates(User{Name: "李四四", Age: 29})

	db.Model(&users).Where("id = ?", 1).Updates(User{Name: "张三", Age: 0, Email: "zhangsan@qq.com"})
	db.Model(&users).Where("id = ?", 1).Updates(map[string]interface{}{"Name": "张三", "Age": 0, "Email": "zhangsan@qq.com"})
	db.Model(&users).Where("id = ?", 1).Update("age", gorm.Expr("age + ?", 1))
	db.Model(&users).Where("id = ?", 1).Select("Name", "Age").Updates(User{Name: "张三三", Age: 0})
	rowsAffected, err := SafeUpdate(db, 1, map[string]interface{}{"name": "张三三", "age": 0})
	if err != nil {
		log.Fatal("更新失败：", err)
	}
	fmt.Println(rowsAffected)
}