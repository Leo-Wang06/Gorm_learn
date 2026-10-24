# 第21章：关联关系概述与 BelongsTo

## 本章目标
学完本章后，你将能够：
1. 理解 GORM 四种关联关系的概念和区别
2. 掌握外键（ForeignKey）和引用（References）的配置
3. 定义和使用 BelongsTo 关联
4. 创建和查询带 BelongsTo 关联的记录
5. 理解 GORM 默认外键的命名规则

## 前置知识
- 需要先学习：第05-08章（模型定义）、第09-20章（CRUD与查询）
- 需要了解：MySQL 外键概念

---

## 21.1 四种关联关系全景图

GORM 支持四种关联关系，正好对应数据库中最常见的表关系：

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│BelongsTo│     │ HasOne   │     │ HasMany  │     │Many2Many │
│  属于    │     │  有一个   │     │  有多个   │     │ 多对多   │
└──────────┘     └──────────┘     └──────────┘     └──────────┘

User 属于 Company  User 有一张信用卡  User 有多个订单  Student 选修多门 Course
    ↑ 外键在 User      ↑ 外键在 CreditCard   ↑ 外键在 Order     ↑ 中间表
```

| 关联类型 | 关系 | 外键位置 | 数据库体现 |
|---------|------|---------|----------|
| `BelongsTo` | A 属于 B | 外键在 A | `A.b_id` → `B.id` |
| `HasOne` | A 有一个 B | 外键在 B | `B.a_id` → `A.id` |
| `HasMany` | A 有多个 B | 外键在 B | `B.a_id` → `A.id` |
| `Many2Many` | A 和 B 多对多 | 中间表 | `a_b` 中间表 |

> 📌 **理解关联的关键**：外键放在哪张表？→ BelongsTo 外键在"自己"身上，HasOne/HasMany 外键在"对方"身上。

---

## 21.2 BelongsTo：属于关系

### 基础示例：User 属于 Company

```go
// 公司（被归属方）
type Company struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

// 用户（归属方，外键在 User）
type User struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string
	CompanyID uint    // 外键字段（数据库列）
	Company   Company // 关联对象（Go 字段，不是数据库列）
}
```

数据库表结构：
```sql
-- users 表
CREATE TABLE users (
    id         BIGINT PRIMARY KEY,
    name       VARCHAR(255),
    company_id BIGINT,           -- 外键列
    FOREIGN KEY (company_id) REFERENCES companies(id)  -- 外键约束
);

-- companies 表
CREATE TABLE companies (
    id   BIGINT PRIMARY KEY,
    name VARCHAR(255)
);
```

Go 代码中使用：
```go
// 创建公司和用户
company := Company{Name: "Google"}
db.Create(&company)

user := User{Name: "张三", CompanyID: company.ID}  // 设置外键值
db.Create(&user)
```

### BelongsTo 标签配置

```go
type User struct {
	ID        uint
	Name      string
	CompanyID uint
	Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
	//                      ↑ 指定外键字段      ↑ 指定引用字段
}
```

| 标签选项 | 含义 | 默认值 |
|---------|------|--------|
| `foreignKey` | 当前结构体中哪个字段是外键 | `{关联类型}ID` → `CompanyID` |
| `references` | 关联到对方表的哪个字段 | 对方表的主键（通常是 `ID`） |

### 最简单的写法（使用默认约定）

```go
type User struct {
	ID        uint
	Name      string
	CompanyID uint    // 按约定：类型名(Company) + ID = 外键
	Company   Company // 按约定：关联类型为 Company，外键为 CompanyID
}
// 不需要写任何标签！GORM 自动推断：
// foreignKey: CompanyID（类型名 Company + ID）
// references: ID（Company 表的主键）
```

> 📌 **约定大于配置**：只要你遵循命名规则（`类型名 + ID`），GORM 就能自动识别关联关系。

---

## 21.3 创建带 BelongsTo 的记录

### 方式一：手动设置外键

```go
// 先创建公司
company := Company{Name: "Google"}
db.Create(&company)

// 创建用户时手动设置 CompanyID
user := User{
	Name:      "张三",
	CompanyID: company.ID,  // 手动设置外键
}
db.Create(&user)
```

### 方式二：直接赋值关联对象（GORM 自动提取 ID）

```go
// 创建用户时直接传入 Company 结构体
user := User{
	Name:    "张三",
	Company: company,  // GORM 会自动提取 company.ID 设置到 CompanyID
}
db.Create(&user)
// SQL: INSERT INTO `users` (`name`, `company_id`) VALUES ('张三', 1)
```

### 方式三：创建时同时创建关联记录

```go
// User 创建时也创建 Company（如果 Company 不存在）
user := User{
	Name:    "张三",
	Company: Company{Name: "新公司"},
}

// ⚠️ 默认不会自动创建 Company！需要指定
db.Select("Company").Create(&user)
// 或
db.Omit("Company").Create(&user)  // 不处理关联

// 同时创建 User 和 Company
db.Create(&user)  // 只创建 User，不创建 Company（默认行为）
db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&user)  // 同时创建
```

---

## 21.4 查询带 BelongsTo 的记录

### 普通查询（不加载关联）

```go
var user User
db.First(&user, 1)
// SQL: SELECT * FROM `users` WHERE id=1
// user.Company 是零值（没有被填充）
fmt.Println(user.Company.Name)  // 空字符串！
```

### Preload：预加载关联

```go
var user User
db.Preload("Company").First(&user, 1)
// SQL: SELECT * FROM `users` WHERE id=1
// SQL: SELECT * FROM `companies` WHERE id=1  (自动查关联)
fmt.Println(user.Company.Name)  // "Google" ✅
```

### Joins：使用 JOIN 加载

```go
var user User
db.Joins("Company").First(&user, 1)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `companies` AS `Company` ON `users`.`company_id` = `Company`.`id`
//      WHERE `users`.`id`=1
// 注意：虽然 JOIN 了 companies 表，但 user.Company 需要单独获取
```

> 📌 Preload 和 Joins 的区别：Preload 用两条 SQL 加载，Joins 用一条 SQL 加载。Preload 自动填充关联结构体，Joins 需要通过 Select+Scan 获取。关联查询的详细对比在第25章。

---

## 21.5 更新带 BelongsTo 的记录

```go
var user User
db.First(&user, 1)

// 换公司：修改外键
db.Model(&user).Update("CompanyID", 2)

// 或者用关联对象
user.Company = Company{ID: 3}  // 设置新的 Company
db.Model(&user).Select("CompanyID").Updates(user)
// 或
db.Model(&user).Update("CompanyID", 3)
```

---

## 21.6 完整示例：用户属于公司

```go
package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Company struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);not null"`
}

type User struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"type:varchar(100);not null"`
	CompanyID uint    // 外键
	Company   Company // BelongsTo 关联
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&Company{}, &User{})

	// 1. 创建公司
	google := Company{Name: "Google"}
	apple := Company{Name: "Apple"}
	db.Create(&google)
	db.Create(&apple)

	// 2. 创建用户，分配到公司
	db.Create(&User{Name: "张三", CompanyID: google.ID})
	db.Create(&User{Name: "李四", CompanyID: apple.ID})
	db.Create(&User{Name: "王五", CompanyID: google.ID})

	// 3. 查询用户及其公司
	var users []User
	db.Preload("Company").Find(&users)
	for _, u := range users {
		fmt.Printf("%s 在 %s 工作\n", u.Name, u.Company.Name)
	}
	// 输出：
	// 张三 在 Google 工作
	// 李四 在 Apple 工作
	// 王五 在 Google 工作

	// 4. 查询在 Google 工作的用户
	var googleUsers []User
	db.Joins("Company").Where("Company.name = ?", "Google").Find(&googleUsers)
	fmt.Printf("Google 有 %d 名员工\n", len(googleUsers))
}
```

---

## 常见错误

### 错误1：外键命名不符合约定导致关联失败

```go
// ❌ 外键字段名不匹配约定
type User struct {
	ID       uint
	CompID   uint    // 不叫 CompanyID
	Company  Company // GORM 找不到外键 CompID！
}

// ✅ 显式指定 foreignKey
type User struct {
	ID      uint
	CompID  uint
	Company Company `gorm:"foreignKey:CompID"`
}
```

### 错误2：忘记 Preload 导致关联字段为空

```go
var user User
db.First(&user, 1)
fmt.Println(user.Company.Name)  // ❌ 空！没有 Preload
// 因为 GORM 只查了 users 表
```

### 错误3：循环 Preload 导致无限加载

```go
type User struct {
	ID      uint
	Company Company
}
type Company struct {
	ID    uint
	Users []User  // Company 又引用了 User
}

// ❌ 危险：如果两边都 Preload 会无限循环
db.Preload("Company.Users.Company.Users...").Find(&users)
```

---

## 本章小结

- BelongsTo：外键在"自己"身上，`A.b_id → B.id`
- 约定外键名：`{关联类型名}ID`，如 `CompanyID`
- 外键标签：`gorm:"foreignKey:CompanyID;references:ID"`
- 查询关联用 `Preload("Company")` 或 `Joins("Company")`
- 创建时赋值 `Company` 对象，GORM 自动提取 ID

## 练习题

1. 定义 `User` 和 `Department` 两个模型，User BelongsTo Department。
2. 创建 3 个部门和 5 个用户，分别分配到不同部门。
3. 用 `Preload` 查询所有用户及其部门信息。
4. 用 `Joins` 查询在某个部门工作的所有用户。
5. 更新一个用户的部门。
6. （思考题）BelongsTo 和 HasOne 都是"一对一"关系，它们的区别是什么？什么时候该用 BelongsTo，什么时候该用 HasOne？
