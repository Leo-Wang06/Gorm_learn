# 第22章：HasOne —— 一对一关系

## 本章目标
学完本章后，你将能够：
1. 定义和使用 HasOne 关联
2. 理解 HasOne 与 BelongsTo 的本质区别
3. 创建、查询、更新 HasOne 关联记录
4. 正确选择 HasOne vs BelongsTo

## 前置知识
- 需要先学习：第21章（BelongsTo）
- 需要了解：MySQL 外键、UNIQUE 约束

---

## 22.1 HasOne：拥有一个

HasOne 表示"A 拥有一个 B"，外键在 B 表中：

```
┌──────────┐         ┌──────────────┐
│   User   │ ──Has──>│  CreditCard  │
│  id: 1   │         │  id: 1       │
│  name:张三│         │  number: xxxx│
│          │         │  user_id: 1  │  ← 外键在 CreditCard
└──────────┘         └──────────────┘
```

### 基础示例：User 有一张 CreditCard

```go
type User struct {
	ID         uint       `gorm:"primaryKey"`
	Name       string
	CreditCard CreditCard // HasOne 关联
}

type CreditCard struct {
	ID     uint   `gorm:"primaryKey"`
	Number string
	UserID uint   `gorm:"uniqueIndex"` // 外键（指向 User）；唯一索引约束每个用户最多一张卡
}
```

数据库表结构：
```sql
-- users 表
CREATE TABLE users (id BIGINT PRIMARY KEY, name VARCHAR(255));

-- credit_cards 表
CREATE TABLE credit_cards (
    id      BIGINT PRIMARY KEY,
    number  VARCHAR(255),
    user_id BIGINT UNIQUE,           -- UNIQUE 保证一对一！
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

> 📌 `HasOne` 描述的是 GORM 模型中的关联方向；数据库不会仅凭 Go 字段类型保证一对一。要在数据库层面限制每个 User 最多一张卡，应给 `credit_cards.user_id` 添加 UNIQUE 约束。没有该约束时，模型仍可声明 HasOne，但数据库允许出现多张卡，数据完整性就无法保证。

---

## 22.2 HasOne 标签配置

```go
type User struct {
	ID         uint
	Name       string
	CreditCard CreditCard `gorm:"foreignKey:UserID;references:ID"`
	//                      ↑ 外键在 CreditCard 表的 UserID 字段
	//                                    ↑ 引用 User 表的 ID
}
```

对比 BelongsTo：
```go
// BelongsTo：外键在 User（自己）
type User struct {
	CompanyID uint
	Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
}

// HasOne：外键在 CreditCard（对方）
type User struct {
	CreditCard CreditCard `gorm:"foreignKey:UserID;references:ID"`
}
```

> 📌 **HasOne 的外键在"对方"表**。`foreignKey` 指的是对方表（CreditCard）中的外键字段。

---

## 22.3 创建带 HasOne 的记录

```go
// 创建用户及其信用卡
user := User{
	Name: "张三",
	CreditCard: CreditCard{
		Number: "4111-1111-1111-1111",
	},
}

// 写法一：Create 默认会保存非零值 HasOne 关联，并将新 User 的 ID 写入 CreditCard.UserID
if err := db.Create(&user).Error; err != nil {
	panic(err)
}

// 写法二：需要显式控制字段时，改用 Select（不要再对同一个 user 重复 Create）
// err := db.Select("Name", "CreditCard").Create(&user).Error

// Omit("CreditCard") 可跳过关联保存；也可在主记录创建后单独创建子记录
// 但单独创建时要先确保 CreditCard.UserID 已正确设置。
```

### 全关联保存

```go
db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&user)
// 会同时保存 User 和 CreditCard
```

---

## 22.4 查询 HasOne 关联

```go
// Preload 方式
var user User
db.Preload("CreditCard").First(&user, 1)
// SQL: SELECT * FROM `users` WHERE id=1
// SQL: SELECT * FROM `credit_cards` WHERE user_id=1
fmt.Println(user.CreditCard.Number)  // ✅ 有值

// Joins 方式
db.Joins("CreditCard").First(&user, 1)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `credit_cards` ON `users`.`id` = `credit_cards`.`user_id`
//      WHERE `users`.`id` = 1
```

---

## 22.5 HasOne vs BelongsTo：决策指南

两者的核心区别是**外键放在哪张表**：

```
HasOne:
  User ──Has──> CreditCard
  外键在 CreditCard（user_id）

BelongsTo:
  User ──BelongsTo──> Company
  外键在 User（company_id）
```

### 如何选择

| 场景 | 用什么 | 外键位置 |
|------|--------|---------|
| 用户有一张信用卡 | HasOne | credit_cards.user_id |
| 用户属于一个公司 | BelongsTo | users.company_id |
| 文章有一个详情 | HasOne | article_details.article_id |
| 订单属于一个用户 | BelongsTo | orders.user_id |

### 判断口诀

- **谁"拥有"关系，外键就在对方身上** → HasOne（User "拥有" CreditCard，外键在 CreditCard）
- **谁"归属"于对方，外键就在自己身上** → BelongsTo（User "归属"于 Company，外键在 User）

### 数据库体现

```
HasOne：外键在"被拥有"的表
  users ───< credit_cards (user_id UNIQUE)

BelongsTo：外键在"归属"的表
  companies ───< users (company_id)
```

---

## 22.6 完整示例：用户与个人资料

```go
package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string
	Profile Profile // HasOne
}

type Profile struct {
	ID     uint   `gorm:"primaryKey"`
	Bio    string `gorm:"type:text"`
	Avatar string `gorm:"type:varchar(255)"`
	UserID uint   `gorm:"uniqueIndex"` // uniqueIndex 确保一对一
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&User{}, &Profile{})

	// 1. 创建用户及个人资料
	user := User{
		Name: "张三",
		Profile: Profile{
			Bio:    "全栈工程师",
			Avatar: "https://example.com/avatar.png",
		},
	}
	db.Select("Profile").Create(&user)

	// 2. 查询用户及其个人资料
	var found User
	db.Preload("Profile").First(&found, user.ID)
	fmt.Printf("%s: %s, 头像: %s\n",
		found.Name, found.Profile.Bio, found.Profile.Avatar)

	// 3. 更新个人资料
	db.Model(&found.Profile).Updates(map[string]interface{}{
		"bio":    "高级全栈工程师",
		"avatar": "https://example.com/new_avatar.png",
	})
}
```

---

## 常见错误

### 错误1：忘记 UNIQUE 约束（一对一变一对多）

```go
type CreditCard struct {
	UserID uint  // ❌ 没有 unique 约束
}
// 一个 User 可以有多张 CreditCard → 实际是 HasMany！

// ✅ HasOne 需要 UNIQUE
type CreditCard struct {
	UserID uint `gorm:"uniqueIndex"`
}
```

### 错误2：HasOne 和 BelongsTo 方向搞反

```go
// ❌ 用 HasOne 表示 User 属于 Company
type User struct {
	Company Company  // 外键应该在 User，这是 BelongsTo！
}

// ✅ 用 BelongsTo
type User struct {
	CompanyID uint
	Company   Company `gorm:"foreignKey:CompanyID"`
}
```

### 错误3：误以为 Create 默认忽略关联

```go
user := User{
	Name:    "张三",
	Profile: Profile{Bio: "工程师"},
}

// Create 默认会保存非零值关联；检查返回的 Error 以确认操作成功
if err := db.Create(&user).Error; err != nil {
	panic(err)
}

// 需要控制本次参与创建的字段时，改用 Select 或 Omit，不要重复 Create 同一实例
// db.Select("Name", "Profile").Create(&user)
```

---

## 本章小结

- HasOne：外键在"对方"表，需要 `uniqueIndex` 确保一对一
- BelongsTo：外键在"自己"表
- 选择：谁"拥有" → HasOne，谁"归属" → BelongsTo
- 创建关联用 `Select("Relation")` 或 `FullSaveAssociations`
- 查询关联用 `Preload("Relation")` 或 `Joins("Relation")`

## 练习题

1. 定义 `User` 和 `IDCard`（身份证）模型，使用 HasOne 关联。
2. 创建一个用户及其身份证信息。
3. 用 Preload 查询用户及其身份证。
4. 对比 HasOne 和 BelongsTo：写出两者的代码示例，标注外键位置。
5. 如果 CreditCard 的 UserID 没有 `uniqueIndex`，会发生什么？
6. （思考题）什么时候应该把两个一对一关系的数据存在同一张表，什么时候应该分成两张表？
