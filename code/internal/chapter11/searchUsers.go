package chapter11

import (
	"fmt"

	"gorm.io/gorm"
)

func SearchUsers(db *gorm.DB, name string, minAge int) []User {
	query := db.Model(&User{})

	if name != "" {
		query = query.Where("name Like ?", "%"+name+"%")
		fmt.Println(query)
	}
	if minAge > 0 {
		query = query.Where("Age = ?", minAge)
	}

	var users []User
	query.Find(&users)
	return users
}