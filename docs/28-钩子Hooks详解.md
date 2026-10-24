# 第28章：钩子（Hooks）详解

## 本章目标
学完本章后，你将能够：
1. 理解 GORM 的完整生命周期和钩子执行顺序
2. 使用 BeforeCreate/AfterCreate 等创建钩子
3. 使用 BeforeUpdate/AfterUpdate 等更新钩子
4. 使用 BeforeDelete/AfterDelete 删除钩子
5. 使用 AfterFind 查询钩子
6. 在钩子中实现密码加密、UUID 生成等实战功能

## 前置知识
- 需要先学习：第09-13章（CRUD）、第27章（事务）
- 需要了解：Go 的方法和接口

---

## 28.1 GORM 生命周期

GORM 在执行数据库操作时，会按固定的顺序触发钩子：

```
Create 流程：
  BeforeSave → BeforeCreate → [SQL INSERT] → AfterCreate → AfterSave

Update 流程：
  BeforeSave → BeforeUpdate → [SQL UPDATE] → AfterUpdate → AfterSave

Delete 流程：
  BeforeDelete → [SQL DELETE] → AfterDelete

Query 流程：
  [SQL SELECT] → AfterFind
```

---

## 28.2 完整的钩子列表

| 钩子 | 触发时机 | 典型用途 |
|------|---------|---------|
| `BeforeSave` | Create/Update 之前 | 通用字段校验、数据清理 |
| `AfterSave` | Create/Update 之后 | 缓存更新、日志记录 |
| `BeforeCreate` | INSERT 之前 | 生成 UUID、设置默认值 |
| `AfterCreate` | INSERT 之后 | 记录创建日志、关联创建 |
| `BeforeUpdate` | UPDATE 之前 | 修改时间戳、变更审计 |
| `AfterUpdate` | UPDATE 之后 | 缓存失效、通知 |
| `BeforeDelete` | DELETE 之前 | 删除前校验 |
| `AfterDelete` | DELETE 之后 | 清理关联数据 |
| `AfterFind` | SELECT 之后 | 数据脱敏、格式化 |

---

## 28.3 定义钩子

钩子是定义在模型上的方法，GORM 通过接口检测：

```go
// 钩子方法的几种签名
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// tx 是当前操作的 *gorm.DB 实例
	return nil  // 返回 nil 继续，返回 error 中止
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
	return nil
}
```

---

## 28.4 创建钩子

### BeforeCreate：生成 UUID

```go
type User struct {
	ID   string `gorm:"primaryKey;type:varchar(36)"`
	Name string
	Age  int
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 自动生成 UUID 作为主键
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
```

### BeforeCreate：密码加密

```go
type User struct {
	gorm.Model
	Username string
	Password string
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}
```

### AfterCreate：记录日志

```go
func (u *User) AfterCreate(tx *gorm.DB) error {
	// 记录创建日志
	log.Printf("用户 %s (ID:%d) 被创建", u.Name, u.ID)
	return nil
}
```

---

## 28.5 更新钩子

### BeforeUpdate：更新时间戳

```go
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 自动更新 UpdatedAt（GORM 本身会自动处理，这里只是示例）
	// u.UpdatedAt = time.Now()  // 实际上 gorm.Model 自动做了
	return nil
}
```

### BeforeUpdate：变更审计

```go
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 判断哪些字段发生了变化
	if tx.Statement.Changed("Name") {
		log.Printf("用户 %d 的名字从 %s 改为 %s",
			u.ID, tx.Statement.Dest, u.Name)
	}
	return nil
}
```

---

## 28.6 删除钩子

```go
func (u *User) BeforeDelete(tx *gorm.DB) error {
	// 删除前检查：不允许删除超级管理员
	if u.ID == 1 {
		return fmt.Errorf("不允许删除超级管理员")
	}
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
	// 删除后清理关联数据
	tx.Where("user_id = ?", u.ID).Delete(&Order{})
	return nil
}
```

---

## 28.7 查询钩子

```go
func (u *User) AfterFind(tx *gorm.DB) error {
	// 数据脱敏：隐藏手机号中间四位
	// u.Phone = u.Phone[:3] + "****" + u.Phone[7:]
	
	// 格式化数据
	// u.Avatar = "https://cdn.example.com/" + u.Avatar
	return nil
}
```

> ⚠️ AfterFind 对性能影响最大：每次 SELECT 都会触发。如果数据量大，谨慎处理。

---

## 28.8 跳过钩子

```go
// UpdateColumn / UpdateColumns 跳过所有更新钩子
db.Model(&user).UpdateColumn("status", "active")

// 用 Session 跳过所有钩子
db.Session(&gorm.Session{SkipHooks: true}).Create(&user)
```

---

## 28.9 完整实战：用户注册

```go
package main

import (
	"fmt"
	"log"
	"time"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	UUID      string    `gorm:"type:varchar(36);uniqueIndex"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null;->:false"` // 不可读
	Email     string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 1. 生成 UUID
	if u.UUID == "" {
		u.UUID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), u.ID)
	}

	// 2. 密码加密
	if u.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}
		u.Password = string(hashed)
	}

	// 3. 邮箱默认值
	if u.Email == "" {
		u.Email = u.Username + "@example.com"
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 如果密码被修改了，重新加密
	if tx.Statement.Changed("Password") && u.Password != "" {
		// 判断是否已经是加密过的（bcrypt 密文长度是 60）
		if len(u.Password) != 60 {
			hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			u.Password = string(hashed)
		}
	}
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	// 不允许删除 ID 为 1 的管理员
	if u.ID == 1 {
		return fmt.Errorf("不允许删除系统管理员")
	}
	return nil
}
```

---

## 常见错误

### 错误1：钩子中操作同一张表导致死循环

```go
// ❌ BeforeUpdate 中又 Update 同一条记录
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	tx.Model(u).Update("updated_at", time.Now())  // 又触发 BeforeUpdate！
	return nil  // 无限循环
}
```

### 错误2：钩子中返回 nil 但实际出错

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashed, err := bcrypt.GenerateFromPassword(...)
	if err != nil {
		log.Println(err)
		// ❌ 只打印日志，return nil，密码实际没加密但插入继续
	}
	u.Password = string(hashed)
	return nil
}
// ✅ 应该 return err
```

### 错误3：忘记 UpdateColumn 不触发钩子

```go
// 使用 UpdateColumn 跳过钩子 → BeforeUpdate 中的密码加密不执行！
db.Model(&user).UpdateColumn("password", "new_password")
// 密码以明文存入！非常危险！

// ✅ 用 Update 触发钩子
db.Model(&user).Update("password", "new_password")
```

---

## 本章小结

- 钩子定义在模型上，GORM 自动检测和调用
- 创建：BeforeCreate → AfterCreate
- 更新：BeforeUpdate → AfterUpdate（BeforeSave/AfterSave 两者都触发）
- 删除：BeforeDelete → AfterDelete
- 查询：AfterFind
- 返回 error → 中止操作；返回 nil → 继续
- 跳过钩子：`UpdateColumn`、`Session{SkipHooks: true}`

## 练习题

1. 为 User 模型添加 BeforeCreate 钩子，自动生成 UUID 作为主键。
2. 为 User 模型添加 BeforeCreate 钩子，自动将密码 bcrypt 加密。
3. 为 User 模型添加 BeforeDelete 钩子，禁止删除 ID=1 的超级管理员。
4. 使用 AfterFind 钩子实现手机号脱敏（138****8000）。
5. 比较 `Update` 和 `UpdateColumn` 对钩子的影响，写代码验证。
6. （思考题）为什么 AfterFind 钩子对性能影响最大？在什么情况下你不应该使用它？
