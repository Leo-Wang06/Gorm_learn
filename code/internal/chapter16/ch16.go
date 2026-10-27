package chapter16

import (
	"learn/internal/database"
	"log"
)

type User struct {
	ID   int
	Name string
	Age  int
}

type UserStat struct {
	Name  string
	Count int64
}

type CityCount struct {
	City string
	Count int64
}

type UserCount struct {
	Count int64
	avg_age int64
}

func LearnCode() {
	db, _ := database.InitDB("16_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	users := []User{{Name: "张三", Age: 21},{Name: "李四", Age: 28}}
	db.Create(&users)

	var stats []UserStat
	db.Model(&User{}).Select("name, COUNT(*) as count").Group("name").Scan(&stats)

	// citys := []CityCount{{City: "北京", Count: 15},{City: "天津", Count: 18},{City: "保定", Count: 20}}

	// var results []CityCount

	// db.Model(&User{}).

	var count int64

	db.Model(&User{}).Count(&count)

	db.Model(&User{}).Where("age > ?", 18).Count(&count)

	db.Select("name", "age").Find(&users)

	var UserCount []UserCount

	db.Model(&User{}).Select("COUNT(*) as count, AVG(age) as avg_age").Order("count desc").Scan(&UserCount)

}