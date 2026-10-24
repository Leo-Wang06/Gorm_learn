# 第36章：GORM Gen —— 代码生成器入门

## 本章目标
学完本章后，你将能够：
1. 理解 Gen 是什么以及为什么要用
2. 安装和配置 Gen
3. 从数据库表生成模型代码
4. 使用生成的 Query 进行类型安全的查询

## 前置知识
- 需要先学习：第05-20章（模型定义 + CRUD + 查询）
- 需要了解：代码生成的基本概念

---

## 36.1 什么是 GORM Gen

Gen 是 GORM 官方出品的代码生成器，解决手写代码的三个痛点：

```go
// 手写 GORM（字符串字段名，拼错了编译不报错）
db.Where("nmae = ?", "张三")  // "name" 拼成 "nmae" 编译通过，运行时才发现

// Gen 生成的代码（类型安全，拼错编译就报错）
u := query.User
u.Where(u.Name.Eq("张三"))
```

| 特性 | 手写 GORM | Gen |
|------|----------|-----|
| 字段名 | 字符串，拼错编译不报错 | 代码生成，编译期检查 |
| IDE 提示 | 无 | 有完整提示 |
| 类型安全 | 弱 | 强 |
| 学习成本 | API | 额外学习 Gen API |

---

## 36.2 安装和配置

```bash
go get -u gorm.io/gen
```

### 基本配置

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	// 1. 连接数据库
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// 2. 创建生成器
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./query",     // 生成代码输出目录
		ModelPkgPath: "./model",     // 模型代码输出目录
		Mode:         gen.WithDefaultQuery, // 启用默认查询
	})

	// 3. 设置数据库连接
	g.UseDB(db)

	// 4. 生成模型
	g.ApplyBasic(
		g.GenerateModel("users"),
		g.GenerateModel("orders"),
	)

	// 5. 执行生成
	g.Execute()
}
```

运行：
```bash
go run main.go
# 生成代码：
# ./model/users.gen.go
# ./model/orders.gen.go
# ./query/gen.go
```

---

## 36.3 使用生成的代码

```go
import (
	"your-project/query"
	"your-project/model"
)

func main() {
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	q := query.Use(db)  // 使用生成的 Query

	// 创建（类型安全！）
	u := q.User
	u.Create(&model.User{Name: "张三", Age: 25})

	// 查询：WHERE name = '张三'
	users, _ := u.Where(u.Name.Eq("张三")).Find()

	// 查询：WHERE age > 18 AND status = 'active'
	users, _ = u.Where(u.Age.Gt(18), u.Status.Eq("active")).Find()

	// 排序 + 分页
	users, _ = u.Where(u.Age.Gt(18)).Order(u.CreatedAt.Desc()).Limit(10).Find()

	// First
	user, _ := u.Where(u.ID.Eq(1)).First()
}
```

---

## 36.4 Gen 的条件方法

```go
u := query.User

// 等值
u.Where(u.Name.Eq("张三"))
u.Where(u.Name.Neq("张三"))  // Not Eq

// 比较
u.Where(u.Age.Gt(18))   // Greater Than
u.Where(u.Age.Gte(18))  // Greater Than or Equal
u.Where(u.Age.Lt(60))   // Less Than
u.Where(u.Age.Lte(60))  // Less Than or Equal

// 包含
u.Where(u.ID.In(1, 2, 3))
u.Where(u.ID.NotIn(1, 2, 3))

// LIKE
u.Where(u.Name.Like("%张%"))

// BETWEEN
u.Where(u.Age.Between(18, 30))

// NULL
u.Where(u.Email.IsNull())
u.Where(u.Email.IsNotNull())
```

---

## 36.5 Gen vs 手写对比

```go
// == 手写 GORM ==
var users []User
db.Where("name LIKE ? AND age > ? AND status = ?",
	"%张%", 18, "active").
	Order("created_at desc").
	Limit(10).
	Find(&users)

// == Gen ==
u := q.User
users, _ := u.Where(
	u.Name.Like("%张%"),
	u.Age.Gt(18),
	u.Status.Eq("active"),
).Order(u.CreatedAt.Desc()).Limit(10).Find()
```

---

## 本章小结

- Gen 将手写字符串字段名变成类型安全的代码
- `g.GenerateModel("table_name")` 生成模型
- 生成的查询方法：`Eq`、`Gt`、`Lt`、`In`、`Like`、`Between`
- Gen 消除字段名拼写错误、提供 IDE 提示

## 练习题

1. 安装 Gen 并为你的 users 表生成模型和查询代码。
2. 用 Gen 重写一个之前的查询（如：按年龄和状态过滤用户）。
3. 对比 Gen 和手写 GORM 的代码量，分析优劣。
4. （思考题）Gen 生成的代码是"只读不写"的吗？如果你修改了生成的代码，重新生成时会怎样？
