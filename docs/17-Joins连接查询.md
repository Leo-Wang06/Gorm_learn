# 第17章：连接查询 —— Joins

## 本章目标
学完本章后，你将能够：
1. 使用 `Joins` 进行多表连接查询
2. 区分 `Joins` 和 `InnerJoins`
3. 使用关联名 Join 简化代码
4. 将 Join 结果映射到自定义结构体
5. 理解 Joins 与 Preload 的本质区别
6. 进行自连接查询

## 前置知识
- 需要先学习：第11章（WHERE 条件查询）
- 需要了解：MySQL JOIN 语法（INNER JOIN、LEFT JOIN）

---

## 17.1 `Joins`：左外连接

### 字符串 Join

```go
// 基础 LEFT JOIN
db.Joins("LEFT JOIN companies ON users.company_id = companies.id").Find(&users)
// SQL: SELECT `users`.* FROM `users` LEFT JOIN companies ON users.company_id = companies.id

// 多表 JOIN
db.Joins("LEFT JOIN companies ON users.company_id = companies.id").
   Joins("LEFT JOIN departments ON users.dept_id = departments.id").
   Find(&users)
```

### Where 条件中使用 JOIN 的表字段

```go
// 查询在"Google"公司工作的用户
db.Joins("LEFT JOIN companies ON users.company_id = companies.id").
   Where("companies.name = ?", "Google").
   Find(&users)
```

### 扫描到自定义结构体

Join 的结果通常不只是 `users` 表的字段，还包含关联表的字段：

```go
// 自定义结构体接收 Join 结果
type UserWithCompany struct {
	ID          uint
	Name        string
	Age         int
	CompanyID   uint
	CompanyName string  // 来自 companies 表
}

var results []UserWithCompany
db.Model(&User{}).
	Select("users.*, companies.name AS company_name").
	Joins("LEFT JOIN companies ON users.company_id = companies.id").
	Scan(&results)

for _, r := range results {
	fmt.Printf("%s 在 %s 工作\n", r.Name, r.CompanyName)
}
```

---

## 17.2 `InnerJoins`：内连接

```go
// 内连接：只返回"有公司"的用户
db.InnerJoins("Company").Find(&users)
// SQL: SELECT `users`.* FROM `users`
//      INNER JOIN `companies` ON `users`.`company_id` = `companies`.`id`

// 对比 LEFT JOIN
db.Joins("Company").Find(&users)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `companies` ON `users`.`company_id` = `companies`.`id`
```

### 区别

```go
// 数据库内容：
// users: {id:1, name:"张三", company_id:1}, {id:2, name:"李四", company_id:NULL}
// companies: {id:1, name:"Google"}

// LEFT JOIN：返回所有用户（包括没有公司的）
db.Joins("Company").Find(&users)
// 结果：张三（有公司）+ 李四（无公司，公司相关字段为 NULL）

// INNER JOIN：只返回有公司的用户
db.InnerJoins("Company").Find(&users)
// 结果：只有张三
```

---

## 17.3 关联名 Join（配合关联定义）

如果你已经在模型中定义了关联关系（Phase 5 详解），可以用关联名 Join：

```go
type User struct {
	gorm.Model
	Name      string
	CompanyID uint
	Company   Company  // 定义了 BelongsTo 关联
}

type Company struct {
	ID   uint
	Name string
}

// 使用关联名 Join（不需要手写 ON 条件！）
db.Joins("Company").Find(&users)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `companies` AS `Company` ON `users`.`company_id` = `Company`.`id`
```

> 📌 GORM 自动根据关联定义生成 ON 条件。`CompanyID` 是外键，关联到 `Company.ID`。

---

## 17.4 Joins 与 Where 条件

```go
// Join 后对关联表加条件
db.Joins("LEFT JOIN companies ON users.company_id = companies.id").
   Where("companies.name IN ?", []string{"Google", "Apple"}).
   Find(&users)

// 多条件
db.Joins("LEFT JOIN companies ON users.company_id = companies.id").
   Where("users.age > ? AND companies.size = ?", 25, "large").
   Find(&users)
```

### 带条件的关联 Join

```go
// 关联名 Join + 条件
db.Joins("Company", db.Where(&Company{Name: "Google"})).Find(&users)
// SQL: SELECT `users`.* FROM `users`
//      LEFT JOIN `companies` AS `Company` ON `users`.`company_id` = `Company`.`id`
//      AND `Company`.`name` = 'Google'
```

---

## 17.5 自连接查询

自连接是把同一张表 Join 到自己：

```go
type Employee struct {
	ID       uint
	Name     string
	ManagerID *uint      // 上级的 ID
	Manager  *Employee   // 上级
}

// 查询每个员工及其上级
type EmpWithManager struct {
	ID          uint
	Name        string
	ManagerName string
}

var results []EmpWithManager
db.Model(&Employee{}).
	Select("employees.name, employees.manager_id, m.name AS manager_name").
	Joins("LEFT JOIN employees AS m ON employees.manager_id = m.id").
	Scan(&results)
```

---

## 17.6 Joins vs Preload：核心区别

这是 GORM 用户最容易混淆的地方：

| 特性 | Joins | Preload |
|------|-------|---------|
| SQL 条数 | 1 条（JOIN） | 多条（主查询 + 关联查询） |
| 结果组装 | 需要自义结构体 + Scan | 自动填充关联字段 |
| 关联条件查询 | ✅ 可以 | ❌ 不能（Preload 不参与主查询过滤） |
| 一对多/多对多 | ❌ 结果会膨胀 | ✅ 正确组装嵌套数据 |
| 适用场景 | 需要 JOIN 过滤或统计 | 纯展示关联数据 |

```go
// 场景1：需要过滤"在 Google 工作的用户" → 用 Joins
db.Joins("Company").Where("Company.name = ?", "Google").Find(&users)

// 场景2：展示所有用户及其公司信息 → 用 Preload
db.Preload("Company").Find(&users)
```

---

## 常见错误

### 错误1：Join 后 SELECT * 导致字段歧义

```go
// ❌ 两张表都有 name 字段，结果不确定
db.Joins("Company").Find(&users)

// ✅ 用 Select 明确字段来源
db.Select("users.*, companies.name AS company_name").
   Joins("LEFT JOIN companies ON users.company_id = companies.id").
   Scan(&customStruct)
```

### 错误2：用 Joins 做一对多查询导致数据膨胀

```go
// 一个用户有 10 个订单
// Joins 结果：用户数据重复 10 次
db.Joins("Orders").Find(&users)
// ❌ 每个用户出现 10 次（和每个订单组合成一行）

// 一对多 → 用 Preload
db.Preload("Orders").Find(&users)
// ✅ 每个用户出现 1 次，Orders 嵌套在 Orders 字段里
```

### 错误3：忘记 `Model()`

```go
// ❌ GORM 不知道查哪张表
db.Joins("LEFT JOIN companies ON ...").Find(&users)
// users 的类型是 []User，GORM 可以推断，但显式指定更清晰

// ✅ 显式指定
db.Model(&User{}).Joins("LEFT JOIN companies ON ...").Find(&users)
```

---

## 本章小结

- `Joins("LEFT JOIN ...")` 左连接
- `InnerJoins("Company")` 内连接（只返回有关联的记录）
- 关联名 Join：`Joins("Company")` 利用已定义的关联自动生成 ON
- 多表 Join 链式调用多个 `Joins()`
- 自连接：表 Join 自身，需要别名
- Joins 用于过滤/统计，Preload 用于展示（第25章详解）

## 练习题

1. 使用 Joins 查询每个用户及其公司名称。
2. 使用 InnerJoins 查询至少有一笔订单的用户。
3. 使用关联名 Join（如果你已经定义了关联关系）查询用户及其公司。
4. 写一个自连接查询：查询每个员工及其直属上级的姓名。
5. 对比 Joins 和 Preload 的实现差异：分别用它们查询用户-订单数据，观察生成的 SQL。
6. （思考题）什么时候绝对不能用 Joins 而必须用 Preload？为什么？
