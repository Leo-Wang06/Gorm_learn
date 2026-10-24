# 第09章：创建数据 Create

## 本章目标
学完本章后，你将能够：
1. 使用 `Create` 插入单条和多条数据
2. 使用 `CreateInBatches` 分批大批量插入
3. 使用 `Select`/`Omit` 控制插入字段
4. 使用 Map 插入数据
5. 理解 UPSERT（`OnConflict`）的用法
6. 使用 `Save`、`FirstOrCreate` 等辅助方法
7. 避开创建数据的常见陷阱

## 前置知识
- 需要先学习：第05-08章（模型定义与迁移）
- 需要了解：MySQL 的 INSERT 语句基础
- 需要准备：已创建好的表结构

---

## 9.1 `db.Create()` 基础

### 单条插入

```go
type User struct {
	gorm.Model
	Name string
	Age  int
}

// 创建一条记录
user := User{Name: "张三", Age: 25}
result := db.Create(&user)

// 检查结果
fmt.Println("插入ID：", user.ID)           // 1（自动回填的自增 ID）
fmt.Println("影响行数：", result.RowsAffected) // 1
fmt.Println("错误：", result.Error)         // nil（成功则为 nil）
```

生成的 SQL：
```sql
INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`age`) 
VALUES ('2024-01-15 12:00:00','2024-01-15 12:00:00',NULL,'张三',25)
```

> 📌 **重要**：`db.Create(&user)` 必须传指针。因为 GORM 需要把数据库生成的自增 ID 写回到 `user.ID`。传值的话，你的 `user.ID` 永远是 0。

### 查看回填的 ID

```go
user := User{Name: "李四", Age: 30}
fmt.Println("创建前 ID：", user.ID)  // 0

db.Create(&user)
fmt.Println("创建后 ID：", user.ID)  // 2（数据库生成的自增 ID）
```

### 批量插入

```go
// 方式一：传入切片
users := []User{
	{Name: "张三", Age: 25},
	{Name: "李四", Age: 30},
	{Name: "王五", Age: 22},
}
db.Create(&users)
// 一条 INSERT 语句插入所有行
// INSERT INTO `users` (...) VALUES (...), (...), (...)
```

```go
// 方式二：逐个创建
users := []User{
	{Name: "张三", Age: 25},
	{Name: "李四", Age: 30},
}
for _, u := range users {
	db.Create(&u)
}
// 生成多条 INSERT 语句（性能差）
```

> 📌 批量插入用切片传参，GORM 会生成一条 INSERT 多行 VALUES 的 SQL。性能远超逐条插入。

---

## 9.2 `CreateInBatches`：分批插入

当数据量很大时（比如 10 万条），一条 INSERT 放 10 万行 VALUES 会导致：
- SQL 语句过长
- 事务太大
- 内存占用高

```go
// 每批 100 条，分 1000 次插入（10 万条）
db.CreateInBatches(users, 100)
```

```go
// 完整示例：生成并分批插入 10 万条数据
var users []User
for i := 0; i < 100000; i++ {
	users = append(users, User{
		Name: fmt.Sprintf("用户%d", i),
		Age:  rand.Intn(60) + 18,
	})
}
db.CreateInBatches(users, 100)
// 每次插入 100 条，总共 1000 次 INSERT
```

生成的 SQL 日志（部分）：
```sql
INSERT INTO `users` (...) VALUES (...), (...), ...  -- 100条
INSERT INTO `users` (...) VALUES (...), (...), ...  -- 100条
INSERT INTO `users` (...) VALUES (...), (...), ...  -- 100条
-- ... 共1000次
```

---

## 9.3 `Select` 和 `Omit`：控制插入字段

### `Select`：指定要插入哪些字段

```go
type User struct {
	Name  string
	Age   int
	Email string
	Phone string
}

user := User{
	Name:  "张三",
	Age:   25,
	Email: "zhangsan@example.com",
	Phone: "13800138000",
}

// 只插入 Name 和 Age
db.Select("Name", "Age").Create(&user)
// SQL: INSERT INTO `users` (`name`,`age`) VALUES ('张三',25)
// Email 和 Phone 被忽略了
```

### `Omit`：排除哪些字段不插入

```go
// 所有字段都插入，除了 Phone
db.Omit("Phone").Create(&user)
// SQL: INSERT INTO `users` (`name`,`age`,`email`) VALUES ('张三',25,'zhangsan@example.com')
```

### 常见场景

```go
// 场景1：插入时忽略自动管理的时间字段
db.Omit("CreatedAt", "UpdatedAt").Create(&user)

// 场景2：强制插入零值字段
db.Select("Name", "Age").Create(&User{Name: "张三", Age: 0})
// Age 为 0 也会被插入（因为明确 Select 了）

// 场景3：插入时排除密码外的所有字段...反过来更常见
db.Select("Username", "Password").Create(&user)
```

---

## 9.4 使用 Map 插入

当你不使用结构体，而是用动态数据插入时：

```go
// Map 插入：键是列名，值是数据
db.Model(&User{}).Create(map[string]interface{}{
	"Name": "张三",
	"Age":  25,
})
// SQL: INSERT INTO `users` (`name`,`age`) VALUES ('张三',25)

// 多行 Map 插入
db.Model(&User{}).Create([]map[string]interface{}{
	{"Name": "张三", "Age": 25},
	{"Name": "李四", "Age": 30},
})
```

> 📌 Map 插入的优势：零值不会被跳过。`"Age": 0` 会被正常插入，不像结构体那样被忽略。

---

## 9.5 UPSERT：存在则更新，不存在则插入

UPSERT = UPDATE + INSERT。GORM 通过 `clause.OnConflict` 实现：

```go
import "gorm.io/gorm/clause"

// 如果主键（或唯一约束）冲突，则更新指定字段
user := User{ID: 1, Name: "张三（已更新）", Age: 26}

db.Clauses(clause.OnConflict{
	UpdateAll: true,  // 冲突时更新所有字段
}).Create(&user)
```

MySQL 生成的 SQL：
```sql
INSERT INTO `users` (`id`,`name`,`age`,...) VALUES (1,'张三（已更新）',26,...) 
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`),`age`=VALUES(`age`),...
```

### 更精细的控制

```go
// 冲突时只更新指定列
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "email"}},     // 冲突判断列（uniqueIndex）
	DoUpdates: clause.AssignmentColumns([]string{"name", "age"}), // 冲突时更新这些列
}).Create(&user)

// 冲突时什么都不做（忽略）
db.Clauses(clause.OnConflict{
	DoNothing: true,
}).Create(&user)
```

### 批量 UPSERT

```go
users := []User{
	{ID: 1, Name: "张三-改", Age: 26},
	{ID: 2, Name: "李四-改", Age: 31},
}
db.Clauses(clause.OnConflict{
	UpdateAll: true,
}).Create(&users)
```

---

## 9.6 `Save()`：有则更新，无则插入

```go
// 主键为零值 → INSERT
user := User{Name: "新用户"}
db.Save(&user)

// 主键有值 → UPDATE（全字段！）
user2 := User{ID: 1, Name: "更新后的名字"}
db.Save(&user2)
// ⚠️ 全字段更新！即使用户只改了 Name，所有字段都会被更新
```

生成的 SQL：
```sql
-- 主键为零值
INSERT INTO `users` (`name`,`created_at`,`updated_at`) VALUES ('新用户',...)

-- 主键有值
UPDATE `users` SET `name`='更新后的名字',`age`=0,`email`='',... WHERE `id`=1
```

> ⚠️ **Save 的零值危险**：`Save` 是全字段更新！你只设置了 `Name`，但 `Age=0`、`Email=""` 也会被写入数据库，覆盖原有数据。**大多数场景下，用 `Update`/`Updates` 而不是 `Save`。**

---

## 9.7 `FirstOrCreate`：先查后插

```go
// 根据条件查找，找不到就创建
var user User
db.FirstOrCreate(&user, User{Name: "张三", Age: 25})
// 生成：
// SELECT * FROM users WHERE name='张三' AND age=25 ORDER BY id LIMIT 1
// 如果没找到：INSERT INTO users (name,age) VALUES ('张三',25)
```

### `Attrs` vs `Assign`

```go
// Attrs：只在"创建"时使用这些字段
db.Where(User{Name: "张三"}).Attrs(User{Age: 30}).FirstOrCreate(&user)
// 如果找到了：Age 保持原值
// 如果没找到：Age = 30（Attrs 的赋值生效）

// Assign：无论"找到"还是"创建"，都使用这些字段
db.Where(User{Name: "张三"}).Assign(User{Age: 30}).FirstOrCreate(&user)
// 如果找到了：Age 更新为 30
// 如果没找到：Age = 30（创建时也使用）
```

| 方法 | 找到记录时 | 未找到记录时 |
|------|----------|------------|
| `FirstOrCreate`（无 Attrs/Assign） | 返回已有记录 | 用条件创建 |
| `Attrs` | 不修改 | 用 Attrs 的值补充创建 |
| `Assign` | 更新为 Assign 的值 | 用 Assign 的值创建 |

---

## 常见错误

### 错误1：忘记传指针

```go
// ❌ 错误
user := User{Name: "张三"}
db.Create(user)
fmt.Println(user.ID)  // 0！ID 没有回填

// ✅ 正确
db.Create(&user)
fmt.Println(user.ID)  // 1（回填成功）
```

### 错误2：批量插入时传的不是切片

```go
// ❌ 错误
users := []User{{Name: "张三"}, {Name: "李四"}}
db.Create(users)  // 传了切片的值，不是指针
// ❌ users[0].ID 还是 0

// ✅ 正确
db.Create(&users)  // 传切片的指针
// ✅ users[0].ID = 1, users[1].ID = 2
```

### 错误3：Save 覆盖了其他字段

```go
// 数据库中有：id=1, name='张三', age=25, email='zhangsan@example.com'

user := User{ID: 1, Name: "张三-改"}
db.Save(&user)
// ❌ 数据库变为：id=1, name='张三-改', age=0, email=''
// Age 和 Email 被零值覆盖了！

// ✅ 改用 Updates
db.Model(&User{}).Where("id = ?", 1).Update("name", "张三-改")
// 数据库：id=1, name='张三-改', age=25, email='zhangsan@example.com'（不变）
```

### 错误4：OnConflict 没指定冲突列

```go
// ❌ 不明确
db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&user)
// 什么情况下算冲突？不明确

// ✅ 明确冲突判断依据
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "email"}},  // 邮箱冲突时触发
	DoUpdates: clause.AssignmentColumns([]string{"name", "age"}),
}).Create(&user)
```

---

## 本章小结

| 方法 | 用途 | 特点 |
|------|------|------|
| `Create(&user)` | 插入单条 | 最常用，ID 自动回填 |
| `Create(&users)` | 批量插入 | 一条 SQL 完成 |
| `CreateInBatches(users, 100)` | 大批量分批插入 | 防止 SQL 过长 |
| `Select/Omit + Create` | 控制插入字段 | 跳过不需要的列 |
| `Clauses(OnConflict) + Create` | UPSERT | 存在则更新 |
| `Save(&user)` | 有则更新，无则插入 | ⚠️ 全字段操作，谨慎使用 |
| `FirstOrCreate` | 先查后插 | 避免重复数据 |
| `Model().Create(map)` | Map 方式插入 | 动态数据插入 |

## 练习题

1. 写代码插入一条 User 记录，验证 ID 是否回填。
2. 定义 10 个 User 记录，用 `Create` 批量插入，观察生成的 SQL。
3. 使用 `CreateInBatches` 分批插入 1000 条数据，每批 200 条。
4. 使用 `Omit` 插入 User 时排除 `Email` 字段。
5. 使用 `Clauses(OnConflict)` 实现：如果 Email 冲突，更新 Name 和 Age。
6. （思考题）`Save` 为什么危险？在什么情况下可以使用 `Save`？
