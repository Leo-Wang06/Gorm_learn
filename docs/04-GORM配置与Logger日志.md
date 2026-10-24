# 第04章：GORM 配置与 Logger 日志

## 本章目标
学完本章后，你将能够：
1. 理解 `gorm.Config{}` 中每个配置项的作用
2. 根据需求选择合适的命名策略（单复数、前缀、自定义规则）
3. 配置 GORM 的日志级别和输出格式
4. 使用 DryRun 模式预览生成的 SQL 而不执行
5. 自定义 Logger，将 GORM 日志输出到文件
6. 在生产环境和开发环境选择合适的配置

## 前置知识
- 需要先学习：第03章（安装 GORM 与连接数据库）
- 需要了解：基本的日志概念

---

## 4.1 `gorm.Config{}` 完整配置项

`gorm.Open()` 的第二个参数是一个配置对象，控制 GORM 的全局行为：

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    // 以下是所有常用配置项
})
```

### 配置项总览

| 配置项 | 类型 | 默认值 | 用途 |
|--------|------|--------|------|
| `SkipDefaultTransaction` | `bool` | `false` | 是否跳过默认事务 |
| `NamingStrategy` | `Namer` | 默认策略 | 表名/列名的命名规则 |
| `DisableForeignKeyConstraintWhenMigrating` | `bool` | `false` | 迁移时是否创建物理外键 |
| `PrepareStmt` | `bool` | `false` | 是否缓存预编译语句 |
| `DryRun` | `bool` | `false` | 只生成 SQL 不执行 |
| `Logger` | `logger.Interface` | 默认 Logger | 日志输出配置 |
| `NowFunc` | `func() time.Time` | `time.Now` | 自定义当前时间函数 |
| `DisableAutomaticPing` | `bool` | `false` | 是否禁止自动 Ping |
| `AllowGlobalUpdate` | `bool` | `false` | 是否允许全局更新 |
| `QueryFields` | `bool` | `false` | 是否用字段名做查询条件 |
| `CreateBatchSize` | `int` | `0` | 默认批量插入大小 |
| `TranslateError` | `bool` | `false` | 是否翻译数据库错误 |
| `ClauseBuilders` | `map` | — | 自定义 Clause 构建器 |
| `ConnPool` | `ConnPool` | — | 自定义连接池 |
| `Dialector` | `Dialector` | — | 自定义方言 |
| `Plugins` | `map[string]Plugin` | — | 注册插件 |

别被这么多配置项吓到。**90% 的情况下，你只需要关心前 6 个。** 我们来逐一详解。

---

## 4.2 `SkipDefaultTransaction`：关闭默认事务

**这是 GORM v2 最重要的性能配置之一。**

```go
// 默认行为（SkipDefaultTransaction = false）
// 每个写操作（Create/Update/Delete）都会在事务中执行
db.Create(&user)

// 等价于：
// BEGIN
// INSERT INTO users ...
// COMMIT

// 优化后（SkipDefaultTransaction = true）
// 写操作直接执行，不包裹事务
// INSERT INTO users ...
```

**什么时候关**：
- 大部分的读写操作不需要事务
- 关闭后写入性能提升约 **30%**
- 需要事务时再手动用 `db.Transaction()` 开启

**什么时候不关**：
- 你的应用和数据一致性要求极高
- 不想操心事务管理，让 GORM 全自动

```go
// 推荐配置
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    SkipDefaultTransaction: true,  // 关闭自动事务
})
```

> 📌 **建议从一开始就设为 `true`**，需要事务的地方手动管理。第27章会详解事务。

---

## 4.3 `NamingStrategy`：命名策略

命名策略控制 Go 结构体名 → 数据库表名、结构体字段名 → 数据库列名的转换规则。

### 默认命名规则（不加任何配置）

```go
type User struct {}            // → 表名：users（复数蛇形）
type ProductCategory struct {} // → 表名：product_categories

type User struct {
    ID        uint      // → 列名：id
    UserName  string    // → 列名：user_name（蛇形）
    CreatedAt time.Time // → 列名：created_at
}
```

### 自定义命名策略

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    NamingStrategy: schema.NamingStrategy{
        TablePrefix:   "t_",       // 表名前缀：users → t_users
        SingularTable: true,       // 禁用复数：User → user（单数）
        NameReplacer:  strings.NewReplacer("CID", "Cid"), // 自定义替换
        NoLowerCase:   false,      // 是否禁用小写（默认不禁止，即转小写）
    },
})
```

### 三个最常用的配置

```go
// 场景一：公司要求表名前缀（如 t_user, t_order）
NamingStrategy: schema.NamingStrategy{
    TablePrefix: "t_",
}

// 场景二：不喜欢复数表名（User → user 而非 users）
NamingStrategy: schema.NamingStrategy{
    SingularTable: true,
}

// 场景三：组合使用
NamingStrategy: schema.NamingStrategy{
    TablePrefix:   "tb_",
    SingularTable: true,
}
```

### 覆盖单张表的命名：`TableName()` 方法

有时你不想改全局规则，只想给特定表取个特殊名字：

```go
type User struct {
    ID uint
}

// 实现 TableName 方法（优先级最高！）
func (User) TableName() string {
    return "sys_users"  // 不管全局规则，这个表就叫 sys_users
}
```

---

## 4.4 `PrepareStmt`：预编译缓存

```go
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    PrepareStmt: true,  // 开启预编译缓存
})
```

**开启后的效果**：

```go
// 第一次执行：编译 SQL 并缓存
db.Where("name = ?", "张三").Find(&users)
// 生成：SELECT * FROM users WHERE name = ?
// 缓存这条 SQL

// 第二次执行：直接用缓存的 SQL，只换参数
db.Where("name = ?", "李四").Find(&users)
// 直接执行缓存的 SQL，参数换成 "李四"
```

优点：
- 减少 SQL 解析开销
- 提高 SQL 安全性（预编译天然防 SQL 注入）

缺点：
- 会占用一些内存（缓存全部预编译的 SQL 变体）
- 每种不同的 SQL 都会缓存一份（如果动态 SQL 很多，缓存会很大）

> 📌 **建议**：中大型项目开启，小项目或 SQL 变体特别多的场景可以不开。

---

## 4.5 `DryRun`：只生成 SQL 不执行

DryRun 是学习和调试 GORM 的**超级神器**：

```go
// 全局开启 DryRun（所有操作只生成 SQL，不执行）
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    DryRun: true,
})

// 执行"查询"
var users []User
stmt := db.Where("age > ?", 18).Find(&users)
// 数据库没有任何变化！但你可以看到生成的 SQL
fmt.Println(stmt.Statement.SQL.String())
// 输出：SELECT * FROM users WHERE age > 18
```

更常用的方式是在 Session 级别开启（推荐）：

```go
// 创建一个 DryRun 会话
dryDB := db.Session(&gorm.Session{DryRun: true})

// 验证生成的 SQL
stmt := dryDB.Create(&User{Name: "张三", Age: 25})
fmt.Println(stmt.Statement.SQL.String())
// 输出：INSERT INTO `users` (`name`,`age`,`created_at`,`updated_at`) VALUES ('张三',25,'2024-01-15 12:00:00','2024-01-15 12:00:00')
```

> 📌 **开发技巧**：不确认 GORM 会生成什么 SQL 时，先用 DryRun 跑一遍，确认 SQL 无误再正式执行。

---

## 4.6 Logger 日志配置

### 默认 Logger 的行为

GORM v2 默认的日志行为：
- 慢查询阈值：200ms（超过这个时间的 SQL 会打 Warn 日志）
- 日志级别：Warn（只打印慢查询和错误）
- 输出位置：标准输出（控制台）
- 彩色：支持

### 四种日志级别

```go
import "gorm.io/gorm/logger"

// Silent：静默模式，什么 SQL 都不打印
logger.Silent

// Error：只打印出错的 SQL
logger.Error

// Warn：打印慢查询和出错的 SQL（默认）
logger.Warn

// Info：打印所有 SQL（开发调试必备）
logger.Info
```

### 在 Config 中设置

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),  // 打印所有 SQL
})
```

### 自定义 Logger 配置

```go
newLogger := logger.New(
    log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（输出目标）
    logger.Config{
        SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
        LogLevel:                  logger.Info,            // 日志级别
        IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 错误
        Colorful:                  true,                   // 彩色输出
    },
)

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: newLogger,
})
```

### 日志输出示例

当 LogLevel = Info 时，你会看到类似这样的输出：

```
2024/01/15 12:00:00 /path/to/main.go:25
[1.234ms] [rows:1] INSERT INTO `users` (`name`,`age`,`created_at`,`updated_at`) VALUES ('张三',25,'2024-01-15 12:00:00','2024-01-15 12:00:00')

2024/01/15 12:00:01 /path/to/main.go:30
[0.567ms] [rows:3] SELECT * FROM `users` WHERE age > 18
```

每行日志的含义：

| 部分 | 含义 |
|------|------|
| `2024/01/15 12:00:00` | 执行时间 |
| `/path/to/main.go:25` | 调用 GORM 的源文件和行号 |
| `[1.234ms]` | SQL 执行耗时 |
| `[rows:1]` | 影响或返回的行数 |
| `INSERT INTO ...` | 实际执行的 SQL 语句 |

> 📌 **开发时务必开启 Info 模式！** 看到 GORM 实际生成的 SQL 是理解其行为的最佳途径。

---

## 4.7 完整配置示例

### 开发环境配置

```go
package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"

	// 开发环境：打印所有SQL，彩色输出
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,  // 关闭默认事务，提升性能
		PrepareStmt:            false, // 开发时不缓存，方便调试
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Info,  // 开发：打印所有 SQL
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		log.Fatal(err)
	}

	// ... 你的业务代码
	_ = db
}
```

### 生产环境配置

```go
// 生产环境：日志输出到文件，只打印错误和慢查询
file, _ := os.OpenFile("gorm.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	SkipDefaultTransaction: false, // 生产环境保持事务一致性
	PrepareStmt:            true,  // 开启预编译，提升性能
	Logger: logger.New(
		log.New(file, "\r\n", log.LstdFlags), // 日志写文件
		logger.Config{
			SlowThreshold:             100 * time.Millisecond, // 生产更严格
			LogLevel:                  logger.Warn,             // 只打 Warn 和 Error
			IgnoreRecordNotFoundError: true,                    // 忽略 RecordNotFound
			Colorful:                  false,                   // 文件中不要彩色
		},
	),
})
```

---

## 常见错误

### 错误1：日志太吵（生产环境开了 Info）

```
// ❌ 生产环境打印所有 SQL
Logger: logger.Default.LogMode(logger.Info)
// 结果：每秒几百条日志，磁盘爆炸，性能直线下降

// ✅ 生产环境只打 Warn
Logger: logger.Default.LogMode(logger.Warn)
```

### 错误2：不打印 SQL 导致调试困难

```
// ❌ 开发环境日志级别太低
Logger: logger.Default.LogMode(logger.Silent)
// 结果：SQL 执行了啥完全不知道，bug 查不到

// ✅ 开发环境用 Info
Logger: logger.Default.LogMode(logger.Info)
```

### 错误3：忘记配置命名策略导致表名对不上

```go
// 定义结构体
type User struct {}

// 期望表名：user（单数）
// 实际表名：users（GORM 默认复数）

// ✅ 需要单数就配置
NamingStrategy: schema.NamingStrategy{
    SingularTable: true,
}
```

### 错误4：DryRun 忘了关

```go
// 测试完 DryRun，忘了去掉配置
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
    DryRun: true,  // ⚠️ 忘了改成 false
})

// 结果：所有数据操作全都不生效，查了半天不知道为啥
db.Create(&user)  // SQL 生成了但没执行！user 没存进数据库
```

**建议**：不要在 Config 中全局开 DryRun，用 Session 临时开。

---

## 本章小结

| 配置项 | 开发环境推荐 | 生产环境推荐 | 为什么 |
|--------|------------|------------|--------|
| `SkipDefaultTransaction` | `true` | `false` 或 `true` | 开发求快，生产求稳 |
| `PrepareStmt` | `false` | `true` | 开发常改代码，生产追求性能 |
| `Logger.LogLevel` | `Info` | `Warn` | 开发要看 SQL，生产要省资源 |
| `NamingStrategy` | 按需 | 按需 | 和团队约定一致 |
| `DryRun` | Session 级别 | 不用 | 只在调试时用 |

## 练习题

1. 配置一个开发环境的 GORM，要求：打印所有 SQL、彩色输出、关闭默认事务。
2. 配置 `NamingStrategy` 使表名满足以下规则：前缀 `t_`、单数（User → t_user）。
3. 使用 DryRun 模式预览以下操作生成的 SQL：
   - `db.Create(&User{Name: "测试"})`
   - `db.Where("age > ?", 18).Find(&users)`
4. 四种日志级别分别适用于什么场景？
5. 写一个程序，将 GORM 日志输出到 `app.log` 文件中，日志级别为 Warn，慢查询阈值设为 100ms。
6. （思考题）为什么开发环境推荐 `SkipDefaultTransaction: true`，而生产环境要考虑 `false`？提示：想想数据一致性。
