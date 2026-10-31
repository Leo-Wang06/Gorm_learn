package chapter27

import (
	// "fmt"
	"learn/internal/database"
	// "log"

	"gorm.io/gorm"
)

// type Account struct {
// 	ID uint
// 	Balance int
// }

func transfer(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 任何提前返回都回滚；提交成功后 Rollback 不会撤销已提交事务。
	defer tx.Rollback()

	if err := tx.Model(&Account{}).Where("id = ?", 1).
		Update("balance", gorm.Expr("balance - ?", 100)).Error; err != nil {
		return err
	}
	if err := tx.Model(&Account{}).Where("id = ?", 2).
		Update("balance", gorm.Expr("balance + ?", 100)).Error; err != nil {
		return err
	}

	/**这里为什么要手动 commit 呢？
	db.Transaction()的自动模式下，是判断传进去的函数是否返回 nil，
	来自动调用 commit 的，
	也就是说，让此函数返回 nil 时才会去调用commit，所以 Transaction 返回的 error 实则就是是否 commit 成功，
	而我们手动开始事务，是需要手动进行 commit 的，并接受它返回的错误。
	*/
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func LearnCode() {
	db, _ := database.InitDB("27_learn")

	// if err := db.AutoMigrate(&Account{}); err != nil {
	// 	log.Fatal("建表失败：", err)
	// }

	// if err := db.Create(&[]Account{{Balance: 200},{Balance: 200}}); err != nil {
	// 	log.Fatal("插入数据失败:", err)
	// }

	// err := db.Transaction(func (tx *gorm.DB) error {
	// 	if err := tx.Model(&Account{}).Where("id = ?", 1).Update("balance", gorm.Expr("balance - ?", 100)).Error; err != nil {
	// 		return err
	// 	}

	// 	if err := tx.Model(&Account{}).Where("id = ?", 2).Update("balance", gorm.Expr("balance + ?", 100)).Error; err != nil {
	// 		return err
	// 	}

	// 	return nil
	// })

	// if err != nil {
	// 	fmt.Println("转账失败，已回滚：", err)
	// } else {
	// 	fmt.Println("转账成功！")
	// }

	PracticeCode(db)
}