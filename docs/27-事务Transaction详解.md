# 第27章：事务（Transaction）详解

## 本章目标
学完本章后，你将能够：
1. 理解事务的 ACID 特性在 GORM 中的体现
2. 使用自动事务和手动事务
3. 使用 SavePoint 实现嵌套事务
4. 设置事务隔离级别
5. 使用悲观锁和乐观锁
6. 写出正确的转账业务代码

## 前置知识
- 需要先学习：第12-13章（Update、Delete）
- 需要了解：MySQL 事务基础（BEGIN/COMMIT/ROLLBACK）

---

## 27.1 什么是事务

事务是一组要么全部成功、要么全部失败的操作。用转账举例：

```
A 账户扣 100 元 + B 账户加 100 元 = 一个事务

✅ 全部成功：A-100, B+100，数据一致
❌ A 扣了但 B 没加：钱凭空消失了！
❌ A 没扣但 B 加了：钱凭空产生了！
```

数据库的 ACID 保证：
- **原子性（Atomicity）**：A扣+B加，要么全做，要么全不做
- **一致性（Consistency）**：转账前后总金额不变
- **隔离性（Isolation）**：并发转账互不干扰
- **持久性（Durability）**：提交后不丢失

---

## 27.2 `db.Transaction()`：自动事务（推荐）

```go
err := db.Transaction(func(tx *gorm.DB) error {
	// 在这个闭包中，所有操作使用 tx 而不是 db

	// 扣 A 的钱
	if err := tx.Model(&Account{}).Where("id = ?", 1).
		Update("balance", gorm.Expr("balance - ?", 100)).Error; err != nil {
		return err  // 返回 error → 自动回滚
	}

	// 加 B 的钱
	if err := tx.Model(&Account{}).Where("id = ?", 2).
		Update("balance", gorm.Expr("balance + ?", 100)).Error; err != nil {
		return err  // 返回 error → 自动回滚
	}

	// 返回 nil → 自动提交
	return nil
})

if err != nil {
	fmt.Println("转账失败，已回滚:", err)
} else {
	fmt.Println("转账成功！")
}
```

**自动事务的行为**：
- 闭包正常返回 `nil` → `COMMIT`
- 闭包返回 `error` → `ROLLBACK`
- 闭包中 `panic` → `ROLLBACK`

> 📌 **重要**：闭包内必须使用 `tx` 而不是 `db`。用 `db` 的操作不在事务中！

---

## 27.3 手动事务：Begin/Commit/Rollback

当你需要更灵活的控制时：

```go
// 开启事务
tx := db.Begin()

// 操作1
if err := tx.Model(&Account{}).Where("id = ?", 1).
	Update("balance", gorm.Expr("balance - ?", 100)).Error; err != nil {
	tx.Rollback()
	return err
}

// 操作2
if err := tx.Model(&Account{}).Where("id = ?", 2).
	Update("balance", gorm.Expr("balance + ?", 100)).Error; err != nil {
	tx.Rollback()
	return err
}

// 提交事务
tx.Commit()
```

> ⚠️ 手动事务必须记得 Rollback。推荐用 `db.Transaction()`（自动处理）。

---

## 27.4 嵌套事务与 SavePoint

```go
db.Transaction(func(tx *gorm.DB) error {
	// 外层事务

	tx.Transaction(func(tx2 *gorm.DB) error {
		// 内层事务：创建一个 SavePoint
		// 如果失败，只回滚到 SavePoint，不影响外层
		return nil
	})

	return nil
})
```

手动使用 SavePoint：
```go
tx := db.Begin()

tx.SavePoint("sp1")  // 设置保存点

// ... 一些操作 ...

tx.RollbackTo("sp1")  // 回滚到保存点（只撤销 sp1 之后的操作）

tx.Commit()
```

---

## 27.5 事务隔离级别

```go
import "database/sql"

// 设置隔离级别
err := db.Transaction(func(tx *gorm.DB) error {
	tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ")
	// ... 事务操作 ...
	return nil
}, &sql.TxOptions{
	Isolation: sql.LevelRepeatableRead,  // 设置隔离级别
	ReadOnly:  false,
})
```

四种隔离级别：

| 级别 | 解决 | 不解决 | Go 常量 |
|------|------|--------|---------|
| READ UNCOMMITTED | — | 脏读、不可重复读、幻读 | `LevelReadUncommitted` |
| READ COMMITTED | 脏读 | 不可重复读、幻读 | `LevelReadCommitted` |
| REPEATABLE READ（MySQL默认） | 脏读、不可重复读 | 幻读（部分） | `LevelRepeatableRead` |
| SERIALIZABLE | 全部 | — | `LevelSerializable` |

---

## 27.6 悲观锁与乐观锁

### 悲观锁（SELECT ... FOR UPDATE）

```go
db.Transaction(func(tx *gorm.DB) error {
	var product Product

	// 锁定这行数据，其他事务不能修改
	tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", 1).
		First(&product)
	// SQL: SELECT * FROM `products` WHERE id=1 FOR UPDATE

	// 检查库存
	if product.Stock < 10 {
		return fmt.Errorf("库存不足")
	}

	// 扣库存
	tx.Model(&product).Update("stock", gorm.Expr("stock - ?", 10))

	return nil
})
```

### 乐观锁（版本号）

```go
type Product struct {
	ID      uint
	Name    string
	Stock   int
	Version int  // 版本号字段
}

// 扣库存（乐观锁）
result := db.Model(&Product{}).
	Where("id = ? AND version = ?", id, currentVersion).
	Updates(map[string]interface{}{
		"stock":   gorm.Expr("stock - ?", quantity),
		"version": gorm.Expr("version + 1"),
	})

if result.RowsAffected == 0 {
	// 没有行被更新 → 版本号变了 → 并发冲突，重试
	return fmt.Errorf("并发冲突，请重试")
}
```

---

## 27.7 完整实战：转账

```go
package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID      uint    `gorm:"primaryKey"`
	Name    string
	Balance float64
}

func Transfer(db *gorm.DB, fromID, toID uint, amount float64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 查询转出账户（悲观锁）
		var from Account
		if err := tx.Clauses(gorm.Clauses{clause.Locking{Strength: "UPDATE"}}).
			First(&from, fromID).Error; err != nil {
			return fmt.Errorf("转出账户不存在: %w", err)
		}

		// 2. 检查余额
		if from.Balance < amount {
			return fmt.Errorf("余额不足：当前 %.2f，需要 %.2f", from.Balance, amount)
		}

		// 3. 查询转入账户
		var to Account
		if err := tx.First(&to, toID).Error; err != nil {
			return fmt.Errorf("转入账户不存在: %w", err)
		}

		// 4. 扣款
		if err := tx.Model(&from).Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}

		// 5. 加款
		if err := tx.Model(&to).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
			return err
		}

		// 6. 记录转账日志
		log := TransferLog{
			FromID: fromID,
			ToID:   toID,
			Amount: amount,
		}
		if err := tx.Create(&log).Error; err != nil {
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

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&Account{}, &TransferLog{})

	// 创建测试账户
	db.Create(&Account{Name: "张三", Balance: 1000})
	db.Create(&Account{Name: "李四", Balance: 500})

	// 执行转账
	err := Transfer(db, 1, 2, 200)
	if err != nil {
		fmt.Println("转账失败:", err)
	} else {
		fmt.Println("转账成功！")
		var a1, a2 Account
		db.First(&a1, 1)
		db.First(&a2, 2)
		fmt.Printf("张三: %.2f, 李四: %.2f\n", a1.Balance, a2.Balance)
	}
}
```

---

## 常见错误

### 错误1：事务中用 `db` 而不是 `tx`

```go
// ❌ 在事务闭包中用 db
db.Transaction(func(tx *gorm.DB) error {
	db.Create(&record)  // 这个操作不在事务中！
	return nil
})

// ✅ 用 tx
db.Transaction(func(tx *gorm.DB) error {
	tx.Create(&record)  // 在事务中
	return nil
})
```

### 错误2：事务闭包中 panic 被吞掉

```go
db.Transaction(func(tx *gorm.DB) error {
	panic("oops")  // GORM 会 recover，自动 ROLLBACK
	// 但你需要知道出了什么事
})
```

### 错误3：手动事务忘记 Rollback

```go
tx := db.Begin()
// ... 中间某个操作失败 ...
// 直接 return 了，忘记 tx.Rollback()
// 连接被占用！事务一直开着！

// ✅ 用 defer + recover
tx := db.Begin()
defer func() {
	if r := recover(); r != nil {
		tx.Rollback()
	}
}()
```

---

## 本章小结

- 自动事务：`db.Transaction(func(tx) { ... })`（推荐）
- 手动事务：`tx := db.Begin()` / `tx.Commit()` / `tx.Rollback()`
- SavePoint：`tx.SavePoint("name")` / `tx.RollbackTo("name")`
- 悲观锁：`SELECT ... FOR UPDATE`
- 乐观锁：版本号 + `RowsAffected` 判断
- 事务中必须用 `tx`，不能用 `db`

## 练习题

1. 写一个自动事务：A 账户扣 100，B 账户加 100，失败则回滚。
2. 写一个手动事务：插入订单 + 扣库存，失败则回滚。
3. 使用悲观锁实现"抢购"：100 人同时抢 10 件商品，确保不超卖。
4. 使用乐观锁实现：通过版本号防止并发更新冲突。
5. 写一个嵌套事务的 SavePoint 示例。
6. （思考题）为什么自动事务比手动事务更推荐？自动事务有什么限制吗？
