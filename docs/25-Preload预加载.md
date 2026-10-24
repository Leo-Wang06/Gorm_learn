# 第25章：Preload 预加载 —— 解决 N+1 问题

## 本章目标
学完本章后，你将能够：
1. 理解 N+1 查询问题及其危害
2. 使用 Preload 正确预加载关联数据
3. 使用条件预加载和嵌套预加载
4. 理解 Preload 与 Joins 的区别和选择
5. 在实际项目中避免 N+1 问题

## 前置知识
- 需要先学习：第21-24章（四种关联关系）
- 需要了解：MySQL JOIN 语法

---

## 25.1 N+1 查询问题

### 什么是 N+1 问题

N+1 问题是最常见的 ORM 性能杀手。看这个例子：

```go
// 查询所有用户
var users []User
db.Find(&users)  // 第 1 条 SQL

// 循环打印每个用户的订单
for _, u := range users {
	var orders []Order
	db.Where("user_id = ?", u.ID).Find(&orders)  // 第 N 条 SQL！
	fmt.Printf("%s 有 %d 个订单\n", u.Name, len(orders))
}
```

**10 个用户 = 11 条 SQL（1 + 10）**

```
SQL 1: SELECT * FROM users                    ← 1次
SQL 2: SELECT * FROM orders WHERE user_id=1   ← N次开始
SQL 3: SELECT * FROM orders WHERE user_id=2
SQL 4: SELECT * FROM orders WHERE user_id=3
...
SQL 11: SELECT * FROM orders WHERE user_id=10  ← N次结束
```

如果 1000 个用户 = 1001 条 SQL！数据库压力巨大。

### 用 Preload 解决

```go
// 一条 SQL 查用户，一条 SQL 查所有关联订单
db.Preload("Orders").Find(&users)
// SQL 1: SELECT * FROM users
// SQL 2: SELECT * FROM orders WHERE user_id IN (1,2,3,...,10)
// 只需 2 条 SQL！
```

---

## 25.2 Preload 基础用法

### 基本预加载

```go
// 加载所有用户及其订单
var users []User
db.Preload("Orders").Find(&users)
// SQL: SELECT * FROM `users`
// SQL: SELECT * FROM `orders` WHERE `orders`.`user_id` IN (1,2,3,...,n)

// 使用关联数据
for _, u := range users {
	fmt.Printf("%s 的订单:\n", u.Name)
	for _, o := range u.Orders {
		fmt.Printf("  #%d: %.2f元\n", o.ID, o.Amount)
	}
}
```

### 预加载多个关联

```go
// 加载用户的订单和个人资料
db.Preload("Orders").Preload("Profile").Find(&users)

// 链式写法
db.Preload("Orders").Preload("Orders.Items").Find(&users)
```

---

## 25.3 条件预加载

### 带条件的预加载

```go
// 只预加载金额大于 100 的订单
db.Preload("Orders", "amount > ?", 100).Find(&users)
// SQL: SELECT * FROM `orders` WHERE user_id IN (1,2,...) AND amount > 100

// 只预加载已支付的订单
db.Preload("Orders", "status = ?", "paid").Find(&users)
```

### 自定义预加载函数

```go
// 更灵活的控制
db.Preload("Orders", func(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "paid").Order("amount desc").Limit(5)
}).Find(&users)
// SQL: SELECT * FROM `orders`
//      WHERE user_id IN (1,2,...) AND status='paid'
//      ORDER BY amount DESC LIMIT 5
```

---

## 25.4 嵌套预加载

```go
// User → Orders → OrderItems → Product
db.Preload("Orders.OrderItems.Product").Find(&users)
// SQL 1: SELECT * FROM users
// SQL 2: SELECT * FROM orders WHERE user_id IN (...)
// SQL 3: SELECT * FROM order_items WHERE order_id IN (...)
// SQL 4: SELECT * FROM products WHERE id IN (...)

// 嵌套 + 条件
db.Preload("Orders", "status = ?", "paid").
   Preload("Orders.OrderItems").
   Preload("Orders.OrderItems.Product").
   Find(&users)
```

> 📌 嵌套预加载每层都是单独的 SQL，不是一条大 JOIN。对于深层嵌套（3-4层），SQL 条数会快速增长，注意合理性。

---

## 25.5 Joins 预加载

Preload 用多条 SQL 加载关联数据。如果你想用一条 SQL（LEFT JOIN）加载关联，用 `Joins`：

```go
// Preload 方式：2条SQL
db.Preload("Company").Find(&users)

// Joins 方式：1条SQL
db.Joins("Company").Find(&users)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `companies` ON `users`.`company_id` = `companies`.`id`
```

### Joins 预加载的写法

```go
// Joins + Preload（混合使用）
db.Joins("Company").Preload("Orders").Find(&users)
// Company 用 JOIN（1条SQL），Orders 用 Preload（+1条SQL）
```

---

## 25.6 Preload vs Joins 对比

| 维度 | Preload | Joins |
|------|---------|-------|
| SQL 条数 | 1 + N（N=关联表数） | 1 条 |
| 适用关联类型 | 全部（BelongsTo/HasOne/HasMany/Many2Many） | 主要 BelongsTo/HasOne |
| 一对多/多对多 | ✅ 正确（ORM 层组装） | ❌ 结果膨胀（行重复） |
| 关联过滤 | ❌ 不能参与主查询过滤 | ✅ 可以在 WHERE 中使用 |
| 关联排序 | ✅ 可以在 func 中自定义 | ❌ 会影响主查询的排序 |

### 选择决策

```
你需要在主查询中过滤关联数据？
├── 是 → 用 Joins
│   例："查询在 Google 工作的用户" → db.Joins("Company").Where("Company.name=?", "Google")
│
└── 否 → 用 Preload
    例："查询所有用户及其公司名称" → db.Preload("Company").Find(&users)
```

---

## 25.7 Preload All

GORM 支持自动预加载所有关联：

```go
// 这个用户模型定义了哪些关联，全部自动预加载
db.Preload(clause.Associations).Find(&users)
// 等价于：
db.Preload("Company").Preload("Orders").Preload("Profile").Find(&users)
```

> ⚠️ 谨慎使用：会加载所有关联（包括不需要的），浪费性能。

---

## 常见错误

### 错误1：N+1 问题

```go
// ❌ N+1 问题
var users []User
db.Find(&users)
for _, u := range users {
	db.Where("user_id = ?", u.ID).Find(&u.Orders)  // 每条记录一次查询
}

// ✅ Preload
db.Preload("Orders").Find(&users)
```

### 错误2：Preload 之后又在循环中查询

```go
// ❌ Preload 了还手动查（白 Preload 了）
db.Preload("Orders").Find(&users)
for _, u := range users {
	db.Where("user_id = ?", u.ID).Find(&u.Orders)  // 多余！
}

// ✅ Preload 后直接用
db.Preload("Orders").Find(&users)
for _, u := range users {
	fmt.Println(len(u.Orders))  // Orders 已经有值了
}
```

### 错误3：一对多用 Joins 导致数据膨胀

```go
// ❌ 一个 User 有 10 个 Order → 结果中 User 重复 10 行
db.Joins("Orders").Find(&users)
// 10个订单=10行，每行都有 User 的数据，内存浪费

// ✅ 一对多用 Preload
db.Preload("Orders").Find(&users)
// User 只有1行，Orders 在嵌套数组中
```

---

## 本章小结

- N+1 问题：循环中查关联 = 1 条主查询 + N 条关联查询
- `Preload("Relation")` 用 2 条 SQL 加载关联数据
- 条件预加载：`Preload("Orders", "status = ?", "paid")`
- 嵌套预加载：`Preload("Orders.OrderItems.Product")`
- Joins 用于需要关联过滤时，Preload 用于纯展示关联数据
- 一对多/多对多 → 必须用 Preload（不能用 Joins）

## 练习题

1. 写一个有 N+1 问题的代码，然后用 Preload 修复它。
2. 用 Preload 加载用户及其订单、订单的商品（三层嵌套）。
3. 用条件 Preload 只加载已支付且金额大于 100 的订单。
4. 用 Joins 查询在"上海"公司工作的所有用户。
5. 分析以下场景该用 Preload 还是 Joins：
   - 展示用户列表及其头像
   - 查询有逾期订单的用户
   - 展示文章列表及所有评论
6. （思考题）三层嵌套 Preload 会生成几条 SQL？如果每层都有 100 条记录，会有什么问题？
