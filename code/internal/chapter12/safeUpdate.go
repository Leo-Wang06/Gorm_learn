package chapter12

import (
	"errors"

	"gorm.io/gorm"
)

var updatableUserFields = map[string]struct{}{
	"name":  {},
	"age":   {},
	"email": {},
}

func SafeUpdate(db *gorm.DB, id uint, data map[string]interface{}) (int64, error) {
	if id == 0 {
		return 0, errors.New("id 不能为 0")
	}
	if len(data) == 0 {
		return 0, errors.New("没有要更新的字段")
	}

	clean := make(map[string]interface{}, len(data))
	for k, v := range data {
		if _, ok := updatableUserFields[k]; ok {
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		return 0, errors.New("没有允许更新的字段")
	}

	result := db.Model(&User{}).Where("id = ?", id).Updates(clean)
	return result.RowsAffected, result.Error
}
