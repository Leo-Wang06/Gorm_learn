# 第12章：更新数据 Update

## 本章目标
学完本章后，你将能够：
1. 区分 `Save`、`Update`、`Updates`、`UpdateColumn`、`UpdateColumns` 的差异
2. 正确处理结构体更新的零值问题
3. 使用 `Select` 和 `Omit` 精确控制更新字段
4. 使用 `gorm.Expr` 实现表达式更新
5. 安全地进行批量更新
6. 根据场景选择正确的更新方法

## 前置知识
- 需要先学习：第09-11章（创建和查询）
- 需要了解：MySQL UPDATE 语句基础

---

## 12.1 更新方法一览

GORM 提供了好几种更新方法，初学者很容易搞混：

| 方法 | 更新字段数 | 是否触发 Hook | 零值处理 |
|------|----------|-------------|---------|
| `Save(&user)` | 全字段 | ✅ | ⚠️ 零值会覆盖原有数据 |
| `Update("col", val)` | 单列 | ✅ | ✅ 正常 |
| `Updates(struct)` | 多列（跳过零值） | ✅ | ⚠️ 零值被跳过 |
| `Updates(map)` | 多列 | ✅ | ✅ 正常（零值不跳过） |
| `UpdateColumn("col", val)` | 单列 | ❌ | ✅ 正常 |
| `UpdateColumns(struct/map)` | 多列 | ❌ | struct 跳过零值，map 不跳过 |

先用一张决策树帮你选择：

```
你是不是需要触发 Hook（如自动更新时间戳）？
├── 是 → 用 Update/Updates
│   ├── 只更新一列 → Update("col", val)
│   ├── 更新多列，有零值 → Updates(map)
│   └── 更新多列，无零值 → Updates(struct)
└── 否 → 用 UpdateColumn/UpdateColumns
    ├── 只更新一列 → UpdateColumn("col", val)
    ├── 更新多列，有零值 → UpdateColumns(map)
    └── 更新多列，无零值 → UpdateColumns(struct)
```

---

## 12.2 `Save`：全字段保存（谨慎使用）

```go
var user User
db.First(&user, 1)  // 查出 id=1 的用户
// user: {ID:1, Name:"张三", Age:25, Email:"zhangsan@example.com"}

user.Name = "张三三"  // 只改了 Name
db.Save(&user)
```

生成的 SQL：
```sql
UPDATE `users` SET 
    `created_at`='2024-01-15 12:00:00',
    `updated_at`='2024-01-15 12:05:00',
    `deleted_at`=NULL,
    `name`='张三三',
    `age`=25,
    `email`='zhangsan@example.com'
WHERE `id`=1
```

所有字段都被更新了（包括没改的）。这在大多数场景下不是你想要的：
- ❌ 不必要的写操作（Age 和 Email 没有变化也被更新）
- ❌ 并发问题（如果两个请求同时 Save，后执行的会覆盖先执行的）
- ❌ 如果 user 中有字段为零值，数据库的原有数据会被覆盖

> 📌 **什么时候用 Save**：你有一个完整的、新的结构体数据要整体替换数据库中的记录。比如从 Excel 导入数据覆盖旧数据。

---

## 12.3 `Update`：更新单列

```go
// 更新一个字段
db.Model(&User{}).Where("id = ?", 1).Update("name", "张三三")
// SQL: UPDATE `users` SET `name`='张三三' WHERE id = 1

// 或者先查出模型再更新
var user User
db.First(&user, 1)
db.Model(&user).Update("name", "张三三")
// SQL: UPDATE `users` SET `name`='张三三' WHERE `id`=1
```

更新的值可以是任何类型：

```go
db.Model(&user).Update("age", 26)
db.Model(&user).Update("is_active", true)
db.Model(&user).Update("birthday", time.Now())
```

---

## 12.4 `Updates`：更新多列（注意零值！）

### struct 方式（零值跳过）

```go
// 更新多列
db.Model(&user).Updates(User{Name: "张三三", Age: 26})
// SQL: UPDATE `users` SET `name`='张三三',`age`=26 WHERE `id`=1

// ⚠️ 零值陷阱！
db.Model(&user).Updates(User{Name: "张三三", Age: 0, Email: ""})
// SQL: UPDATE `users` SET `name`='张三三' WHERE `id`=1
// Age 和 Email 被跳过了！因为 0 和 "" 是零值
```

### Map 方式（零值不跳过）

```go
// ✅ Map 方式，零值正常更新
db.Model(&user).Updates(map[string]interface{}{
	"name":  "张三三",
	"age":   0,      // ✅ 零值正常更新
	"email": "",     // ✅ 空字符串正常更新
})
// SQL: UPDATE `users` SET `name`='张三三',`age`=0,`email`='' WHERE `id`=1
```

### `Select`：强制更新指定字段

```go
// struct 方式 + Select：强制 Age 也更新
db.Model(&user).Select("Name", "Age").Updates(User{Name: "张三三", Age: 0})
// SQL: UPDATE `users` SET `name`='张三三',`age`=0 WHERE `id`=1
// 虽然 Age=0 是零值，但被 Select 指定了，所以会更新
```

### `Omit`：排除字段

```go
// 更新除了 Email 之外的所有字段
db.Model(&user).Omit("Email").Updates(User{Name: "张三三", Age: 26, Email: "new@example.com"})
// SQL: UPDATE `users` SET `name`='张三三',`age`=26 WHERE `id`=1
// Email 被排除了
```

---

## 12.5 `UpdateColumn` / `UpdateColumns`：跳过 Hook

当你不希望触发 `BeforeUpdate`/`AfterUpdate` Hook 时（比如只需要快速改个状态）：

```go
// 单列，跳过 Hook
db.Model(&user).UpdateColumn("status", "verified")
// SQL: UPDATE `users` SET `status`='verified' WHERE `id`=1
// 注意：UpdatedAt 不会自动更新！

// 多列，跳过 Hook
db.Model(&user).UpdateColumns(map[string]interface{}{
	"status": "verified",
	"count":  0,
})
```

---

## 12.6 表达式更新

### `gorm.Expr`：使用数据库函数和表达式

```go
// 年龄加 1
db.Model(&user).Update("age", gorm.Expr("age + ?", 1))
// SQL: UPDATE `users` SET `age`=age + 1 WHERE `id`=1

// 库存扣减
db.Model(&product).Where("stock >= ?", 10).
	Update("stock", gorm.Expr("stock - ?", 10))
// SQL: UPDATE `products` SET `stock`=stock - 10 WHERE stock >= 10

// 多个列用表达式
db.Model(&user).Updates(map[string]interface{}{
	"age":        gorm.Expr("age + ?", 1),
	"login_at":   gorm.Expr("NOW()"),
	"login_count": gorm.Expr("login_count + 1"),
})
```

### 常用表达式场景

```go
// 计数器 +1
db.Model(&article).UpdateColumn("views", gorm.Expr("views + 1"))

// 字符串拼接
db.Model(&user).Update("full_name", gorm.Expr("CONCAT(first_name, ' ', last_name)"))

// 使用 IF 条件
db.Model(&user).Update("status", gorm.Expr("IF(age > 18, 'adult', 'minor')"))

// 当前时间
db.Model(&user).Update("last_login", gorm.Expr("NOW()"))
```

---

## 12.7 批量更新

### 根据条件批量更新

```go
// 给所有 60 岁以上用户标记为"退休"
result := db.Model(&User{}).Where("age >= ?", 60).Update("status", "retired")
// SQL: UPDATE `users` SET `status`='retired' WHERE age >= 60
fmt.Println("影响行数：", result.RowsAffected)  // 比如：5

// 批量更新多列
db.Model(&User{}).Where("age >= ?", 60).Updates(map[string]interface{}{
	"status": "retired",
	"level":  0,
})
```

### 全局更新保护

```go
// ❌ 没有 Where 条件的批量更新会被阻止
result := db.Model(&User{}).Update("status", "active")
// GORM 会报错：ErrMissingWhereClause

// ✅ 方法一：加上 Where 条件
db.Model(&User{}).Where("1 = 1").Update("status", "active")

// ✅ 方法二：使用 AllowGlobalUpdate（显式声明"我知道我在做什么"）
db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&User{}).Update("status", "active")
```

> 📌 这个保护机制防止了 `UPDATE users SET status='active'` 这种"忘写 Where，把全表都改了"的灾难。

---

## 12.8 Save vs Updates vs UpdateColumns 选择流程

```
开始
│
├─ 需要更新全字段（整体替换）？
│   └─ 是 → 用 Save(&user)
│
├─ 需要触发 Hook（更新时间戳等）？
│   ├─ 是 → 用 Update/Updates
│   │     ├─ 只更新一列 → Update("col", val)
│   │     └─ 更新多列 → Updates(struct|map)
│   │
│   └─ 否 → 用 UpdateColumn/UpdateColumns
│         ├─ 只更新一列 → UpdateColumn("col", val)
│         └─ 更新多列 → UpdateColumns(struct|map)
│
└─ 有零值字段需要更新？
    ├─ 是 → 用 Map 或 Select
    └─ 否 → 用 struct
```

---

## 常见错误

### 错误1：Save 覆盖了不应改的字段

```go
// 数据库中有：id=1, name='张三', age=25, balance=100.50

var user User
db.First(&user, 1)  // 查出完整数据
user.Name = "张三三"
db.Save(&user)
// ✅ 这次没问题，因为查出来的 user 有完整的值

// ⚠️ 但如果你手动构造一个只改了 Name 的 user：
partialUser := User{ID: 1, Name: "张三三"}
db.Save(&partialUser)
// ❌ Name 改了，但 Age=0, Balance=0.00 也覆盖了原有数据！
```

### 错误2：忘记 `Model()` 导致无法更新

```go
// ❌ 缺少 Model()
db.Where("id = ?", 1).Update("name", "新名字")
// GORM 不知道要更新哪张表！

// ✅ 指定 Model
db.Model(&User{}).Where("id = ?", 1).Update("name", "新名字")
```

### 错误3：struct Updates 零值不更新

```go
// 我想把用户年龄设为 0（表示未知）
db.Model(&user).Updates(User{Age: 0})
// ❌ SQL: UPDATE `users` SET `updated_at`=... WHERE `id`=1
// Age 没更新！因为 0 是零值

// ✅ 用 Map
db.Model(&user).Updates(map[string]interface{}{"age": 0})
// SQL: UPDATE `users` SET `age`=0 WHERE `id`=1
```

### 错误4：批量更新忘加 Where 条件

```go
// ❌ 意图：更新 id=1 的用户状态
db.Model(&User{}).Update("status", "active")
// 实际错误：ErrMissingWhereClause（GORM v2 保护机制）
// 如果不是这个保护，全表 status 都变成 'active'！

// ✅ 正确
db.Model(&User{}).Where("id = ?", 1).Update("status", "active")
```

---

## 本章小结

| 方法 | Hook | 零值 | 适用场景 |
|------|------|------|---------|
| `Save` | ✅ | 覆盖 | 整体替换数据（谨慎使用） |
| `Update` | ✅ | 正常 | 更新单列 |
| `Updates(struct)` | ✅ | ⚠️ 跳过 | 更新多列（无零值） |
| `Updates(map)` | ✅ | 正常 | 更新多列（有零值） |
| `UpdateColumn` | ❌ | 正常 | 快速更新单列，不需 Hook |
| `UpdateColumns(map)` | ❌ | 正常 | 快速更新多列，不需 Hook |

- 有零值 → 用 Map 或 `Select`
- 不需要 Hook → 用 `UpdateColumn`/`UpdateColumns`
- 表达式更新 → `gorm.Expr`
- 批量更新必须有 Where

## 练习题

1. 分别用 `Update` 和 `Updates` 更新用户信息，观察生成的 SQL 差异。
2. 用 struct 方式尝试把 Age 更新为 0，发生了什么？用 Map 方式重做。
3. 使用 `gorm.Expr` 实现：给所有用户年龄加 1，给文章阅读量加 1。
4. 使用 `Select` 强制更新零值字段。
5. 写一个完整的更新函数：接收用户 ID 和要更新的字段（map），安全更新并返回影响行数。
6. （思考题）为什么 GORM 要设计 `UpdateColumn`（跳过 Hook）？什么场景下你会故意不触发 Hook？
