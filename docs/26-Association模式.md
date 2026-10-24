# 第26章：Association 模式 —— 关联的增删改查

## 本章目标
学完本章后，你将能够：
1. 使用 Association 模式管理关联数据
2. 区分 Append、Replace、Delete、Clear 的行为
3. 理解 Association 与直接操作的区别
4. 正确处理多对多关联的增删操作

## 前置知识
- 需要先学习：第21-25章（关联关系 + Preload）
- 需要了解：四种关联类型的基本操作

---

## 26.1 什么是 Association 模式

Association 模式是 GORM 提供的专门用于**操作关联数据**的 API：

```go
// 获取 User 的 Orders 关联的"操作器"
db.Model(&user).Association("Orders")

// 然后可以执行各种操作
db.Model(&user).Association("Orders").Append(&order1)
db.Model(&user).Association("Orders").Delete(&order2)
db.Model(&user).Association("Orders").Count()
```

与直接操作的区别：
```go
// 直接操作（修改数据库字段）
db.Create(&Order{UserID: user.ID, Amount: 100})

// Association 模式（修改关联关系，自动设置外键）
db.Model(&user).Association("Orders").Append(&Order{Amount: 100})
```

---

## 26.2 Association 方法详解

### Append：追加关联

```go
var user User
db.First(&user, 1)

// 追加一个关联
order := Order{Amount: 100}
db.Model(&user).Association("Orders").Append(&order)
// SQL: UPDATE `orders` SET `user_id`=1 WHERE `id`=order.ID
// 如果 order 是新记录，先创建：INSERT INTO `orders` (...) VALUES (...)

// 追加多个关联
orders := []Order{{Amount: 200}, {Amount: 300}}
db.Model(&user).Association("Orders").Append(&orders)
```

对于不同关联类型的行为：
- BelongsTo/HasOne：设置外键值
- HasMany：设置外键值
- Many2Many：在中间表插入记录

### Delete：删除关联

```go
// 删除特定关联（不删除记录本身！）
db.Model(&user).Association("Orders").Delete(&order1)
```

不同关联类型的行为：

| 关联类型 | Delete 行为 |
|---------|-----------|
| BelongsTo | 外键设为 NULL |
| HasOne | 外键设为 NULL |
| HasMany | 外键设为 NULL（默认） |
| Many2Many | 删除中间表记录 |

### Replace：替换关联

```go
// 用新列表替换全部关联
newOrders := []Order{{Amount: 500}, {Amount: 600}}
db.Model(&user).Association("Orders").Replace(&newOrders)
// 1. 清除旧关联（外键设为 NULL）
// 2. 设置新关联
```

### Clear：清空所有关联

```go
// 清空所有关联
db.Model(&user).Association("Orders").Clear()
// HasMany: UPDATE orders SET user_id = NULL WHERE user_id = 1
// Many2Many: DELETE FROM student_courses WHERE student_id = 1
```

### Count：统计关联数量

```go
// 不加载全部数据，只统计数量
count := db.Model(&user).Association("Orders").Count()
fmt.Printf("用户有 %d 个订单\n", count)
// SQL: SELECT COUNT(*) FROM `orders` WHERE user_id = 1
```

### Find：查询关联数据

```go
// 单独查询关联数据
var orders []Order
db.Model(&user).Association("Orders").Find(&orders)
// SQL: SELECT * FROM `orders` WHERE user_id = 1
```

---

## 26.3 Association 完整实战

```go
package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID     uint    `gorm:"primaryKey"`
	Name   string
	Orders []Order // HasMany
}

type Order struct {
	ID     uint    `gorm:"primaryKey"`
	Amount float64
	UserID uint
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&User{}, &Order{})

	// 准备数据
	user := User{Name: "张三"}
	db.Create(&user)

	order1 := &Order{Amount: 100}
	order2 := &Order{Amount: 200}
	order3 := &Order{Amount: 300}
	db.Create(order1)
	db.Create(order2)
	db.Create(order3)

	// 1. Append：给用户添加订单
	fmt.Println("=== Append ===")
	db.Model(&user).Association("Orders").Append(order1, order2)
	count := db.Model(&user).Association("Orders").Count()
	fmt.Printf("Append 后订单数: %d\n", count) // 2

	// 2. Find：查询用户的订单
	fmt.Println("\n=== Find ===")
	var orders []Order
	db.Model(&user).Association("Orders").Find(&orders)
	for _, o := range orders {
		fmt.Printf("  订单#%d: %.2f元\n", o.ID, o.Amount)
	}

	// 3. Replace：替换全部订单
	fmt.Println("\n=== Replace ===")
	db.Model(&user).Association("Orders").Replace([]Order{*order3})
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Replace 后订单数: %d\n", count) // 1

	// 4. Delete：删除特定订单
	fmt.Println("\n=== Delete ===")
	db.Model(&user).Association("Orders").Delete(order3)
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Delete 后订单数: %d\n", count) // 0

	// 5. Clear：清空所有关联
	fmt.Println("\n=== Clear ===")
	db.Model(&user).Association("Orders").Append(order1, order2, order3)
	db.Model(&user).Association("Orders").Clear()
	count = db.Model(&user).Association("Orders").Count()
	fmt.Printf("Clear 后订单数: %d\n", count) // 0
}
```

---

## 26.4 多对多的 Association

```go
type Student struct {
	ID      uint
	Name    string
	Courses []Course `gorm:"many2many:student_courses"`
}

type Course struct {
	ID   uint
	Name string
}

var student Student
db.First(&student, 1)

// Append：选课
course := Course{Name: "数据结构"}
db.Create(&course)
db.Model(&student).Association("Courses").Append(&course)
// 中间表插入记录：INSERT INTO student_courses (student_id, course_id) VALUES (1, 1)

// Delete：退课（只删中间表记录，不删课程本身）
db.Model(&student).Association("Courses").Delete(&course)
// 中间表删除记录：DELETE FROM student_courses WHERE student_id=1 AND course_id=1

// Clear：退选所有课程
db.Model(&student).Association("Courses").Clear()

// Count：统计选课数
count := db.Model(&student).Association("Courses").Count()
```

---

## 常见错误

### 错误1：Append 已经存在关联的记录

```go
// user 已经有 order1 的关联
db.Model(&user).Association("Orders").Append(order1)
// 不会报错，但会产生重复关联（如果外键允许）
// 对于 Many2Many：中间表插入重复记录！
```

### 错误2：混淆 Delete 和删除记录

```go
// Association Delete：只删除关联（user_id 设为 NULL）
db.Model(&user).Association("Orders").Delete(&order)
// SQL: UPDATE `orders` SET `user_id`=NULL WHERE `id`=?

// db.Delete：删除记录本身
db.Delete(&order)
// SQL: DELETE FROM `orders` WHERE `id`=?
// 如果设置了软删除，是 UPDATE SET deleted_at=...
```

### 错误3：忘记 Model()

```go
// ❌ 没有 Model，GORM 不知道操作哪个模型的关联
db.Association("Orders")  // 编译错误！

// ✅ 先 Model，再 Association
db.Model(&user).Association("Orders")
```

---

## 本章小结

| 方法 | HasMany 行为 | Many2Many 行为 |
|------|-------------|---------------|
| `Append` | 设置外键值 | 插入中间表记录 |
| `Delete` | 外键设为 NULL | 删除中间表记录 |
| `Replace` | 清空旧 + 设置新 | 清空旧中间表 + 插入新 |
| `Clear` | 所有外键设为 NULL | 删除所有中间表记录 |
| `Count` | SELECT COUNT(*) | SELECT COUNT(*) FROM 中间表 |
| `Find` | SELECT 关联数据 | SELECT 关联数据（通过中间表） |

- Association 操作关联关系，不删除记录本身
- `Append` 追加，`Replace` 替换，`Delete` 移除，`Clear` 清空
- 必须 `db.Model(&record).Association("RelationName")`

## 练习题

1. 用 Association 给用户追加 3 个订单，然后统计数量。
2. 用 Replace 将用户订单替换为新列表，验证旧订单的 user_id 是否被清空。
3. 实现学生选课：Append 选课、Delete 退课、Count 选课数。
4. 比较 `db.Delete(&order)` 和 `db.Model(&user).Association("Orders").Delete(&order)` 的区别。
5. （思考题）Association 模式在什么场景下比直接操作外键更方便？
