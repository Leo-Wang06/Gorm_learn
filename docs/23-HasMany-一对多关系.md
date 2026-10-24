# 第23章：HasMany —— 一对多关系

## 本章目标
学完本章后，你将能够：
1. 定义和使用 HasMany 关联
2. 创建带一对多关联的记录
3. 查询和过滤 HasMany 关联数据
4. 理解 HasMany 的增删改操作
5. 在实际项目中使用一对多关系

## 前置知识
- 需要先学习：第21-22章（BelongsTo、HasOne）
- 需要了解：MySQL 的外键和一对多关系

---

## 23.1 HasMany：有一个对多个

HasMany 表示"A 拥有多个 B"，外键在 B 表中（和 HasOne 一样，但没有 UNIQUE 约束）：

```
┌──────────┐         ┌──────────┐
│   User   │ ──Has──>│  Order   │
│  id: 1   │         │  id: 1   │
│  name:张三│         │  amount:100│
│          │         │  user_id:1│ ← 外键
└──────────┘         ├──────────┤
                     │  Order   │
                     │  id: 2   │
                     │  amount:200│
                     │  user_id:1│ ← 同一个 user_id
                     └──────────┘
```

### 基础示例：User 有多个 Order

```go
type User struct {
	ID     uint    `gorm:"primaryKey"`
	Name   string
	Orders []Order // HasMany：一个 User 有多个 Order
}

type Order struct {
	ID     uint    `gorm:"primaryKey"`
	Amount float64
	UserID uint    // 外键（不设 uniqueIndex，允许多条相同 user_id）
}
```

---

## 23.2 HasMany 标签配置

```go
type User struct {
	ID     uint
	Name   string
	Orders []Order `gorm:"foreignKey:UserID;references:ID"`
	//               ↑ Order 表中的外键字段      ↑ User 表中的引用字段
}
```

### 完整的双向定义

通常在一对多关系中，两边都定义关联：

```go
type User struct {
	ID     uint
	Name   string
	Orders []Order  // HasMany：User 有多个 Order
}

type Order struct {
	ID     uint
	Amount float64
	UserID uint
	User   User    // BelongsTo：Order 属于 User（反向关联）
}
```

这样你可以从两端访问：
- `user.Orders` → 用户的全部订单
- `order.User` → 订单所属的用户

---

## 23.3 创建 HasMany 记录

### 方式一：手动设置外键

```go
user := User{Name: "张三"}
db.Create(&user)

// 创建订单时手动设置 UserID
order1 := Order{Amount: 100, UserID: user.ID}
order2 := Order{Amount: 200, UserID: user.ID}
db.Create(&order1)
db.Create(&order2)
```

### 方式二：通过关联对象创建

```go
// 创建用户时同时创建订单
user := User{
	Name: "张三",
	Orders: []Order{
		{Amount: 100},
		{Amount: 200},
		{Amount: 300},
	},
}
db.Select("Orders").Create(&user)  // 同时创建 User 和 Orders

// 给已有用户追加订单
var user User
db.First(&user, 1)
db.Model(&user).Association("Orders").Append(
	&Order{Amount: 400},
	&Order{Amount: 500},
)
```

---

## 23.4 查询 HasMany 关联

### 预加载

```go
// 查询所有用户及其订单
var users []User
db.Preload("Orders").Find(&users)
// SQL: SELECT * FROM `users`
// SQL: SELECT * FROM `orders` WHERE user_id IN (1,2,3,...)

for _, u := range users {
	fmt.Printf("%s 有 %d 个订单\n", u.Name, len(u.Orders))
}
```

### 条件预加载

```go
// 只预加载金额大于 100 的订单
db.Preload("Orders", "amount > ?", 100).Find(&users)

// 预加载时排序
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
	return db.Order("amount desc").Limit(5)
}).Find(&users)
```

### 用 Joins 查询有订单的用户

```go
// 查有订单的用户（用 DISTINCT 去重）
var users []User
db.Joins("Orders").Distinct().Find(&users)
// 等价于：SELECT DISTINCT users.* FROM users
//         JOIN orders ON users.id = orders.user_id
```

---

## 23.5 一对多关系的增删改操作

### 追加关联

```go
var user User
db.First(&user, 1)

// 方式一：Association Append（推荐）
db.Model(&user).Association("Orders").Append(
	&Order{Amount: 600},
	&Order{Amount: 700},
)
// 自动设置 order.user_id = user.id

// 方式二：直接创建（手动设置外键）
db.Create(&Order{Amount: 600, UserID: user.ID})
```

### 替换关联

```go
// 把用户的订单替换为新列表（旧订单的 user_id 被清空）
newOrders := []Order{
	{Amount: 800},
	{Amount: 900},
}
db.Model(&user).Association("Orders").Replace(&newOrders)
```

### 删除关联

```go
// 删除特定订单的关联（把 user_id 设为 NULL，不删除记录）
db.Model(&user).Association("Orders").Delete(&order1)

// 清空所有关联（把所有订单的 user_id 设为 NULL）
db.Model(&user).Association("Orders").Clear()
```

### 统计关联数量

```go
count := db.Model(&user).Association("Orders").Count()
fmt.Printf("用户 %s 有 %d 个订单\n", user.Name, count)
```

---

## 23.6 完整示例：用户与订单

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
	Name   string  `gorm:"type:varchar(100);not null"`
	Orders []Order // HasMany
}

type Order struct {
	ID      uint    `gorm:"primaryKey"`
	Amount  float64 `gorm:"type:decimal(10,2)"`
	Status  string  `gorm:"type:varchar(20);default:'pending'"`
	UserID  uint    // 外键
	User    User    // BelongsTo（反向关联）
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&User{}, &Order{})

	// 1. 创建用户和订单
	user := User{
		Name: "张三",
		Orders: []Order{
			{Amount: 150.00, Status: "paid"},
			{Amount: 89.50, Status: "pending"},
			{Amount: 299.00, Status: "paid"},
		},
	}
	db.Select("Orders").Create(&user)

	// 2. 查询所有用户及订单
	var users []User
	db.Preload("Orders").Find(&users)
	for _, u := range users {
		var total float64
		for _, o := range u.Orders {
			total += o.Amount
		}
		fmt.Printf("%s: %d个订单, 总金额 %.2f\n", u.Name, len(u.Orders), total)
	}

	// 3. 查询有已支付订单的用户
	var paidUsers []User
	db.Joins("Orders").Where("orders.status = ?", "paid").Distinct().Find(&paidUsers)

	// 4. 追加订单
	var foundUser User
	db.First(&foundUser, user.ID)
	db.Model(&foundUser).Association("Orders").Append(
		&Order{Amount: 500, Status: "paid"},
	)
}
```

---

## 常见错误

### 错误1：一对多用 Preload，N+1 问题

```go
// ❌ 不 Preload 却在循环中访问 Orders
var users []User
db.Find(&users)
for _, u := range users {
	fmt.Println(len(u.Orders))  // Orders 为空！没有 Preload
}

// ✅ 加 Preload
db.Preload("Orders").Find(&users)
```

### 错误2：Association 和 Preload 混淆

```go
// Association：操作关联数据（增删改查关联）
db.Model(&user).Association("Orders").Append(&order)

// Preload：加载关联数据到结构体中（只查询）
db.Preload("Orders").Find(&users)
```

### 错误3：直接用 Find 查 HasMany 导致大量数据

```go
// ❌ 查全部订单（可能百万级！）
db.Preload("Orders").Find(&users)
// 如果每个 User 有 1000 个 Order，10 个 User = 10000 条数据到内存

// ✅ 用条件 Preload 限制
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
	return db.Where("created_at > ?", time.Now().Add(-30*24*time.Hour)).Limit(10)
}).Find(&users)
```

---

## 本章小结

- HasMany：A 有多个 B，外键在 B（没有 UNIQUE 约束）
- 字段类型：`[]Order`（切片，不是单个结构体）
- Preload 加载关联，Association 操作关联
- 条件预加载避免加载不必要的数据
- 双向定义（HasMany + BelongsTo）可以从两端访问

## 练习题

1. 定义 `Department` 和 `Employee` 模型，Department HasMany Employees。
2. 创建一个部门及其多个员工，用 Select 同时创建。
3. 用 Preload 查询所有部门及其员工，打印每个部门的人数。
4. 使用 Association 追加、删除、替换员工。
5. 写查询：找出员工数量大于 5 的部门。
6. （思考题）如果 User 有 10000 个 Order，`Preload("Orders")` 会有什么问题？如何优化？
