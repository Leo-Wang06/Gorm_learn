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

> 本节假设生成模型中已定义关联（例如 `User.Orders`），并且生成配置包含关联查询所需的 Query API。不同 Gen 版本生成的关联方法签名可能略有差异，请以本项目生成结果为准。

```go
q := query.Use(db)
u := q.User
	o := q.Order

// Preload：预加载关联
users, err := u.Preload(u.Orders).Find()
if err != nil {
	panic(err)
}

// 带条件的 Preload
users, err = u.Preload(u.Orders.On(o.Amount.Gt(100))).Find()
if err != nil {
	panic(err)
}
_ = users

// 关联过滤：也可以用 Join 组合条件；结果形状和分页要按一对一/集合关联区分
users, err = u.Join(o, o.UserID.EqCol(u.ID)).Where(o.Amount.Gt(100)).Distinct().Find()
if err != nil {
	panic(err)
}
_ = users

// Association 操作：先查出父记录，再通过关联 DAO 操作
user, err := u.Where(u.ID.Eq(1)).First()
if err != nil {
	panic(err)
}
if err := u.Orders.Model(user).Append(&model.Order{Amount: 200}); err != nil {
	panic(err)
}
```

---

## 37.2 Gen 的事务

```go
q := query.Use(db)

// q.Transaction 会把回调中的 Query 绑定到同一个事务
err := q.Transaction(func(tx *query.Query) error {
	// 扣库存：Update 返回影响行数和错误
	result, err := tx.Product.Where(
		 tx.Product.ID.Eq(1),
		 tx.Product.Stock.Gte(10),
	).Update(tx.Product.Stock, tx.Product.Stock.Sub(10))
	if err != nil {
		return err
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("库存不足或商品不存在")
	}

	// 创建订单；后续操作必须继续使用 tx，而不是原始 db
	if err := tx.Order.Create(&model.Order{Amount: 200}); err != nil {
		return err
	}
	return nil // 返回 error 时回滚，返回 nil 时提交
})
if err != nil {
	panic(err)
}
```

---

## 37.3 自定义 SQL 方法

Gen 可以生成自定义 SQL 的类型安全方法：

```go
g := gen.NewGenerator(cfg)
g.UseDB(db)

type UserMethod interface {
	// SQL 模板注释放在方法声明前；gen.T 代表当前 Apply 的模型。
	// SELECT * FROM @@table WHERE name = @name
	FindByName(name string) ([]gen.T, error)

	// SELECT * FROM @@table WHERE age > @age
	FindByAgeGt(age int) ([]gen.T, error)
}

// 生成器配置和 g.UseDB(db) 需已完成；ApplyInterface 会为应用的模型生成这些方法。
g.ApplyInterface(func(UserMethod) {}, g.GenerateModel("users"))
// 运行 g.Execute() 后，返回类型会映射为生成的 model.User。
// 例如：users, err := query.User.WithContext(ctx).FindByName("张三")
```

---

## 37.4 实战：用 Gen 重写 CRUD

生成器程序和业务程序通常是两个文件/命令。先运行第36章的生成程序，再在业务程序中使用已生成的 `model` 与 `query` 包；下面只展示业务侧 CRUD。

```go
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"your-project/model"
	"your-project/query"
)

func main() {
	// 业务程序只使用第36章已经生成的代码。
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// === 使用生成的代码 ===
	q := query.Use(db)
	u := q.User

	// Create
	if err := u.Create(&model.User{Name: "张三", Age: 25}); err != nil {
		log.Fatal(err)
	}

	// Read（类型安全！）
	users, err := u.Where(u.Age.Gt(18), u.Status.Eq("active")).
		Order(u.CreatedAt.Desc()).Limit(10).Find()
	if err != nil {
		log.Fatal(err)
	}
	_ = users

	// Update
	if _, err := u.Where(u.ID.Eq(1)).Update(u.Name, "张三三"); err != nil {
		log.Fatal(err)
	}

	// Delete
	if _, err := u.Where(u.ID.Eq(1)).Delete(); err != nil {
		log.Fatal(err)
	}
}
```

---

## 37.5 Gen 的适用场景

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
- 事务中使用 `q.Transaction(func(tx *query.Query) error { ... })`，回调内始终使用 `tx`
- 自定义 SQL 通过接口方法 + 注释实现
- Gen 适合大型项目，小型项目手写即可

## 练习题

1. 用 Gen 实现一个带事务的转账操作。
2. 在 Gen 中定义自定义查询方法（如：FindActiveUsers）。
3. 用 Gen 重写一个之前手写的复杂查询。
4. （思考题）如果项目中同时使用 Gen 和手写 GORM，需要注意什么？
