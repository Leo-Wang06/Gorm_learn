package chapter28

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	UUID      string `gorm:"type:varchar(36);uniqueIndex"`
	Username  string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string `gorm:"type:varchar(255);not null"`
	Email     string `gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdateAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.NewString()
	}

	if u.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败：%w", err)
		}
		u.Password = string(hashed)
	}

	if u.Email == "" {
		u.Email = u.Username + "@example.com"
	}
	return  nil
}

func UpdatePassword(db *gorm.DB, userID uint, plainPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Model(&User{}).Where("id = ?", userID).Update("Password", string(hashed)).Error
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	// 不允许删除 ID 为 1 的管理员
	if u.ID == 1 {
		return fmt.Errorf("不允许删除系统管理员")
	}
	return nil
}

func LearnCode() {

}