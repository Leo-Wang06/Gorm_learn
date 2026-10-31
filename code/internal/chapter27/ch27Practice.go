package chapter27

import (
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	ID      uint    `gorm:"primaryKey"`
	Name    string
	Balance float64
}

func Transfer(db *gorm.DB, fromID, toID uint, amount float64) error {
	if fromID == toID {
		return fmt.Errorf("转出和转入账户不能相同")
	}
	if amount <= 0 {
		return fmt.Errorf("转账金额必须大于 0")
	}

	return db.Transaction(func (tx *gorm.DB) error {
		ids := []uint{fromID, toID}
		sort.Slice(ids, func (i, j int) bool {
			return ids[i] < ids[j]
		})
		var locked []Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ?", ids).Order("id ASC").Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != 2 {
			return fmt.Errorf("转出或转入账户不存在")
		}

		accounts := make(map[uint]Account, len(locked))
		for _, account := range locked {
			accounts[account.ID] = account
		}
		/**这里为什么还要再取一遍？
		当 ids := []uint{2, 1}, 先 form 再 to
		从上方 mysql 取值的时候，不会按照你给的 IN [2, 1] 的顺序来返回值，
		MySQL 语句为：SELECT * FORM account WHERE id IN [2, 1] ORDER BY id ASC
		返回的却是：
		locked = []Account{
			{ID:1, ...},   // locked[0] —— 是"转入"账户
			{ID:2, ...},   // locked[1] —— 是"转出"账户
		}
		所以这样会把转入和转出弄反。
		*/
		from, formOK := accounts[fromID]
		_, toOK := accounts[toID]
		if !formOK || !toOK {
			return fmt.Errorf("转出或转入账户不存在")
		}
		if from.Balance < amount {
			return fmt.Errorf("余额不足：当前 %.2f，需要 %.2f", from.Balance, amount)
		}

		if err := tx.Model(&Account{}).Where("id = ?", fromID).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&Account{}).Where("id = ?", toID).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}

		transferLog := TransferLog{FromID: fromID, ToID: toID, Amount: amount}
		if err := tx.Create(&transferLog).Error; err != nil {
			return err
		}
		return nil
	})
}

type TransferLog struct {
	ID     uint
	FromID uint
	ToID   uint
	Amount float64
}

func PracticeCode(db *gorm.DB) {
	if err := db.AutoMigrate(&Account{}, &TransferLog{}); err != nil {
		panic(err)
	}
	// 仅限专用练习数据库：下面会删除这两张表的全部记录，不要对真实或共享数据库运行。
	if err := db.Exec("DELETE FROM transfer_logs").Error; err != nil {
		panic(err)
	}
	if err := db.Exec("DELETE FROM accounts").Error; err != nil {
		panic(err)
	}
	if err := db.Create(&Account{ID: 1, Name: "张三", Balance: 1000}).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&Account{ID: 2, Name: "李四", Balance: 500}).Error; err != nil {
		panic(err)
	}

	if err := Transfer(db, 1, 2, 200); err != nil {
		fmt.Println("转账失败:", err)
		return
	}
	var a1, a2 Account
	if err := db.First(&a1, 1).Error; err != nil {
		panic(err)
	}
	if err := db.First(&a2, 2).Error; err != nil {
		panic(err)
	}
	fmt.Printf("转账成功！张三: %.2f, 李四: %.2f\n", a1.Balance, a2.Balance)

}