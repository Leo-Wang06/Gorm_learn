# 第14章：原生 SQL 与高级 CRUD 技巧

## 本章目标
学完本章后，你将能够：
1. 使用 `Raw` 执行原生 SQL 查询
2. 使用 `Exec` 执行原生 SQL 写操作
3. 使用 `Scan` 将查询结果映射到任意结构体
4. 使用 `Pluck` 提取单列数据
5. 使用 `Count` 和 `Distinct` 进行计数和去重
6. 使用 `Clauses` 和 `gorm.Expr` 处理复杂场景
7. 使用 DryRun 模式安全地验证 SQL

## 前置知识
- 需要先学习：第09-13章（完整 CRUD）
- 需要了解：MySQL 高级查询（子查询、表达式等）

---

## 14.1 `Raw`：原生 SQL 查询

当 GORM 的链式 API 不够灵活时，直接写 SQL：

```go
// 基础原生查询
var users []User
db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&users)
// 执行你写的 SQL，结果扫描到 users 切片

// 多参数
db.Raw("SELECT * FROM users WHERE name = ? AND age > ?", "张三", 18).Scan(&users)
```

### `Raw` + `Scan` vs `Exec`

```go
// Raw + Scan：用于查询（有返回结果）
var users []User
db.Raw("SELECT * FROM users WHERE status = ?", "active").Scan(&users)

// Exec：用于写操作（INSERT/UPDATE/DELETE，无返回结果）
result := db.Exec("UPDATE users SET status = ? WHERE age > ?", "senior", 60)
fmt.Println(result.RowsAffected)  // 影响行数
```

### 逐行读取（大结果集）

```go
// Row()：单行
row := db.Raw("SELECT name, age FROM users WHERE id = ?", 1).Row()
var name string
var age int
row.Scan(&name, &age)

// Rows()：多行，手动迭代（适合超大结果集）
rows, _ := db.Raw("SELECT name, age FROM users WHERE age > ?", 18).Rows()
defer rows.Close()

for rows.Next() {
	var name string
	var age int
	rows.Scan(&name, &age)
	fmt.Println(name, age)
}
```

> 📌 使用 `Rows()` 逐行迭代时，记得 `defer rows.Close()`，否则连接不会释放。

---

## 14.2 `Exec`：原生 SQL 执行

```go
// DELETE
db.Exec("DELETE FROM users WHERE status = ?", "banned")

// UPDATE
db.Exec("UPDATE users SET login_count = login_count + 1 WHERE id = ?", 1)

// INSERT
db.Exec("INSERT INTO logs (user_id, action, created_at) VALUES (?, ?, NOW())", 1, "login")

// DDL（慎用）
db.Exec("TRUNCATE TABLE temp_data")
```

> 📌 `Exec` 绕过了 GORM 的 Hook 机制。如果你用 GORM 的 `Delete` 有软删除保护，用 `Exec("DELETE ...")` 会直接物理删除。

---

## 14.3 `Scan`：扫描到任意结构体

`Scan` 让你把查询结果映射到任何结构体，不限于 GORM 模型：

```go
// 定义一个非模型结构体
type UserSummary struct {
	Name  string
	Count int
}

// 聚合查询结果扫描到自定义结构体
var summaries []UserSummary
db.Raw("SELECT name, COUNT(*) as count FROM users GROUP BY name").Scan(&summaries)

// 或者用 GORM 链式 API + Scan
db.Model(&User{}).Select("name, COUNT(*) as count").Group("name").Scan(&summaries)

fmt.Println(summaries[0].Name)   // 张三
fmt.Println(summaries[0].Count)  // 3
```

### 扫描到 Map

```go
// 扫描到 []map[string]interface{}
var results []map[string]interface{}
db.Model(&User{}).Select("name, age").Find(&results)

for _, row := range results {
	fmt.Println(row["name"], row["age"])
}
```

---

## 14.4 `Pluck`：提取单列到切片

```go
// 提取所有用户名
var names []string
db.Model(&User{}).Pluck("name", &names)
// SQL: SELECT `name` FROM `users`
fmt.Println(names)  // ["张三", "李四", "王五"]

// 提取所有 ID
var ids []uint
db.Model(&User{}).Pluck("id", &ids)

// 带条件提取
var emails []string
db.Model(&User{}).Where("status = ?", "active").Pluck("email", &emails)
```

> 📌 `Pluck` 比 `Find` 高效：不需要查询所有列，只查你需要的列。

---

## 14.5 `Count`：计数

```go
// 统计总数
var count int64
db.Model(&User{}).Count(&count)
fmt.Println("总用户数：", count)

// 条件计数
db.Model(&User{}).Where("age > ?", 18).Count(&count)
fmt.Println("成年用户数：", count)

// 分组计数
var results []struct {
	Status string
	Count  int64
}
db.Model(&User{}).Select("status, COUNT(*) as count").Group("status").Scan(&results)
```

> 📌 `Count` 必须在 `Find` 之前调用，或者放在一个独立的查询中。`Count` 会清除前面的 `Select`（除了 `Select` 只包含简单字段的情况）。

---

## 14.6 `Distinct`：去重

```go
// 查询所有不重复的年龄值
var ages []int
db.Model(&User{}).Distinct("age").Pluck("age", &ages)

// 查询有多少种不同的状态
var count int64
db.Model(&User{}).Distinct("status").Count(&count)
```

---

## 14.7 `Clauses`：使用数据库特有子句

```go
import "gorm.io/gorm/clause"

// ON CONFLICT（PostgreSQL）
db.Clauses(clause.OnConflict{
	DoNothing: true,
}).Create(&user)

// RETURNING（PostgreSQL）
db.Clauses(clause.Returning{}).Create(&user)

// FOR UPDATE（行锁）
db.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&user)
// SQL: SELECT * FROM users WHERE id = 1 FOR UPDATE
```

---

## 14.8 `gorm.Expr`：SQL 表达式

```go
// 在查询中使用数据库函数
db.Select("name, TIMESTAMPDIFF(YEAR, birthday, NOW()) as age").Find(&users)

// 在 WHERE 中使用表达式
db.Where("LENGTH(name) > ?", 3).Find(&users)

// 在 UPDATE 中使用表达式
db.Model(&user).UpdateColumn("age", gorm.Expr("age + 1"))

// 复杂表达式
db.Model(&user).Updates(map[string]interface{}{
	"full_name": gorm.Expr("CONCAT(?, ' ', ?)", firstName, lastName),
	"score":     gorm.Expr("score + ?", bonus),
})
```

---

## 14.9 DryRun：验证 SQL 不执行

在不确定 GORM 会生成什么 SQL 时，先用 DryRun 验证：

```go
// 创建 DryRun 会话
dryDB := db.Session(&gorm.Session{DryRun: true})

// 执行操作（不会真正操作数据库）
stmt := dryDB.Where("age > ?", 18).Find(&users)

// 获取生成的 SQL
fmt.Println(stmt.Statement.SQL.String())
// 输出：SELECT * FROM `users` WHERE age > 18

// 确认 SQL 没问题后，正式执行
db.Where("age > ?", 18).Find(&users)
```

---

## 常见错误

### 错误1：`Raw` 忘记 `Scan`

```go
// ❌ Raw 执行了但结果没有去处
db.Raw("SELECT * FROM users")
// SQL 执行了，但结果被丢弃

// ✅ Raw + Scan
var users []User
db.Raw("SELECT * FROM users").Scan(&users)
```

### 错误2：`Exec` 用于查询

```go
// ❌ Exec 用于 SELECT
db.Exec("SELECT * FROM users")
// Exec 主要用于写操作（INSERT/UPDATE/DELETE）

// ✅ 查询用 Raw
var users []User
db.Raw("SELECT * FROM users").Scan(&users)
```

### 错误3：原生 SQL 中的 SQL 注入

```go
// ❌ 危险：字符串拼接
name := req.Query("name")
db.Raw("SELECT * FROM users WHERE name = '" + name + "'").Scan(&users)

// ✅ Raw 同样可以使用占位符
db.Raw("SELECT * FROM users WHERE name = ?", name).Scan(&users)
```

### 错误4：`Count` 和 `Find` 顺序弄反

```go
// ❌ 先 Find 再 Count？Count 用的是同一个查询
query := db.Where("age > ?", 18)
query.Find(&users)  // 执行了查询
query.Count(&count)  // 还是同一个查询，但 Count 会清除 Select
// 分开写最安全

// ✅ 分开使用
db.Model(&User{}).Where("age > ?", 18).Find(&users)
db.Model(&User{}).Where("age > ?", 18).Count(&count)
```

---

## 本章小结

- `Raw` + `Scan`：原生 SQL 查询，结果扫描到结构体
- `Exec`：原生 SQL 执行（INSERT/UPDATE/DELETE/DDL）
- `Scan`：把结果映射到**任意**结构体（不限于 GORM 模型）
- `Pluck`：高效提取单列数据到切片
- `Count`：在 `Model` 上调用，统计行数
- DryRun：安全地预览 SQL 而不执行

## 练习题

1. 使用 `Raw` 写一条 JOIN 查询，将结果 `Scan` 到自定义结构体。
2. 使用 `Exec` 批量更新所有 60 岁以上用户的状态。
3. 使用 `Pluck` 提取所有用户的邮箱列表。
4. 使用 `Count` 统计不同状态的用户数量。
5. 使用 DryRun 模式预览一段 GORM 链式调用生成的 SQL。
6. （思考题）什么情况下你应该用 `Raw` 而不是 GORM 链式 API？什么情况下你应该用 GORM 链式 API 而不是 `Raw`？
