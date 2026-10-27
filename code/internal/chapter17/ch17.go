package chapter17

import (
	"learn/internal/database"
	"log"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
	CompanyID uint
	Company Company
}

type Company struct {
	ID uint
	Name string
}

type Employee struct {
	ID uint
	Name string
	ManagerID *uint
	Manager *Employee
}

type EmpWithManager struct {
	ID uint
	Name string
	ManagerName string
}

func LearnCode() {
	db, _ := database.InitDB("17_learn")

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	users := []User{{Name: "张三", CompanyID: 1}, {Name: "李四"}}
	db.Create(&users)

	companies := Company{Name: "Google"}

	db.Create(&companies)

	db.Joins("Company").Find(&users)

	// db.Joins("Company").Find(&users)

	if err := db.AutoMigrate(&Employee{}); err != nil {
		log.Fatal(err)
	}

	var results []EmpWithManager

	db.Model(&Employee{}).
		Select("employees.name, employees.manager_id, m.name AS manager_name").
		Joins("LEFT JOIN employees AS m ON employees.manager_id = m.id").
		Scan(&results)

}