package chapter21

import (
	"fmt"
	"log"
	"math/rand"

	"gorm.io/gorm"
)

type User1 struct {
	ID           uint
	Name         string
	DepartmentID uint
	Department   Department
}

type Department struct {
	ID   uint
	Name string
}

func Pracice(db *gorm.DB) {

	if err := db.AutoMigrate(&User1{}, &Department{}); err != nil {
		log.Fatal("建表失败：", err)
	}

	frontEnd := Department{Name: "前端"}
	backEnd := Department{Name: "后端"}
	operations := Department{Name: "运维"}
	testing := Department{Name: "测试"}

	db.Create(&frontEnd)
	db.Create(&backEnd)
	db.Create(&operations)
	db.Create(&testing)

	departments := []Department{frontEnd, backEnd, operations, testing}

	names := []string{"张三", "李四", "王五", "赵六", "钱七"}
	for _, name := range names {
		dept := departments[rand.Intn(len(departments))]

		if err := db.Create(&User1{Name: name, DepartmentID: dept.ID}).Error; err != nil {
			log.Fatal("创建用户失败：", err)
		}

		fmt.Printf("%s → %s (department_id=%d)\n", name, dept.Name, dept.ID)
	}
	
	var users []User1
	db.Preload("Department").Find(&users)

	var frontEndPeople []User1
	db.Joins("Department").Where("Department.name = ?", "前端").Find(&frontEndPeople)
	fmt.Printf("前端部门有 %d 名员工", len(frontEndPeople))

	var backend Department
	if err := db.Where("Department.name = ?", "后端").First(&backend); err != nil {
		log.Fatal("找不到后端部门")
	}

	if err := db.Model(&Department{}).Where("name = ?", "张三").Update("department_id", &backend.ID).Error; err != nil {
		log.Fatal("更新失败")
	}

	fmt.Println()

}