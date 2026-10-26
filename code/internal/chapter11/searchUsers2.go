package chapter11

import "gorm.io/gorm"

func SearchUsers2(db *gorm.DB, name string, minAge int, maxAge int) []User {
	query := db.Model(&User{})

	if name != "" {
		query.Where("name = ?", name)
	}
	if minAge > 0  {
		query.Where("age >= ?", minAge).Where("age <= ?", maxAge)
	}

	var users []User
	query.Find(&users)
	return users

}