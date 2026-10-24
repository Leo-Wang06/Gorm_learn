# 第13章：删除数据 Delete

## 本章目标
学完本章后，你将能够：
1. 使用 `Delete` 按主键和按条件删除数据
2. 理解软删除（Soft Delete）的原理和行为
3. 使用 `Unscoped()` 查询已删除数据和执行硬删除
4. 自定义软删除字段
5. 根据业务需求选择软删除或硬删除

## 前置知识
- 需要先学习：第09-12章（Create/Read/Update）
- 需要了解：MySQL DELETE 语句基础
- 需要了解：`gorm.Model` 中的 `DeletedAt` 字段

---

## 13.1 `Delete` 基础

### 按主键删除

```go
// 方式一：传入结构体（主键从结构体的 ID 字段获取）
var user User
db.First(&user, 1)     // 先查出来
db.Delete(&user)        // 再删除
// SQL: DELETE FROM `users` WHERE `users`.`id` = 1

// 方式二：传入结构体 + 主键值
db.Delete(&User{}, 1)
// SQL: DELETE FROM `users` WHERE `users`.`id` = 1

// 方式三：条件删除
db.Delete(&User{}, []int{1, 2, 3})
// SQL: DELETE FROM `users` WHERE `users`.`id` IN (1,2,3)
```

### 按条件批量删除

```go
// 删除所有年龄大于 60 的用户
db.Where("age > ?", 60).Delete(&User{})
// SQL: DELETE FROM `users` WHERE age > 60

// 删除所有状态为 banned 的用户
db.Where("status = ?", "banned").Delete(&User{})
```

> ⚠️ **全表删除的危险**：`db.Where("1 = 1").Delete(&User{})` 会执行 `DELETE FROM users`，删除表中所有行。慎之又慎！

---

## 13.2 软删除（Soft Delete）—— GORM 的核心特性

### 什么是软删除

软删除不是真正从数据库中删除数据，而是给记录打上一个"删除标记"（`deleted_at` 时间戳）：

```
物理删除（Hard Delete）：
  DELETE FROM users WHERE id = 1
  → 数据从磁盘上彻底消失，无法恢复

软删除（Soft Delete）：
  UPDATE users SET deleted_at = NOW() WHERE id = 1
  → 数据还在，只是被标记为"已删除"
```

### GORM 的软删除实现

```go
type User struct {
	gorm.Model  // 包含 DeletedAt gorm.DeletedAt `gorm:"index"`
	Name string
	Age  int
}
```

当你调用 `Delete` 时：

```go
var user User
db.First(&user, 1)
db.Delete(&user)

// 实际执行的 SQL：
// UPDATE `users` SET `deleted_at`='2024-01-15 12:30:00' WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL
// 
// 注意：不是 DELETE FROM，是 UPDATE SET deleted_at！
```

`DeletedAt` 字段有两种状态：

| deleted_at 的值 | 含义 |
|----------------|------|
| `NULL` | 记录未删除（正常数据） |
| 有值（时间戳） | 记录已删除 |

### 软删除对查询的影响

```go
// 普通查询：自动过滤已删除的记录
db.Find(&users)
// SQL: SELECT * FROM `users` WHERE `users`.`deleted_at` IS NULL
// 已删除的记录不会被查出来！

// 所有查询方法都自动过滤：
db.First(&user, 1)
// SQL: SELECT * FROM `users` WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL

db.Where("age > ?", 18).Find(&users)
// SQL: SELECT * FROM `users` WHERE age > 18 AND `users`.`deleted_at` IS NULL
```

### 软删除的"不可逆"特性

```go
// 软删除后不能再"正常删除"一次
var user User
db.First(&user, 1)
db.Delete(&user)   // 第一次：设置 deleted_at

db.Delete(&user)   // 第二次：不会报错，但也不会执行
// 因为 deleted_at 已经不是 NULL，WHERE deleted_at IS NULL 找不到这条记录
```

---

## 13.3 `Unscoped()`：绕过软删除

### 查询已删除的记录

```go
// 查所有记录（包括已删除的）
db.Unscoped().Find(&users)
// SQL: SELECT * FROM `users`（没有 deleted_at IS NULL 条件）

// 查已删除的记录
db.Unscoped().Where("deleted_at IS NOT NULL").Find(&users)
// SQL: SELECT * FROM `users` WHERE deleted_at IS NOT NULL
```

### 硬删除（物理删除）

```go
// 真正删除数据（从磁盘上抹去）
db.Unscoped().Delete(&user, 1)
// SQL: DELETE FROM `users` WHERE `users`.`id` = 1
// 注意：是 DELETE，不是 UPDATE！
```

---

## 13.4 软删除 vs 硬删除的适用场景

| 维度 | 软删除 | 硬删除 |
|------|--------|--------|
| 数据恢复 | ✅ 可恢复 | ❌ 不可恢复 |
| 审计追踪 | ✅ 保留删除记录 | ❌ 无痕迹 |
| 存储空间 | ❌ 数据越积越多 | ✅ 不占多余空间 |
| 查询性能 | ❌ 多一个 WHERE 条件 | ✅ 无额外条件 |
| 唯一约束 | ❌ 软删除后原唯一约束仍占用 | ✅ 删除后唯一约束释放 |
| 实现复杂度 | ✅ GORM 内置支持 | ✅ 简单 |

### 哪些表适合软删除

- ✅ 用户表（用户注销后可能恢复）
- ✅ 订单表（订单取消不等于删除）
- ✅ 文章/评论（需要保留历史记录）

### 哪些表适合硬删除

- ❌ 日志表（定期清理旧数据）
- ❌ 验证码表（过期就删）
- ❌ 临时数据（缓存、会话）

### 唯一约束的软删除问题

```go
type User struct {
	gorm.Model
	Email string `gorm:"uniqueIndex"`  // 邮箱唯一
}

// 张三删除账号（软删除）
db.Delete(&zhangsan)
// zhangsan@example.com 的 deleted_at 设为时间戳

// 张三想重新注册同一个邮箱
db.Create(&User{Email: "zhangsan@example.com"})
// ❌ 报错！唯一索引冲突！
// 因为 deleted_at IS NULL 是唯一索引的一部分（不同），
// 但 deleted_at IS NOT NULL 的记录也占着邮箱值
```

解决：手动清理或用复合唯一索引 `(email, deleted_at)`。

---

## 13.5 自定义软删除字段

如果你的表已经有自己的删除标记字段，可以不用 `gorm.Model`：

```go
// 方式一：用 gorm.DeletedAt
type Article struct {
	ID        uint
	IsDel     gorm.DeletedAt `gorm:"softDelete"`  // 自定义名字 ISoftDelete
	Title     string
}

// 方式二：用 int 类型（0=正常, 1=已删除）
type Article struct {
	ID       uint
	IsDel    int `gorm:"softDelete:flag"`  // 0=正常, 1=删除
	Title    string
}

// 方式三：用 time.Time（Unix 时间戳 0 = 正常）
type Article struct {
	ID        uint
	DeletedAt int64 `gorm:"softDelete:milli"`  // 毫秒时间戳
	Title     string
}
```

---

## 13.6 删除后的恢复

```go
// 软删除 → 恢复（把 deleted_at 设回 NULL）
var user User
db.Unscoped().First(&user, 1)  // 找到已删除的记录

// 方法一：直接改字段
db.Unscoped().Model(&user).Update("deleted_at", nil)

// 方法二：使用 Unscoped
db.Unscoped().Model(&user).Update("deleted_at", gorm.Expr("NULL"))
```

---

## 常见错误

### 错误1：忘记 Unscoped 导致"找不到"已删除记录

```go
// ❌ 普通查询找不到已删除的记录
var user User
db.First(&user, 1)  // 如果这条记录被软删了 → ErrRecordNotFound

// ✅ 用 Unscoped 查询
db.Unscoped().First(&user, 1)  // 能找到
```

### 错误2：以为 Delete 返回 ErrRecordNotFound

```go
// Delete 找不到记录不会报错！
result := db.Delete(&User{}, 999)  // id=999 不存在
fmt.Println(result.Error)           // nil（不是 ErrRecordNotFound！）
fmt.Println(result.RowsAffected)    // 0（影响 0 行，说明没删到）
```

> 📌 判断删除是否成功用 `RowsAffected`，不是 `Error`。

### 错误3：批量删除时没加 Where

```go
// ❌ 忘记加条件
db.Delete(&User{})
// SQL: UPDATE `users` SET `deleted_at`=... WHERE `deleted_at` IS NULL
// 把所有正常记录全软删了！

// ✅ 加上条件
db.Where("status = ?", "banned").Delete(&User{})
```

### 错误4：软删除后唯一约束冲突

```go
// 用户邮箱是唯一的，软删除后无法用相同邮箱重新注册
// 解决方案见 13.4 节
```

---

## 本章小结

- 软删除：`UPDATE SET deleted_at = NOW()`，数据还在
- 硬删除：`DELETE FROM`，用 `Unscoped().Delete()` 执行
- 普通查询自动过滤已删除记录（`WHERE deleted_at IS NULL`）
- `Unscoped()` 绕过软删除保护
- 判断删除成功用 `RowsAffected`，不是 `Error`
- `gorm.Model` 自带 `DeletedAt` 字段，自动获得软删除能力

## 练习题

1. 创建一个用户，然后删除它，观察数据库 `deleted_at` 字段的变化。
2. 用 `Find` 查询所有用户，看软删除的用户是否被自动过滤。
3. 用 `Unscoped()` 查出已删除的用户，并验证。
4. 使用 `Unscoped().Delete()` 硬删除一条记录，确认数据从数据库中消失。
5. 写一个"回收站"功能：查询所有已删除的用户，并提供恢复功能。
6. （思考题）软删除的优缺点分别是什么？在什么场景下绝对不能用软删除？
