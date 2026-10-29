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

	// 要找的是一条「部门」记录，直接查 departments 表即可
	var backend Department
	if err := db.Where("name = ?", "后端").First(&backend).Error; err != nil {
		log.Fatal("找不到后端部门：", err)
	}

	// 把张三调到后端部门：department_id 是 users 表的列，所以 Model 要传 User1
	if err := db.Model(&User1{}).Where("name = ?", "张三").Update("department_id", backend.ID).Error; err != nil {
		log.Fatal("更新失败：", err)
	}

	// 方式一：Preload 预加载关联，再取 Department.Name
	var zhangSan User1
	if err := db.Preload("Department").Where("name = ?", "张三").First(&zhangSan).Error; err != nil {
		log.Fatal("找不到张三：", err)
	}
	fmt.Println("张三所在部门：", zhangSan.Department.Name)

}