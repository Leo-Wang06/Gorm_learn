# 第37章：GORM Gen —— 高级用法

## 本章目标
学完本章后，你将能够：
1. 使用 Gen 处理关联关系
2. 在 Gen 中使用事务
3. 编写自定义 SQL 并生成类型安全的查询方法
4. 选择 Gen 的适用场景

## 前置知识
- 需要先学习：第36章（Gen 入门）、第21-27章（关联+事务）

---

## 37.1 Gen 的关联操作

```go
u := query.User
o := query.Order

// Preload：预加载关联
users, _ := u.Preload(u.Orders).Find()

// 带条件的 Preload
users, _ = u.Preload(u.Orders.On(o.Amount.Gt(100))).Find()

// Joins
users, _ = u.Joins(u.Orders).Where(o.Amount.Gt(100)).Find()

// Association 操作
user, _ := u.Where(u.ID.Eq(1)).First()
u.Orders.Model(user).Append(&model.Order{Amount: 200})
```

---

## 37.2 Gen 的事务

```go
q := query.Use(db)

// 自动事务
err := q.Transaction(func(tx *query.Query) error {
	// 扣库存
	_, err := tx.Product.Where(tx.Product.ID.Eq(1)).
		Update(tx.Product.Stock, tx.Product.Stock.Sub(10))
	if err != nil {
		return err
	}
	// 创建订单
	tx.Order.Create(&model.Order{Amount: 200})
	return nil
})
```

---

## 37.3 自定义 SQL 方法

Gen 可以生成自定义 SQL 的类型安全方法：

```go
g := gen.NewGenerator(cfg)
g.UseDB(db)

// 为 User 表生成自定义查询方法
g.ApplyInterface(func(UserMethod) {}, g.GenerateModel("users"))

type UserMethod interface {
	// SELECT * FROM @@table WHERE name = @name
	FindByName(name string) (*model.User, error)

	// SELECT * FROM @@table WHERE age > @age
	FindByAgeGt(age int) ([]*model.User, error)
}
```

---

## 37.4 实战：用 Gen 重写 CRUD

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	g := gen.NewGenerator(gen.Config{
		OutPath:      "./query",
		ModelPkgPath: "./model",
		Mode:         gen.WithDefaultQuery,
	})

	g.UseDB(db)
	g.ApplyBasic(g.GenerateModel("users"), g.GenerateModel("orders"))
	g.Execute()

	// === 使用生成的代码 ===
	q := query.Use(db)
	u := q.User

	// Create
	u.Create(&model.User{Name: "张三", Age: 25})

	// Read（类型安全！）
	users, _ := u.Where(u.Age.Gt(18), u.Status.Eq("active")).
		Order(u.CreatedAt.Desc()).Limit(10).Find()

	// Update
	u.Where(u.ID.Eq(1)).Update(u.Name, "张三三")

	// Delete
	u.Where(u.ID.Eq(1)).Delete()
}
```

---

## 35.5 Gen 的适用场景

| 场景 | 推荐 |
|------|------|
| 大型项目、多人协作 | ✅ Gen（类型安全，减少错误） |
| 快速原型/小项目 | ❌ 手写 GORM 更灵活 |
| 表结构经常变动 | ✅ Gen（重新生成即可） |
| 复杂动态查询 | ⚠️ 混合使用（Gen + 手写） |
| 已有项目引入 | ⚠️ 逐步迁移，不能一步到位 |

---

## 本章小结

- Gen 支持关联操作、事务、自定义 SQL
- 事务中使用 `query.Use(tx)` 传递事务上下文
- 自定义 SQL 通过接口方法 + 注释实现
- Gen 适合大型项目，小型项目手写即可

## 练习题

1. 用 Gen 实现一个带事务的转账操作。
2. 在 Gen 中定义自定义查询方法（如：FindActiveUsers）。
3. 用 Gen 重写一个之前手写的复杂查询。
4. （思考题）如果项目中同时使用 Gen 和手写 GORM，需要注意什么？
