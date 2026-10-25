package chapter07

import (
	"learn/internal/database"
	"log"
	"time"
)

type User struct {
	ID        uint
	Settings Settings `gorm:"serializer:json"`
	Tags      []string  `gorm:"serializer:json"`
	Meta      MetaInfo  `gorm:"serializer:json"`
}

type Settings struct {
	Theme  string `json:"theme"`
	Lang   string `json:"lang"`
	Notify bool   `json:"notify"`
}

type User1 struct {
	CreatedAt time.Time
	Birthday *time.Time
}

type MetaInfo map[string]interface{}

type User2 struct {
	Name string
	Age *int
}

func LearnCode() {

	db, _ := database.InitDB("07_learn")

	if err := db.AutoMigrate(&User{}); err != nil {   // ← 补上
		log.Fatal(err)
	}

	user := User{
		Settings: Settings{Theme: "dark", Lang: "zh", Notify: true},
		Tags:      []string{"golang", "gorm"},
		Meta:      MetaInfo{"login_count": 42},
	}

	db.Create(&user)

	// age := 0
	// fmt.Println(&age)

	age := 0
	user2 := User2{Name: "张三", Age: &age}
	db.Model(&user).Updates(user2)

	Practice(db)

}