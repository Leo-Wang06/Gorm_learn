# 第11章：条件查询 —— WHERE 子句详解

## 本章目标
学完本章后，你将能够：
1. 使用字符串条件、结构体条件、Map 条件进行查询
2. 使用 `Not`、`Or` 组合复杂逻辑条件
3. 使用 BETWEEN、IN、LIKE、IS NULL 等运算符
4. 理解链式 Where 的 AND 组合规则
5. 使用命名参数编写更清晰的查询
6. 彻底避开零值条件陷阱

## 前置知识
- 需要先学习：第10章（First/Take/Last/Find）
- 需要了解：MySQL WHERE 子句基础

---

## 11.1 字符串条件（推荐方式）

```go
// 基础占位符
db.Where("name = ?", "张三").Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三'

// 多个占位符
db.Where("name = ? AND age > ?", "张三", 18).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' AND age > 18

// 直接写完整条件（谨慎：无参数时）
db.Where("status = 'active'").Find(&users)
// SQL: SELECT * FROM `users` WHERE status = 'active'
```

> 📌 **`?` 占位符**：GORM 会自动处理转义，防止 SQL 注入。永远不要用 `fmt.Sprintf` 拼接 SQL！`db.Where(fmt.Sprintf("name = '%s'", name))` 是自杀行为。

### 内联条件

```go
// Find 的第二个参数直接传条件
db.Find(&users, "name = ?", "张三")
// SQL: SELECT * FROM `users` WHERE name = '张三'

// 多条件内联
db.Find(&users, "name = ? AND age > ?", "张三", 18)

// 主键内联
db.Find(&users, 1)       // WHERE id = 1
db.Find(&users, []int{1,2,3})  // WHERE id IN (1,2,3)
```

---

## 11.2 结构体条件 vs Map 条件（零值陷阱）

### 结构体条件

```go
db.Where(&User{Name: "张三", Age: 25}).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' AND age = 25
```

优点：类型安全、IDE 有提示
缺点：**零值字段被忽略**

### Map 条件

```go
db.Where(map[string]interface{}{"name": "张三", "age": 25}).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' AND age = 25
```

优点：零值不会被忽略
缺点：不类型安全（拼错字段名不会编译报错）

### 零值陷阱对比

```go
// 场景：查询 Age=0 的用户

// ❌ 结构体方式：Age=0 被忽略
db.Where(&User{Name: "张三", Age: 0}).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三'
// Age 条件被吞了！

// ✅ Map 方式：Age=0 正常参与查询
db.Where(map[string]interface{}{"name": "张三", "age": 0}).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' AND age = 0

// ✅ 字符串方式：最清晰
db.Where("name = ? AND age = ?", "张三", 0).Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' AND age = 0
```

### 选择指南

| 场景 | 推荐方式 |
|------|---------|
| 简单等值查询，无零值 | 结构体条件 `&User{...}` |
| 包含零值的查询 | Map 或字符串条件 |
| 复杂条件（BETWEEN、LIKE、IN） | 字符串条件 |
| 动态构建条件 | 字符串条件 + 链式调用 |
| 固定条件模板 | 字符串条件 |

> 📌 **最佳实践**：统一使用字符串条件（`db.Where("name = ?", name)`），最灵活、最清晰、最安全。

---

## 11.3 `Not`、`Or` 条件组合

### Not：否定条件

```go
// 查询不是张三的用户
db.Not("name = ?", "张三").Find(&users)
// SQL: SELECT * FROM `users` WHERE NOT name = '张三'

// 查询年龄不是 25 的用户
db.Not("age = ?", 25).Find(&users)

// Not + 结构体
db.Not(&User{Name: "张三"}).Find(&users)
// SQL: SELECT * FROM `users` WHERE name <> '张三'

// Not + IN
db.Not([]int{1, 2, 3}).Find(&users)
// SQL: SELECT * FROM `users` WHERE id NOT IN (1,2,3)
```

### Or：或条件

```go
// 查询叫张三或叫李四的用户
db.Where("name = ?", "张三").Or("name = ?", "李四").Find(&users)
// SQL: SELECT * FROM `users` WHERE name = '张三' OR name = '李四'
```

### 复杂组合：AND 和 OR 的优先级

```go
// 需求：查询 (name='张三' OR name='李四') AND age > 18

// ❌ 错误写法（优先级混乱）
db.Where("name = ?", "张三").Or("name = ?", "李四").Where("age > ?", 18).Find(&users)
// SQL: SELECT * FROM `users` WHERE name='张三' OR name='李四' AND age>18
// 实际含义：name='张三' OR (name='李四' AND age>18)  ← 不是我们想要的！

// ✅ 正确写法：用 Raw 嵌括号
db.Where("(name = ? OR name = ?) AND age > ?", "张三", "李四", 18).Find(&users)
// SQL: SELECT * FROM `users` WHERE (name='张三' OR name='李四') AND age>18
```

---

## 11.4 链式 Where 的 AND 组合

每次调用 `Where()` 都在给查询追加一个 AND 条件：

```go
db.Where("age > ?", 18).
   Where("status = ?", "active").
   Where("city = ?", "北京").
   Find(&users)

// 等价于：
db.Where("age > ? AND status = ? AND city = ?", 18, "active", "北京").Find(&users)

// SQL: SELECT * FROM `users` WHERE age > 18 AND status = 'active' AND city = '北京'
```

### 动态构建条件

```go
func SearchUsers(db *gorm.DB, name string, minAge int, status string, city string) []User {
	query := db.Model(&User{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if minAge > 0 {
		query = query.Where("age >= ?", minAge)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if city != "" {
		query = query.Where("city = ?", city)
	}

	var users []User
	query.Find(&users)
	return users
}
```

---

## 11.5 常用运算符

### IN

```go
// 查询 id 为 1、2、3 的用户
db.Where("id IN ?", []int{1, 2, 3}).Find(&users)
// SQL: SELECT * FROM `users` WHERE id IN (1,2,3)

// 查询状态为 active 或 pending 的用户
db.Where("status IN ?", []string{"active", "pending"}).Find(&users)
```

### BETWEEN

```go
// 查询年龄 18-30 岁的用户
db.Where("age BETWEEN ? AND ?", 18, 30).Find(&users)
// SQL: SELECT * FROM `users` WHERE age BETWEEN 18 AND 30
```

### LIKE

```go
// 模糊查询
db.Where("name LIKE ?", "%张%").Find(&users)
// SQL: SELECT * FROM `users` WHERE name LIKE '%张%'

// 前缀查询
db.Where("name LIKE ?", "张%").Find(&users)
// SQL: SELECT * FROM `users` WHERE name LIKE '张%'

// 后缀查询
db.Where("email LIKE ?", "%@gmail.com").Find(&users)
```

### IS NULL / IS NOT NULL

```go
// 查询邮箱为空的用户
db.Where("email IS NULL").Find(&users)

// 查询邮箱不为空的用户
db.Where("email IS NOT NULL").Find(&users)

// 使用指针类型字段
type User struct {
	DeletedAt *time.Time  // 指针类型，nil 表示未删除
}

// 查询未删除的用户
db.Where("deleted_at IS NULL").Find(&users)

// 查询已删除的用户
db.Where("deleted_at IS NOT NULL").Find(&users)
```

---

## 11.6 命名参数

GORM 支持命名参数，让 SQL 更易读：

```go
// 使用 sql.Named
db.Where("name = @name AND age > @age",
	sql.Named("name", "张三"),
	sql.Named("age", 18),
).Find(&users)

// 使用 Map
db.Where("name = @name AND age > @age",
	map[string]interface{}{
		"name": "张三",
		"age":  18,
	},
).Find(&users)
```

> 📌 命名参数在复杂 SQL 中特别有用：`@name` 比 `?` 更清楚地表达了参数的含义。

---

## 11.7 按主键查询的多种写法

```go
// 写成一万种方式，都行：

var user User

// 1. 最简洁
db.First(&user, 1)

// 2. 标准写法
db.First(&user, "id = ?", 1)

// 3. Where + First
db.Where("id = ?", 1).First(&user)

// 4. Where + 内联
db.Where("id = ?", 1).Find(&user)
// 注意：Find 返回切片，即使只有一条。用 First 更准确

// 5. 多个主键（Find 批量查询）
var users []User
db.Find(&users, []int{1, 2, 3})
```

---

## 常见错误

### 错误1：忘记 `?` 占位符导致 SQL 注入

```go
// ❌ 危险！用户输入直接拼进 SQL
name := req.Query("name")  // 用户输入："' OR '1'='1"
db.Where("name = '" + name + "'").Find(&users)
// SQL: SELECT * FROM users WHERE name = '' OR '1'='1'
// 结果：查出了所有用户！

// ✅ 使用占位符
db.Where("name = ?", name).Find(&users)
// SQL: SELECT * FROM users WHERE name = '\' OR \'1\'=\'1'
// GORM 自动转义了单引号，安全
```

### 错误2：用 `:=` 而不是 `=`（Go 语法混淆）

```go
// ❌ 编译错误
db.Where("age := ?", 18)  // Go 里 := 是短声明，SQL 里不是这回事

// ✅ SQL 用 =
db.Where("age = ?", 18)
```

### 错误3：多个 Or 的优先级问题

```go
// ❌ 意图：查 (城市=北京 OR 城市=上海) AND 年龄>18
db.Where("city = ?", "北京").Or("city = ?", "上海").Where("age > ?", 18)
// 实际SQL：WHERE city='北京' OR city='上海' AND age>18
// AND 优先级高于 OR！实际含义：city='北京' OR (city='上海' AND age>18)

// ✅ 显式用括号
db.Where("(city = ? OR city = ?) AND age > ?", "北京", "上海", 18)
```

### 错误4：Where 不在 Model 之后

```go
// ❌ 不指定 Model，Where 可能不知道字段属于哪个表
db.Where("name = ?", "张三").Find(&users)
// GORM 可以从 Find 的 users 推断出是 User 模型，但不够明确

// ✅ 先指定 Model
db.Model(&User{}).Where("name = ?", "张三").Find(&users)
// 明确告诉 GORM 操作的是 users 表
```

---

## 本章小结

- 字符串条件 `"name = ?"` 是最灵活、最安全的方式
- 结构体条件零值会被忽略，Map 条件不会
- 链式 Where 自动以 AND 连接
- OR 条件组合要注意优先级，显式加括号
- 永远用 `?` 占位符，不要拼接字符串（防止 SQL 注入）
- 命名参数 `@name` 让长 SQL 更易读

## 练习题

1. 用字符串条件查询：姓名是"张三"、年龄大于 18、状态是"active"的用户。
2. 用 Map 条件实现上一题，确保年龄可能为 0 时也能正常查询。
3. 使用 `Not` 和 `Or` 组合查询：不是张三也不是李四的用户。
4. 写一个动态搜索函数：根据用户传入的 name、minAge、maxAge、status 参数（可能为空），构建查询条件。
5. 用 LIKE 查询所有名字中有"张"的用户，并确保不产生 SQL 注入。
6. （思考题）为什么说 `db.Where("name = '" + name + "'")` 是危险的？给出一个具体的攻击例子。
