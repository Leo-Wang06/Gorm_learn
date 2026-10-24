# 第02章：什么是 ORM 与 GORM 概述

## 本章目标
学完本章后，你将能够：
1. 用自己的话解释 ORM 是什么，以及它解决了什么问题
2. 对比手写 SQL 和使用 ORM 的优劣
3. 了解 Go 生态中主流 ORM 的差异
4. 理解 GORM 的核心设计哲学
5. 建立 GORM 的心智模型：struct ↔ 表、字段 ↔ 列、实例 ↔ 行

## 前置知识
- 需要先学习：第01章（Go 语言基础）
- 需要了解：MySQL 基础（数据库、表、SQL 查询——你应该已经学过）

---

## 2.1 什么是 ORM

### 一句人话解释

**ORM（Object-Relational Mapping，对象关系映射）** 是一种技术，它让你用操作"对象"（Go 中的结构体）的方式来操作数据库，而不需要手写 SQL。

打个比方：ORM 就像是一个**翻译官**——
- 你对翻译官说："把这个用户存起来"（操作 Go 结构体）
- 翻译官对数据库说："`INSERT INTO users (name, age) VALUES ('张三', 25)`"（生成 SQL）
- 你想查数据，翻译官帮你执行 `SELECT`，然后把结果翻译成 Go 结构体还给你

```
你的 Go 代码              GORM（翻译官）              数据库
    │                        │                      │
    │ db.Create(&user)       │                      │
    │──────────────────────>│                      │
    │                        │ INSERT INTO users... │
    │                        │─────────────────────>│
    │                        │     OK, ID=1         │
    │                        │<─────────────────────│
    │   user.ID = 1          │                      │
    │<──────────────────────│                      │
```

### 没有 ORM 的世界：手写 SQL 的痛

如果你不用 ORM，只使用 Go 标准库 `database/sql` 操作数据库，代码是这样的：

```go
// 手写 SQL 插入数据（繁琐、容易出错）
result, err := db.Exec(
    "INSERT INTO users (name, age, email, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
    user.Name, user.Age, user.Email, time.Now(), time.Now(),
)
if err != nil {
    return err
}
id, _ := result.LastInsertId()

// 手写 SQL 查询数据（需要手动 Scan 每个字段）
row := db.QueryRow("SELECT id, name, age, email, created_at, updated_at FROM users WHERE id = ?", id)
var u User
err = row.Scan(&u.ID, &u.Name, &u.Age, &u.Email, &u.CreatedAt, &u.UpdatedAt)
```

问题很明显：
1. **写起来累**：每个操作都要手写 SQL 字符串，字段多了手酸
2. **容易出错**：SQL 字符串里的字段名拼错了编译器不报错，运行时才发现
3. **类型不安全**：`Scan` 参数顺序必须和 SELECT 列顺序严格一致，弄反了也不会编译报错
4. **难维护**：表结构变了（加个字段），所有 SQL 字符串都要改

### 有了 ORM 的世界

```go
// GORM 插入（一行搞定）
db.Create(&user)

// GORM 查询（一行搞定）
db.First(&user, id)
```

优势很明显：
- **简洁**：一行代码搞定，不用手写 SQL
- **类型安全**：操作的是 Go 结构体，拼错字段名编译就报错
- **自动映射**：查询结果自动填充到结构体，不用手动 Scan
- **跨数据库**：换数据库（MySQL → PostgreSQL）只需换驱动，业务代码不用变

---

## 2.2 手写 SQL vs ORM 对比

| 维度 | 手写 SQL | ORM（GORM） |
|------|---------|------------|
| 开发速度 | 慢，每个操作都要写 SQL | 快，一行代码完成 CRUD |
| 类型安全 | 差，SQL 字符串拼错编译不报错 | 好，操作结构体，编译期检查 |
| 复杂查询 | 灵活，可以写任意复杂 SQL | 受限于 ORM API，需要学习成本 |
| 性能 | 最优，精确定制每一条 SQL | 有轻微开销，生成的 SQL 可能不是最优 |
| SQL 控制力 | 完全控制 | 大部分控制（也可用 Raw 原生 SQL） |
| 可移植性 | 差，换数据库要重写 SQL | 好，换驱动即可 |
| 学习成本 | 低（会 SQL 就行） | 中高（学 GORM API 相当于学一门新"语言"） |
| 代码量 | 多，模板代码较多 | 少，大幅减少模板代码 |

> **最佳实践**：80% 的场景（简单 CRUD）用 GORM，20% 的场景（复杂报表、性能敏感查询）用 GORM 的 `Raw()` 原生 SQL。**GORM 不是让你完全不写 SQL，而是让你只在必要时写 SQL。**

---

## 2.3 Go 生态中的 ORM 对比

Go 语言主流的数据库操作方案有四种：

| 方案 | 类型 | 特点 | 适用场景 |
|------|------|------|---------|
| **GORM** | 全功能 ORM | 功能最全、社区最大、文档最丰富 | 大多数项目的首选 |
| **ent** | 代码生成 ORM | Facebook 开源，Schema 驱动，类型极安全 | 大型项目、对类型安全要求极高 |
| **sqlx** | SQL 增强库 | 在 `database/sql` 上做薄封装，贴近 SQL | 喜欢手写 SQL 但想减少模板代码 |
| **sqlc** | 代码生成 | 从 SQL 文件生成 Go 代码，极其简单 | 强 SQL 能力团队、微服务 |

### 为什么选 GORM

- ⭐ **GitHub 41k+ Stars**（2024年），Go ORM 中断层第一
- 📚 文档完善（[gorm.io](https://gorm.io)），中文资料丰富
- 🔧 功能全：CRUD、关联、事务、Hook、迁移、插件、Gen 代码生成器...
- 🌍 社区活跃：遇到问题很快能搜到解决方案
- 🚀 简单项目够用，复杂项目也能撑住

---

## 2.4 GORM 的发展历史

| 版本 | 时间 | 核心变化 |
|------|------|---------|
| v1 | 2016-2020 | 创始人 jinzhu 开发，功能完善但架构有局限性 |
| v2 | 2020 至今 | 重写架构，引入 Session 模式、插件系统重构、性能大幅提升 |

> ⚠️ **重要提醒**：v1 和 v2 的 import 路径不同！
> - v1：`github.com/jinzhu/gorm`（已停止维护，不要再用！）
> - v2：`gorm.io/gorm` + `gorm.io/driver/mysql`（本教程使用版本）
>
> **本教程全部基于 GORM v2 编写。**

### v1 vs v2 关键差异

| 特性 | v1 | v2 |
|------|----|----|
| 模块路径 | `github.com/jinzhu/gorm` | `gorm.io/gorm` |
| Session 模式 | 不支持 | 核心特性，链式配置 |
| 插件系统 | 有限 | 重构，更灵活 |
| 批量插入 | 不支持分批 | `CreateInBatches` |
| 预编译缓存 | 不支持 | `PrepareStmt` |
| 性能 | 一般 | 提升 30%+ |
| 错误处理 | 简单 | 更丰富的错误类型 |

---

## 2.5 GORM 的核心设计哲学

### 1. 约定优于配置（Convention over Configuration）

GORM 遵循一套默认约定，你按约定来就不用多写配置：

| 约定 | 说明 | 例子 |
|------|------|------|
| ID 字段 = 主键 | 结构体中名为 `ID` 的字段自动成为主键 | `ID uint` → 主键 |
| CreatedAt/UpdatedAt | 自动管理创建时间和更新时间 | 创建时自动填入当前时间 |
| 表名 = 结构体名复数蛇形 | `User` 结构体 → `users` 表 | `ProductCategory` → `product_categories` |
| 列名 = 字段名蛇形 | `UserName` 字段 → `user_name` 列 | `CreatedAt` → `created_at` |
| 外键 = 关联结构体名 + ID | `Company` 关联 → `CompanyID` 外键 | |

### 2. 链式调用（Method Chaining）

GORM 的查询构建采用链式调用：

```go
db.Where("age > ?", 18).
   Where("status = ?", "active").
   Order("created_at desc").
   Limit(10).
   Find(&users)
```

每次调用返回 `*gorm.DB` 本身，可以一直链下去。这带来一个重要的使用模式：

```go
// 构建"可复用的查询模板"
query := db.Where("status = ?", "active")

// 在不同场景下复用
query.Find(&allActiveUsers)           // 查询所有活跃用户
query.Where("age > ?", 18).Find(&users) // 查询活跃的成年用户
```

### 3. 零值陷阱意识

GORM 默认认为零值字段"没有设置"，这在更新和条件查询时会产生意外行为：

```go
type User struct {
    Name string
    Age  int  // 零值是 0
}

// ⚠️ 危险：Age==0 不会作为条件！
db.Where(&User{Name: "张三", Age: 0}).Find(&users)
// 生成的 SQL: SELECT * FROM users WHERE name = '张三'
// Age = 0 的条件被忽略了！
```

别担心，后续章节会详细讲解如何避免这些陷阱。

---

## 2.6 GORM 支持的数据库

GORM 官方支持四种数据库：

```go
import (
    "gorm.io/driver/mysql"      // MySQL
    "gorm.io/driver/postgres"   // PostgreSQL
    "gorm.io/driver/sqlite"     // SQLite（测试神器！）
    "gorm.io/driver/sqlserver"  // Microsoft SQL Server
)
```

> 📌 本教程以 **MySQL** 为例，因为它是初学者最熟悉的关系型数据库。

换数据库时只需改驱动和 DSN，业务代码几乎不用动：

```go
// MySQL
db, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/mydb"), &gorm.Config{})

// PostgreSQL
db, _ := gorm.Open(postgres.Open("host=localhost user=postgres dbname=mydb"), &gorm.Config{})

// SQLite（测试用）
db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

// 之后的 CRUD 代码完全一样！
db.Create(&user)
db.First(&user)
```

---

## 2.7 建立 GORM 心智模型

学习 GORM 之前，先在脑中建立这张映射表：

| Go 世界 | 数据库世界 | 一句话说明 |
|---------|-----------|-----------|
| 结构体 `type User struct{}` | 表 `users` | 结构体定义了表的结构 |
| 字段 `Name string` | 列 `name VARCHAR(255)` | 每个字段对应一列 |
| 实例 `user := User{}` | 行 `(1, '张三', 25)` | 一个实例代表一行数据 |
| 切片 `[]User{}` | 多行数据 | 切片代表查询结果集 |
| `db.Create(&user)` | `INSERT INTO users ...` | C 操作 → Create |
| `db.First(&user, 1)` | `SELECT * FROM users WHERE id = 1` | R 操作 → First/Find |
| `db.Save(&user)` | `UPDATE users SET ...` | U 操作 → Save/Update |
| `db.Delete(&user)` | `DELETE FROM users ...` | D 操作 → Delete |

**用一句话记住 GORM**：你把 Go 结构体交给 GORM，GORM 自动生成 SQL 发给数据库，然后把结果填回结构体。

---

## 常见错误

### 错误1：用错了 GORM 版本

```go
// ❌ v1 的导入路径（已停止维护！）
import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/mysql"

// ✅ v2 的导入路径
import "gorm.io/gorm"
import "gorm.io/driver/mysql"
```

### 错误2：以为 GORM 完全不需要懂 SQL

很多人以为用了 ORM 就可以完全不懂 SQL，这是最大的误解。GORM 生成的 SQL 如果不理解，出了性能问题完全不知道从哪排查。

**正确态度**：GORM 是帮你写 SQL 的工具，不是替代 SQL 的魔法。开启 SQL 日志，观察每一条生成的 SQL，理解它为什么这样生成。

### 错误3：所有场景都用 GORM

GORM 适合 80% 的 CRUD 场景，但以下情况建议用原生 SQL：
- 复杂的多表统计报表
- 需要数据库特性的操作（如 MySQL 的 `ON DUPLICATE KEY UPDATE`）
- 对性能有极致要求的查询

---

## 本章小结

- ORM = 操作对象 = 操作数据库，GORM 是中间的"翻译官"
- GORM v2 (`gorm.io/gorm`) 是当前版本，v1 (`jinzhu/gorm`) 已停止维护
- GORM 三大哲学：约定优于配置、链式调用、零值陷阱意识
- 心智模型：struct ↔ 表、字段 ↔ 列、实例 ↔ 行
- GORM 让你只在必要时写 SQL，而不是完全不写 SQL

## 练习题

1. 用自己的话解释什么是 ORM，为什么要用它。
2. 列出 GORM v1 和 v2 的三个关键区别。
3. GORM 中 `User` 结构体默认对应什么表名？`ProductCategory` 对应什么？如果不确定，猜猜看，下一章验证。
4. "约定优于配置"在 GORM 中体现在哪些方面？举三个例子。
5. Go 生态中除了 GORM 还有哪些数据库操作方案？各有什么特点？
6. （思考题）如果你需要写一个复杂的年度销售统计报表（涉及 5 张表 JOIN，多个聚合函数），你会用 GORM 还是 Raw SQL？为什么？
