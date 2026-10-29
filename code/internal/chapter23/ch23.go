package chapter23

import (
	"learn/internal/database"
	"log"
)

type Student struct {
	ID      uint `gorm:"primaryKey"`
	Name    string
	Courses []Course `gorm:"many2many:student_courses"` // 多对多
}

type Course struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Students []Student `gorm:"many2many:student_courses"` // 反向多对多
}

func LearnCode() {
	db, _ := database.InitDB("23_learn")

	if err := db.AutoMigrate(&Student{}, &Course{}); err != nil {
		log.Fatal("创建表失败: ", err)
	}

	math := Course{Name: "高等数学"}
	english := Course{Name: "大学英语"}
	physics := Course{Name: "大学物理"}

	if err := db.Create(&math).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&english).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&physics).Error; err != nil {
		panic(err)
	}

	student := Student{
		Name: "张三",
		Courses: []Course{math, english},
	}
	// 不要加 Select("Courses")：它会把 students 表的写入列也限制成"只选 Courses"，
	// 导致 students.name 没被写入（NULL）。普通 Create 会连同多对多关联一起写。
	if err := db.Create(&student).Error; err != nil {
		panic(err)
	}
}