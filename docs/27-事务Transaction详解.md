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
- 需要了解：MySQL 事务基础（BEGIN/COMMIT/ROLLBACK）和 Go 错误处理
- 锁、隔离级别和并发行为依赖数据库与存储引擎；本文以支持事务的 MySQL/InnoDB 为例，具体以目标数据库文档和实测为准。

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
// 开启事务，并立即检查 Begin 错误
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

// 必须检查 Commit 错误
if err := tx.Commit().Error; err != nil {
	return err
}
return nil
```

> ⚠️ 推荐 `db.Transaction()` 处理常规事务。手动事务要检查 `Begin`、各操作和 `Commit` 的错误，并确保所有提前退出路径都会回滚。

---

## 27.4 嵌套事务与 SavePoint

下面假设已定义 `AuditRecord` 模型。GORM 的嵌套 `Transaction` 通常使用 SavePoint 模拟；这不是所有数据库上都能嵌套开启真正独立的物理事务。

```go
err := db.Transaction(func(tx *gorm.DB) error {
	// 内层 GORM Transaction 通常通过 SavePoint 实现嵌套事务（取决于方言支持）。
	if err := tx.Transaction(func(tx2 *gorm.DB) error {
		// 内层操作失败会回滚到保存点；若错误继续返回，外层也会回滚。
		return tx2.Create(&AuditRecord{Message: "inner operation"}).Error
	}); err != nil {
		// 若业务允许内层失败但外层继续，应在这里明确处理该错误。
		return err
	}
	return nil
})
if err != nil {
	return err
}
```

手动使用 SavePoint：
```go
tx := db.Begin()
if tx.Error != nil {
	return tx.Error
}
defer tx.Rollback()

if err := tx.SavePoint("sp1").Error; err != nil {
	return err
}

// ... 一些操作 ...

if err := tx.RollbackTo("sp1").Error; err != nil {
	return err
}
if err := tx.Commit().Error; err != nil {
	return err
}
return nil
```

> 📌 SavePoint 不是独立事务；回滚到保存点后，外层事务仍然存在，最终仍需 Commit 或 Rollback。驱动和数据库对 SavePoint 的支持也可能不同。

---

## 27.5 事务隔离级别

```go
import "database/sql"

// 通过 database/sql 的 TxOptions 请求隔离级别；不再同时执行 SET TRANSACTION，
// 避免重复设置及数据库方言/事务时序差异。
err := db.Transaction(func(tx *gorm.DB) error {
	// ... 事务操作 ...
	return nil
}, &sql.TxOptions{
	Isolation: sql.LevelRepeatableRead,
	ReadOnly:  false,
})
if err != nil {
	return err
}
```

> 📌 驱动/数据库可能不支持或不完全实现请求的隔离级别，务必处理返回错误，并确认数据库默认值和实际事务行为。

四种隔离级别：

| 级别 | 解决 | 不解决 | Go 常量 |
|------|------|--------|---------|
| READ UNCOMMITTED | 通常可读未提交数据 | 具体并发异常依实现而定 | `LevelReadUncommitted` |
| READ COMMITTED | 防止脏读 | 同一事务重复读取可能看到已提交的新值；范围查询可能出现新行 | `LevelReadCommitted` |
| REPEATABLE READ | 防止脏读、普通一致性读通常可重复 | 范围锁定读/写入行为需结合数据库实现；InnoDB 使用间隙/Next-Key 锁等机制处理部分幻读场景 | `LevelRepeatableRead` |
| SERIALIZABLE | 提供更强隔离 | 并发能力可能下降，锁等待/死锁风险增加 | `LevelSerializable` |

---

## 27.6 悲观锁与乐观锁

### 悲观锁（SELECT ... FOR UPDATE）

```go
err := db.Transaction(func(tx *gorm.DB) error {
	var product Product

	// 行锁只在支持该语法的数据库事务中有效，并通常持有到事务结束。
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", 1).
		First(&product).Error; err != nil {
		return err
	}

	if product.Stock < 10 {
		return fmt.Errorf("库存不足")
	}

	result := tx.Model(&product).Update("stock", gorm.Expr("stock - ?", 10))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("库存记录未更新")
	}
	return nil
})
if err != nil {
	return err
}
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
	Where("id = ? AND version = ? AND stock >= ?", id, currentVersion, quantity).
	Updates(map[string]interface{}{
		"stock":   gorm.Expr("stock - ?", quantity),
		"version": gorm.Expr("version + 1"),
	})
if result.Error != nil {
	return result.Error
}
if result.RowsAffected == 0 {
	// 可能是记录不存在、库存不足或版本冲突；必要时重新读取区分原因。
	return fmt.Errorf("更新未生效：记录不存在、库存不足或并发冲突")
}
```

---

## 27.7 完整实战：转账

```go
package main

import (
	"fmt"
	"sort"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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

	return db.Transaction(func(tx *gorm.DB) error {
		// 按 ID 升序锁定两行，降低并发反向转账时的死锁概率。
		ids := []uint{fromID, toID}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		var locked []Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", ids).Order("id ASC").Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != 2 {
			return fmt.Errorf("转出或转入账户不存在")
		}

		accounts := make(map[uint]Account, len(locked))
		for _, account := range locked {
			accounts[account.ID] = account
		}
		from, fromOK := accounts[fromID]
		to, toOK := accounts[toID]
		if !fromOK || !toOK {
			return fmt.Errorf("转出或转入账户不存在")
		}
		if from.Balance < amount {
			return fmt.Errorf("余额不足：当前 %.2f，需要 %.2f", from.Balance, amount)
		}

		if err := tx.Model(&Account{}).Where("id = ?", fromID).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		if err := tx.Model(&Account{}).Where("id = ?", toID).
			Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
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

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}
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

### 错误2：误以为事务会吞掉 panic

```go
err := db.Transaction(func(tx *gorm.DB) error {
	panic("oops")
})
```

GORM 会尝试回滚后重新抛出 panic；它不会把 panic 转换为普通 `error`。如果应用需要恢复，应在合适的上层边界使用 `recover`，不要在事务回调里假装 panic 已被处理。

### 错误3：手动事务忘记 Rollback

```go
tx := db.Begin()
// ... 中间某个操作失败 ...
// 直接 return 了，忘记 tx.Rollback()
// 连接被占用！事务一直开着！

// ✅ 手动事务可用 deferred Rollback 兜底，并检查 Begin/Commit
// 常规流程更推荐 db.Transaction(...)。
tx := db.Begin()
if tx.Error != nil {
	return tx.Error
}
defer tx.Rollback()

// ...执行操作；任何错误直接 return，defer 会回滚...
if err := tx.Commit().Error; err != nil {
	return err
}
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
