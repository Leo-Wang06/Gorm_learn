# 第10章：查询单个记录 —— First/Take/Last/Find

## 本章目标
学完本章后，你将能够：
1. 分辨并正确使用 `First`、`Take`、`Last`、`Find` 四个查询方法
2. 使用主键、结构体、Map 三种方式进行条件查询
3. 正确处理 `ErrRecordNotFound` 错误
4. 理解并使用 `FirstOrInit` 和 `FirstOrCreate`
5. 避开结构体零值条件的陷阱

## 前置知识
- 需要先学习：第09章（创建数据 Create）
- 需要了解：MySQL SELECT 语句基础
- 需要准备：数据库中已有一些测试数据

---

## 10.1 四大查询方法对比

GORM 提供了四个主要的单次查询方法：

```go
// 准备数据
// 假设 users 表有：{id:1, name:"张三"}, {id:2, name:"李四"}, {id:3, name:"王五"}
```

| 方法 | 行为 | 找不到时 |
|------|------|---------|
| `First(dest)` | 按主键升序，取第一条 | 返回 `ErrRecordNotFound` |
| `Take(dest)` | 无排序，取任意一条 | 返回 `ErrRecordNotFound` |
| `Last(dest)` | 按主键降序，取最后一条 | 返回 `ErrRecordNotFound` |
| `Find(dest)` | 取所有匹配记录 | 返回空切片，不报错 |

### First：第一条

```go
var user User
result := db.First(&user)
// SQL: SELECT * FROM `users` ORDER BY `users`.`id` LIMIT 1
fmt.Println(user.ID)   // 1
fmt.Println(user.Name) // 张三
```

### Take：任意一条

```go
var user User
result := db.Take(&user)
// SQL: SELECT * FROM `users` LIMIT 1（无 ORDER BY）
// 不保证每次取到同一条
```

### Last：最后一条

```go
var user User
result := db.Last(&user)
// SQL: SELECT * FROM `users` ORDER BY `users`.`id` DESC LIMIT 1
fmt.Println(user.ID)  // 3
fmt.Println(user.Name) // 王五
```

### Find：所有匹配记录

```go
var users []User
result := db.Find(&users)
// SQL: SELECT * FROM `users`
fmt.Println(len(users))        // 3
fmt.Println(result.RowsAffected) // 3
```

---

## 10.2 带条件的查询

### 主键查询

```go
// 方式一：直接传主键值
var user User
db.First(&user, 1)
// SQL: SELECT * FROM `users` WHERE `users`.`id` = 1 ORDER BY `users`.`id` LIMIT 1

// 方式二：传字符串主键
db.First(&user, "1")
// SQL: SELECT * FROM `users` WHERE `users`.`id` = '1' ...

// 方式三：内联条件
db.First(&user, "id = ?", 1)
// SQL: SELECT * FROM `users` WHERE id = 1 ...

// 方式四：Where + First
db.Where("id = ?", 1).First(&user)
```

### 主键批量查询

```go
var users []User
db.Find(&users, []int{1, 2, 3})
// SQL: SELECT * FROM `users` WHERE `users`.`id` IN (1,2,3)
```

### 结构体条件查询

```go
// 用结构体作为条件
var user User
db.Where(&User{Name: "张三", Age: 25}).First(&user)
// SQL: SELECT * FROM `users` WHERE name='张三' AND age=25 ORDER BY id LIMIT 1
```

> ⚠️ **零值陷阱**：结构体中的零值字段**不会**成为查询条件！

```go
// ⚠️ 危险示例
db.Where(&User{Name: "张三", Age: 0}).First(&user)
// SQL: SELECT * FROM `users` WHERE name='张三' ORDER BY id LIMIT 1
// Age = 0 的条件被跳过了！
// 因为 0 是 int 的零值，GORM 认为"你没设置 Age"
```

### Map 条件查询（零值安全）

```go
// ✅ Map 方式的零值不会被跳过
db.Where(map[string]interface{}{"name": "张三", "age": 0}).First(&user)
// SQL: SELECT * FROM `users` WHERE name='张三' AND age=0 ORDER BY id LIMIT 1
// Age = 0 正常参与查询！
```

---

## 10.3 `ErrRecordNotFound` 处理

### 什么时候返回这个错误

```go
var user User
result := db.First(&user, 999)  // 假设 id=999 不存在

if result.Error != nil {
	// 判断是否是 RecordNotFound
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("查无此人！")
	} else {
		fmt.Println("其他错误：", result.Error)
	}
}
```

### 各方法的找不到行为

```go
// First：找不到 → ErrRecordNotFound
result := db.First(&user, 999)
// result.Error = gorm.ErrRecordNotFound

// Take：找不到 → ErrRecordNotFound
result := db.Take(&user)
// result.Error = gorm.ErrRecordNotFound（如果表是空的）

// Last：找不到 → ErrRecordNotFound
result := db.Last(&user, 999)
// result.Error = gorm.ErrRecordNotFound

// Find：找不到 → 空切片，不报错！
var users []User
result := db.Find(&users, []int{999})
// result.Error = nil
// len(users) = 0（空切片）
```

> 📌 **重要区别**：`First`/`Take`/`Last` 找不到会报错，`Find` 找不到只返回空切片。这是因为"查一条查不到"是异常，"查多条查不到"可能是正常结果。

### 推荐的错误处理方式

```go
func GetUserByID(db *gorm.DB, id uint) (*User, error) {
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户(id=%d)不存在", id)
		}
		return nil, result.Error
	}
	return &user, nil
}
```

---

## 10.4 `FirstOrInit`：查不到就初始化

`FirstOrInit` 查不到记录时，初始化一个结构体但**不写入数据库**：

```go
var user User
db.FirstOrInit(&user, User{Name: "不存在的用户"})
// SELECT * FROM `users` WHERE name='不存在的用户' ORDER BY id LIMIT 1
// 找不到 → user = User{Name: "不存在的用户"}（只在内存中，未写入数据库）
// user.ID = 0（因为没有写入数据库）
```

### `Attrs` 和 `Assign` 的配合

```go
// Attrs：只在"没找到"时附加的字段
db.Where(User{Name: "张三"}).Attrs(User{Age: 30}).FirstOrInit(&user)
// 找到了 → user.Age 是数据库的值
// 没找到 → user = User{Name: "张三", Age: 30}（Attrs 生效）

// Assign：无论找没找到，都赋值
db.Where(User{Name: "张三"}).Assign(User{Age: 30}).FirstOrInit(&user)
// 找到了 → user.Age 被覆盖为 30
// 没找到 → user = User{Name: "张三", Age: 30}
```

---

## 10.5 `FirstOrCreate`：查不到就创建

和 `FirstOrInit` 类似，但没找到时会**写入数据库**：

```go
var user User
db.FirstOrCreate(&user, User{Name: "张三", Age: 25})
// SELECT * FROM `users` WHERE name='张三' AND age=25 ORDER BY id LIMIT 1
// 如果没找到 → INSERT INTO `users` (name,age) VALUES ('张三',25)
```

### `Attrs` 配合

```go
// Attrs：只在创建时追加字段
db.Where(User{Name: "张三"}).Attrs(User{Age: 30}).FirstOrCreate(&user)
// 找到了 → Age 保持数据库的值
// 没找到 → INSERT ... Age=30

// Assign：无论找到与否都赋值
db.Where(User{Name: "张三"}).Assign(User{Age: 30}).FirstOrCreate(&user)
// 找到了 → UPDATE ... SET age=30
// 没找到 → INSERT ... Age=30
```

---

## 10.6 查询性能提示

```go
// ❌ 不好：查出来才发现用不上
var user User
db.First(&user, 1)
fmt.Println(user.Name)  // 只需要 Name，却查了所有字段

// ✅ 更好：只查需要的字段
var user User
db.Select("name").First(&user, 1)
fmt.Println(user.Name)
// SQL: SELECT `name` FROM `users` WHERE id=1 LIMIT 1

// ✅ 更好：只查一列
var name string
db.Model(&User{}).Where("id = ?", 1).Pluck("name", &name)
// SQL: SELECT `name` FROM `users` WHERE id=1
```

---

## 常见错误

### 错误1：结构体零值条件

```go
// ❌ 想查 age=0 的用户
db.Where(&User{Age: 0}).Find(&users)
// SQL: SELECT * FROM `users`（没有 WHERE 条件！）

// ✅ 用 Map 或字符串条件
db.Where("age = ?", 0).Find(&users)
db.Where(map[string]interface{}{"age": 0}).Find(&users)
```

### 错误2：混淆 First 和 Find 的找不到行为

```go
// ❌ 用 First 查多条条件
var users []User  // 注意：users 是切片
db.First(&users, "age > ?", 18)
// 错误！First 只返回一条，但传了切片

// ✅ 查一条用 First
var user User
db.First(&user, "age > ?", 18)

// ✅ 查多条用 Find
var users []User
db.Find(&users, "age > ?", 18)
```

### 错误3：忘记判断 ErrRecordNotFound

```go
var user User
db.First(&user, 999)
// 没判断错误就直接用
fmt.Println(user.Name)  // 空字符串，但不知道是因为不存在还是名字就是空的
```

### 错误4：Last 的理解误区

```go
// 很多人以为 Last 是"最后一条记录"
// 实际上是"按主键降序的第一条"

// 如果你的主键不是自增 ID，而是 UUID：
type Product struct {
	ID   string `gorm:"primaryKey"`  // UUID 主键
	Name string
}
// Last 的行为取决于 UUID 的排序，可能不是"最新插入的"
```

---

## 本章小结

| 方法 | SQL 行为 | 找不到时 | 适用场景 |
|------|---------|---------|---------|
| `First` | ORDER BY id LIMIT 1 | ErrRecordNotFound | 查第一条记录 |
| `Take` | LIMIT 1 | ErrRecordNotFound | 任意取一条 |
| `Last` | ORDER BY id DESC LIMIT 1 | ErrRecordNotFound | 查最后一条记录 |
| `Find` | 无条件 | 空切片 nil error | 查所有/查多条 |
| `FirstOrInit` | 同上 + 初始化 | 内存初始化 | 确保有值但不写库 |
| `FirstOrCreate` | 同上 + 创建 | 写入数据库 | 确保有值且写库 |

- 主键查询：`db.First(&user, id)`（最简洁写法）
- 零值安全：用字符串条件或 Map，别用结构体
- 错误判断：`errors.Is(result.Error, gorm.ErrRecordNotFound)`

## 练习题

1. 分别用 `First`、`Take`、`Last` 查询 users 表，观察生成的 SQL 和返回结果。
2. 用 `First` 查询一个不存在的 ID，处理 `ErrRecordNotFound`。
3. 用 `Find` 查询不存在的数据，和 `First` 对比行为差异。
4. 写一个 `GetUserByEmail` 函数：用 `First` 按邮箱查找用户，不存在则返回 nil 和自定义错误。
5. 用 `FirstOrCreate` 实现"如果用户不存在就自动注册"的逻辑。
6. （思考题）为什么 GORM 设计成：结构体条件忽略零值，Map 条件不忽略？这个设计背后的原因是什么？
