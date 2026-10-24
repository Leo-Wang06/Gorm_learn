# 第03章：安装 GORM 与连接数据库

## 本章目标
学完本章后，你将能够：
1. 使用 `go get` 安装 GORM v2 和 MySQL 驱动
2. 理解 DSN（数据源名称）的格式和每个参数的含义
3. 编写代码连接 MySQL 数据库并验证连接成功
4. 理解 `sql.DB` 和 `gorm.DB` 的关系
5. 配置基础的连接池参数
6. 排查常见的数据库连接错误

## 前置知识
- 需要先学习：第01章（Go 语言基础）、第02章（ORM 概念）
- 需要了解：MySQL 已安装并可以正常连接
- 需要准备：一个已创建的 MySQL 数据库（用于测试）

---

## 3.1 创建项目并安装 GORM

### 第一步：创建项目目录

```bash
mkdir gorm-tutorial
cd gorm-tutorial
```

### 第二步：初始化 Go Module

```bash
go mod init gorm-tutorial
```

执行后生成 `go.mod` 文件：

```
module gorm-tutorial

go 1.21
```

> **什么是 Go Module**：Go Module 是 Go 的依赖管理工具，类似于 npm（Node.js）或 pip（Python）。`go.mod` 记录了你项目依赖了哪些第三方包及其版本。

### 第三步：安装 GORM 和 MySQL 驱动

```bash
# 安装 GORM 核心库
go get -u gorm.io/gorm

# 安装 MySQL 驱动
go get -u gorm.io/driver/mysql
```

安装完成后，查看 `go.mod`：

```
module gorm-tutorial

go 1.21

require (
	gorm.io/driver/mysql v1.5.x  // 具体版本可能不同
	gorm.io/gorm v1.25.x         // 具体版本可能不同
)
```

同时会生成 `go.sum` 文件，记录了每个依赖的校验和（用于安全检查，不需要手动编辑）。

### 为什么要装两个包？

| 包 | 作用 | 类比 |
|----|------|------|
| `gorm.io/gorm` | GORM 框架核心，提供所有 ORM 功能 | 汽车引擎 |
| `gorm.io/driver/mysql` | MySQL 驱动，让 GORM 能连接 MySQL | 车轮（适配 MySQL） |

换数据库只需换驱动包：`gorm.io/driver/postgres`、`gorm.io/driver/sqlite` 等。

---

## 3.2 创建数据库

在 MySQL 中创建一个测试用的空数据库：
5
```sql
-- 连接到 MySQL
mysql -u root -p

-- 创建数据库
CREATE DATABASE gorm_tutorial CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 确认创建成功
SHOW DATABASES LIKE 'gorm_tutorial';
```

---

## 3.3 连接数据库：完整的代码

创建 `main.go`：

```go
package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 第一步：定义 DSN（数据源名称）
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"

	// 第二步：打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印所有 SQL
	})
	if err != nil {
		log.Fatal("数据库连接失败：", err)
	}

	// 第三步：获取底层的 sql.DB，配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取 sql.DB 失败：", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(25)                  // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大存活时间
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)  // 空闲连接最大存活时间

	// 第四步：测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库 Ping 失败：", err)
	}

	fmt.Println("🎉 数据库连接成功！")
	fmt.Printf("连接池配置：最大打开=%d, 最大空闲=%d\n",
		sqlDB.Stats().MaxOpenConnections,
		// 注意：MaxOpenConnections 的返回值需要在连接建立后才能正确反映
	)
}
```

运行：

```bash
go run main.go
```

如果看到 `🎉 数据库连接成功！`，恭喜你，GORM 环境搭建完成！

> ⚠️ **记得修改 DSN**：把 `root:123456` 换成你自己的 MySQL 用户名和密码，`gorm_tutorial` 换成你创建的数据库名。

---

## 3.4 DSN 详解

DSN（Data Source Name，数据源名称）是一串包含连接信息的字符串。格式如下：

```
用户名:密码@tcp(主机:端口)/数据库名?参数1=值1&参数2=值2
```

### DSN 各部分拆解

以 `root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local` 为例：

| 部分 | 值 | 说明 |
|------|-----|------|
| 用户名 | `root` | MySQL 用户名 |
| 密码 | `123456` | MySQL 密码 |
| 协议 | `tcp` | 连接协议，本地用 tcp |
| 主机 | `127.0.0.1` | MySQL 服务器地址 |
| 端口 | `3306` | MySQL 端口（默认 3306） |
| 数据库名 | `gorm_tutorial` | 要连接的数据库 |
| 参数 | `charset=...&parseTime=...&loc=...` | 连接参数 |

### DSN 参数大全

| 参数 | 必须 | 说明 | 推荐值 |
|------|------|------|--------|
| `charset` | 推荐 | 字符集 | `utf8mb4`（支持 emoji） |
| `parseTime` | **必须！** | 是否把 DATE/DATETIME 转换为 `time.Time` | `True`（否则时间字段是字符串） |
| `loc` | **必须！** | 时区设置 | `Local`（使用系统时区）或 `Asia%2FShanghai` |
| `timeout` | 可选 | 连接超时时间 | `10s` |
| `readTimeout` | 可选 | 读超时时间 | `30s` |
| `writeTimeout` | 可选 | 写超时时间 | `30s` |
| `tls` | 可选 | SSL/TLS 配置 | `true` 或 `skip-verify` |

> ⚠️ **`parseTime=True` 极其重要！** 如果不设置，MySQL 的 `DATE`/`DATETIME` 类型会被当作字符串返回，而不是 Go 的 `time.Time`。几乎所有 GORM 教程都会强调这一点。

> ⚠️ **`loc=Local` 也极其重要！** 不设置时，时区默认是 `UTC`，你存的"北京时间 8:00"会变成"UTC 8:00"，查询时再转回来就乱了。

### 几个常用的 DSN 示例

```go
// 最简 DSN（仅开发环境）
dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"

// 带超时的 DSN
dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s"

// 使用亚洲/上海时区
dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"

// 连接远程 MySQL
dsn := "user:pass@tcp(192.168.1.100:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
```

---

## 3.5 `gorm.Open()` 详解

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    // 配置项（下一章详解）
})
```

`gorm.Open()` 接收两个参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| 第一个 | `gorm.Dialector` | 数据库方言（决定如何生成 SQL） |
| 第二个 | `*gorm.Config` | GORM 全局配置（可选） |

第一个参数通过 `mysql.Open(dsn)` 创建，它是一个 `gorm.Dialector` 接口，负责：
1. 建立底层数据库连接
2. 根据数据库类型生成对应的 SQL 语法
3. 处理数据库特定的数据类型映射

```go
// MySQL
import "gorm.io/driver/mysql"
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// PostgreSQL
import "gorm.io/driver/postgres"
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// SQLite（测试神器，不需要安装数据库软件！）
import "gorm.io/driver/sqlite"
db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

// SQLite 内存数据库（测试时超方便，程序结束数据消失）
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

---

## 3.6 `sql.DB` 与 `gorm.DB` 的关系

这是初学者最容易混淆的概念。

```
┌─────────────────────────────────┐
│         *gorm.DB                │
│   (GORM 的高级封装)              │
│   - 链式查询                     │
│   - 模型映射                     │
│   - 关联关系                     │
│   - Hook/事务                    │
│   ┌─────────────────────┐       │
│   │    *sql.DB          │       │
│   │  (Go 标准库的连接池)  │       │
│   │  - 连接管理          │       │
│   │  - 连接池            │       │
│   └─────────────────────┘       │
└─────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│   MySQL 数据库    │
└─────────────────┘
```

- **`*sql.DB`**：Go 标准库 `database/sql` 提供的**连接池**，负责管理 TCP 连接。它不执行 ORM 功能。
- **`*gorm.DB`**：GORM 对 `*sql.DB` 的封装，提供了所有 ORM 功能（CRUD、关联、Hook 等）。

```go
// db 是 *gorm.DB（高级封装，处理 ORM 逻辑）
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// sqlDB 是 *sql.DB（底层连接池，处理 TCP 连接）
sqlDB, err := db.DB()

// 连接池配置是在 sqlDB 上设置的
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)

// ORM 操作是在 db 上执行的
db.Create(&user)
db.First(&user)
```

> **关键认知**：一个 `*gorm.DB` 实例代表一个数据库连接池，不是一条 TCP 连接。你可以全局创建一个 `db`，在程序各处复用。它是并发安全的。

---

## 3.7 连接池基础配置

连接池是什么？打个比方：

- 不用连接池：每次来客人就去买双新筷子，用完就扔 → 浪费
- 用连接池：提前准备好一筒筷子，用完洗洗放回去重复用 → 高效

```go
sqlDB, _ := db.DB()

// 1. 最大打开连接数（包括正在使用的 + 空闲的）
//    默认值：无限！（危险！高并发时可能打爆数据库）
sqlDB.SetMaxOpenConns(25)

// 2. 最大空闲连接数（保持多少空闲连接待命）
//    默认值：2
sqlDB.SetMaxIdleConns(10)

// 3. 连接最大存活时间（超过这个时间，连接被关闭并重建）
//    默认值：永久（危险！MySQL 默认 8 小时断连接）
sqlDB.SetConnMaxLifetime(5 * time.Minute)

// 4. 空闲连接最大存活时间（超过这个时间没被使用，关闭）
//    Go 1.15 新增
sqlDB.SetConnMaxIdleTime(1 * time.Minute)
```

**推荐的生产环境默认值**（后续第32章详解调优）：

```go
sqlDB.SetMaxOpenConns(25)                  // 一般应用 25-100 即可
sqlDB.SetMaxIdleConns(10)                  // 约为 MaxOpenConns 的 1/3
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 小于 MySQL wait_timeout
sqlDB.SetConnMaxIdleTime(1 * time.Minute)  // 及时释放空闲连接
```

---

## 常见错误

### 错误1：Access denied（用户名或密码错误）

```
Error 1045: Access denied for user 'root'@'localhost' (using password: YES)
```

**原因**：MySQL 用户名或密码不对。

**解决**：
1. 确认用户名和密码是否正确
2. 尝试用 MySQL 命令行连接验证：`mysql -u root -p`
3. 检查用户是否有对应数据库的权限

### 错误2：Unknown database（数据库不存在）

```
Error 1049: Unknown database 'gorm_tutorial'
```

**原因**：DSN 中指定的数据库不存在。

**解决**：
```sql
CREATE DATABASE gorm_tutorial CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 错误3：Cannot connect to MySQL（无法连接）

```
dial tcp 127.0.0.1:3306: connect: connection refused
```

**原因**：MySQL 服务未启动或端口不对。

**解决**：
- Windows：检查"服务"中 MySQL 是否在运行
- Mac/Linux：`sudo systemctl status mysql` 或 `brew services list`
- 检查端口：`mysql` 默认是 3306，`mariadb` 可能用 3307

### 错误4：`parseTime` 没设置导致时间变成字符串

```go
// ❌ 没有 parseTime=True
dsn := "root:123456@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4"

// GORM 查询后，CreatedAt 是字符串 "2024-01-15 10:30:00" 而不是 time.Time
// 各种时间操作全挂
```

**解决**：永远加上 `parseTime=True`！

### 错误5：忘记配置连接池

```go
// ❌ 没配置连接池
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// MaxOpenConns 默认无限！高并发时连接数爆炸

// ✅ 配置连接池
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
```

---

## 本章小结

- GORM 依赖两个包：`gorm.io/gorm`（核心）+ `gorm.io/driver/mysql`（驱动）
- DSN 是连接信息的字符串，**必须**包含 `parseTime=True` 和 `loc=Local`
- `gorm.Open()` 返回 `*gorm.DB`（ORM 实例），`db.DB()` 返回 `*sql.DB`（连接池）
- 连接池四个参数：MaxOpenConns、MaxIdleConns、ConnMaxLifetime、ConnMaxIdleTime
- 换数据库只需改驱动和 DSN，业务代码不用变

## 练习题

1. 创建一个新的 Go 项目，安装 GORM 和 MySQL 驱动，写出 `go.mod` 的内容。
2. 分别写出以下场景的 DSN：
   - 连接本地 MySQL，用户 `test`，密码 `pass123`，数据库 `mydb`
   - 连接远程 MySQL（地址 `10.0.0.5:3307`），用户 `admin`，密码 `secret`
// 连接远程 MySQL
dsn := "user:pass@tcp(192.168.1.100:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
3. 解释 `sql.DB` 和 `gorm.DB` 的区别和联系。
4. 为什么要设置 `SetMaxOpenConns`？不设置会有什么风险？
5. 写一个完整的程序：连接 MySQL → Ping 验证 → 打印连接池状态信息。
6. （思考题）如果你要用 SQLite 做单元测试，代码需要改哪些地方？为什么这样做很方便？
