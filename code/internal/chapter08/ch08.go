package chapter08

import (
	"fmt"
	"learn/internal/database"
	"log"
)

type User struct {
	ID uint
	Name string
	Email string
}

func LearnCode() {
	db, _ := database.InitDB("08_learn")

	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}

	migrator := db.Migrator()

	if migrator.HasTable(&User{}) {
		fmt.Println("user 表存在")
	}

	if migrator.HasTable("users") {
		fmt.Println("user 表存在")
	}

	// migrator.CreateTable(&User{})
	// migrator.RenameTable(&User{}, "actived_users")
	// migrator.RenameTable("users", "actived_users")

	if migrator.HasColumn(&User{}, "email") {
		fmt.Println("email 列存在")
	}

}