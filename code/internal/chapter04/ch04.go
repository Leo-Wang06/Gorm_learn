package chapter04

import (
	"fmt"
	"learn/internal/database"

	"gorm.io/gorm"
)

type User struct {
	Name string
	Age  int
}

func LearnCode() {
	db, _ := database.InitDB("04_learn")

	dryDB := db.Session(&gorm.Session{DryRun: true})

	stmt := dryDB.Create(&User{Name: "张三", Age: 25})
	fmt.Println(stmt.Statement.SQL.String())

}
