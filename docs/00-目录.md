# GORM 从入门到精通：完整教程目录

> **教程定位**：本教程面向有 MySQL 基础、Go 语言零基础的读者，从 Go 语言基础讲起，逐步深入到 GORM 模型定义、增删改查、关联关系、事务钩子、性能优化、代码生成器等高级主题。每章配有可运行的 Go 代码、常见错误分析、练习题，帮助你在实践中掌握 GORM v2。

---

## 阅读指南

本教程共 42 章 + 3 个附录，按 10 个阶段组织。每个阶段聚焦一个核心主题，循序渐进：

| 阶段 | 章节范围 | 核心目标 | 建议学习时间 |
|------|----------|----------|-------------|
| Phase 1 | 第01-04章 | Go语言基础速成 + GORM环境搭建 | 1-2 天 |
| Phase 2 | 第05-08章 | 模型定义与表结构映射 | 2-3 天 |
| Phase 3 | 第09-14章 | 增删改查 CRUD 全掌握 | 3-5 天 |
| Phase 4 | 第15-20章 | 高级查询能力 | 3-4 天 |
| Phase 5 | 第21-26章 | 关联关系（HasOne/HasMany/BelongsTo/Many2Many） | 4-6 天 |
| Phase 6 | 第27-31章 | 事务、钩子与高级特性 | 3-4 天 |
| Phase 7 | 第32-35章 | 性能优化、安全与最佳实践 | 2-3 天 |
| Phase 8 | 第36-38章 | Gen代码生成器、插件与扩展 | 2-3 天 |
| Phase 9 | 第39-41章 | 实战项目 | 3-5 天 |
| Phase 10 | 第42章 | 总复习 + 面试宝典 | 1-2 天 |

### 学习建议

1. **按顺序学习**：GORM 知识有很强的依赖关系，前几章的模型定义和基础 CRUD 不学好，后面的关联和事务就寸步难行
2. **必须手敲代码**：每个 Go 代码示例都要自己在电脑上敲一遍，看一遍和敲一遍的效果天差地别
3. **边学边看 SQL 日志**：学习 GORM 时务必开启 SQL 日志，观察 GORM 实际生成的 SQL 语句，这对理解 ORM 行为至关重要
4. **练习必须做**：每章末尾的练习题是巩固知识的关键
5. **理解生成的 SQL 比记住 GORM API 更重要**：GORM 本质是 SQL 的生成器，理解每条操作在数据库层面做了什么，才能写出高效的代码
6. **遇到错误先自己排查**：不要一看到报错就急着问人，试着读懂错误信息，用搜索引擎搜索，排查的过程本身就是最好的学习
7. **搭建一个自己的项目**：学完 Phase 5 后，尝试自己设计一个小型业务系统，把学到的知识串联起来

---

## Phase 1：Go 语言速成与 GORM 环境搭建

### 第 01 章：Go 语言极速入门（为 GORM 学习准备）
- Go 安装与环境配置（GOPATH、Go Modules）
- 变量、常量、基本数据类型
- 函数定义与多返回值
- 结构体（struct）基础 —— 这是 GORM 模型的核心
- 指针基础（`*` 和 `&`）—— GORM 大量使用指针
- 接口（interface）最简理解
- 错误处理（error）—— Go 的 error 不是异常
- 包（package）与导入（import）
- `go mod init` 初始化项目
- 一个完整的 Go 程序结构

### 第 02 章：什么是 ORM 与 GORM 概述
- 什么是 ORM（Object-Relational Mapping），一句人话解释
- 为什么需要 ORM：手写 SQL vs ORM 优劣对比表
- Go 生态中的 ORM 选择（GORM / Ent / sqlx / sqlc 对比）
- GORM 的发展历史与版本（v1 vs v2 差异）
- GORM 的核心设计哲学（约定优于配置、链式调用）
- GORM 支持哪些数据库（MySQL/PostgreSQL/SQLite/SQL Server）
- GORM 的 Stars、社区、文档情况
- 第一个 GORM 心理模型：`struct ↔ 表`、`字段 ↔ 列`、`实例 ↔ 行`

### 第 03 章：安装 GORM 与连接数据库
- `go get -u gorm.io/gorm` 安装 GORM
- `go get -u gorm.io/driver/mysql` 安装 MySQL 驱动
- 其他驱动概览（postgres/sqlite/sqlserver）
- `gorm.Open()` 连接数据库完整代码
- DSN（Data Source Name）详解
- DSN 参数详解（charset/parseTime/loc/timeout/readTimeout/writeTimeout）
- 连接池配置入门（SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime）
- `sql.DB` 与 `gorm.DB` 的关系
- 第一个完整可运行程序：连接 → Ping → 打印成功
- 常见连接错误排查（Access denied/Unknown database/network timeout）

### 第 04 章：GORM 配置与 Logger 日志
- GORM 配置对象 `gorm.Config{}` 详解
- `SkipDefaultTransaction`：关闭默认事务提升性能
- `NamingStrategy`：命名策略（表名/列名的单复数、蛇形转换规则）
- `DisableForeignKeyConstraintWhenMigrating`：迁移时是否创建外键
- `PrepareStmt`：预编译缓存
- `DryRun`：干运行模式（只生成 SQL 不执行）
- Logger 日志：`logger.Interface`
- 四种日志级别（Silent/Error/Warn/Info）
- 慢查询阈值设置（SlowThreshold）
- 彩色日志 vs 纯文本日志
- 自定义 Logger 示例

---

## Phase 2：模型定义与表结构映射

### 第 05 章：模型定义基础 —— 结构体到表的映射
- GORM 模型 = Go 结构体 + 标签（tag）
- `gorm.Model` 内嵌结构体（ID/CreatedAt/UpdatedAt/DeletedAt）
- 约定优于配置：表名默认规则（复数蛇形）、列名默认规则
- 主键约定（ID 字段默认为主键）
- 结构体标签（tag）语法：`` `gorm:"column:xxx;type:xxx;..."` ``
- 可导出字段 vs 不可导出字段
- 用 `AutoMigrate()` 自动创建表
- `Migrator().HasTable()` / `Migrator().DropTable()` 检查表是否存在

### 第 06 章：模型标签全解（上）—— 字段定义
- `column`：自定义列名
- `type`：列数据类型（varchar/int/text/decimal 等）
- `size`：列大小
- `precision` / `scale`：小数精度
- `primaryKey` / `autoIncrement`：主键与自增
- `default`：默认值（零值字段的 default 陷阱）
- `not null`：非空约束
- `unique` / `uniqueIndex`：唯一约束与唯一索引
- `index`：普通索引
- `check`：检查约束
- `comment`：列注释
- `<-` / `->` / `-` ：字段读写权限控制

### 第 07 章：模型标签全解（下）—— 特殊类型与嵌入
- `serializer`：序列化器（json/gob/unixtime）
- `embedded` / `embeddedPrefix`：嵌入结构体
- 时间类型处理：`time.Time` / `*time.Time` / `sql.NullTime`
- `autoCreateTime` / `autoUpdateTime`：自动时间戳
- 枚举类型的 GORM 实现
- JSON 字段的存储与查询
- 自定义 Scanner/Valuer 接口实现
- 指针字段 vs 值字段的行为差异（零值问题）

### 第 08 章：表名约定与迁移详解
- 表名生成规则深度解析（默认复数蛇形 vs 自定义）
- 三种自定义表名方式（TableName() 方法 / db.Table() / NamingStrategy）
- 临时指定表名：`db.Table("my_table").Find(&users)`
- `AutoMigrate()` 能做什么、不能做什么
- `Migrator()` 接口高级操作（HasTable/DropTable/RenameTable/AddColumn/DropColumn 等）
- 迁移在生产环境的最佳实践

---

## Phase 3：增删改查 CRUD 全掌握

### 第 09 章：创建数据 Create
- `db.Create(&user)` 基础语法
- 创建单条记录 / 批量创建 / 分批创建
- 创建后获取自增 ID（自动回填）
- 指定字段创建 / 忽略字段创建 / Map 创建
- UPSERT（存在则更新，不存在则插入）：`clause.OnConflict`
- `Save()` 方法：根据主键判断 Insert 还是 Update
- `FirstOrCreate()` 先查后插
- `Attrs()` / `Assign()` 的区别
- 常见错误：零值不插入、忘记传指针

### 第 10 章：查询单个记录 —— First/Take/Last/Find
- `First()` / `Take()` / `Last()` / `Find()` 四兄弟详解
- 主键查询的多种写法
- 结构体条件查询陷阱（零值字段不会被用作条件！）
- `ErrRecordNotFound` 错误的正确判断
- 查询不存在的记录时的行为差异
- `FirstOrInit` / `FirstOrCreate` 的区别与应用
- 常见错误：零值条件陷阱、忘记传指针

### 第 11 章：条件查询 —— WHERE 子句详解
- 字符串条件（占位符 `?` 防注入）
- 结构体条件 vs Map 条件（零值陷阱详解）
- `Not()` / `Or()` 条件组合
- BETWEEN / IN / LIKE / IS NULL 运算符
- 链式 Where 的 AND 组合
- 命名参数：`@name` 语法
- 常见错误：结构体零值不参与查询、And/Or 优先级

### 第 12 章：更新数据 Update
- `Save(&user)`：全字段保存（基于主键）
- `Update()` / `Updates()` 更新操作
- `UpdateColumn()` / `UpdateColumns()` 跳过钩子的更新
- 结构体 Updates 的零值问题 ⚠️
- Map Updates（可更新零值）
- `Select()` / `Omit()` 控制更新字段
- 表达式更新：`gorm.Expr("age + ?", 1)`
- 批量更新的安全措施（AllowGlobalUpdate）
- Save vs Updates vs UpdateColumns 选择决策树

### 第 13 章：删除数据 Delete
- 按主键删除 / 按条件批量删除
- 软删除（Soft Delete）核心概念
- `gorm.DeletedAt` 字段原理
- 软删除的查询行为（自动过滤已删除记录）
- `Unscoped()` 查询包含已删除记录 / 硬删除
- 自定义软删除字段
- 软删除 vs 硬删除的适用场景分析

### 第 14 章：原生 SQL 与高级 CRUD 技巧
- `Raw()` 原生查询 / `Exec()` 原生执行
- `Scan()` 扫描结果到任意结构体
- `Pluck()` 提取单列到切片
- `Count()` / `Distinct()` 计数与去重
- `Clauses()` 子句使用（ON CONFLICT / RETURNING 等）
- `gorm.Expr` 表达式的高级用法
- DryRun 模式确认 SQL 再执行

---

## Phase 4：高级查询能力

### 第 15 章：排序、分页与限定查询
- `Order()` 单列/多列排序
- `Limit()` / `Offset()` 分页查询
- 分页封装函数（通用 Paginate 模式）
- 大偏移量分页的性能问题与优化方案
- `Scopes()` 作用域初体验
- 常见错误：Offset 过大、忘记 Order

### 第 16 章：选择特定字段与聚合查询
- `Select()` 选择特定列
- `Select()` 中使用聚合函数与表达式
- `Group()` / `Having()` 分组过滤
- `Count()` / `Sum()` / `Avg()` / `Max()` / `Min()`
- 聚合 + 分组查询完整示例

### 第 17 章：连接查询 —— Joins
- `Joins()` 字符串 Join / 关联名 Join
- `InnerJoins()` 内连接
- 多表 Join / 自连接查询
- Join 后 Where 条件与 Scan
- Joins 与 Preload 的区别

### 第 18 章：子查询与高级查询技巧
- Where 中的子查询
- From 中的子查询（派生表）
- Select 中的子查询
- 关联子查询
- `Scan()` 扫描到 map / 自定义结构体切片
- `Row()` / `Rows()` 处理原始 `*sql.Rows`
- 命名参数的使用

### 第 19 章：Scopes 作用域 —— 封装可复用查询
- 什么是 Scope：返回 `func(db *gorm.DB) *gorm.DB` 的函数
- 基础 Scope：分页、过滤、排序
- 带参数的 Scope（闭包）
- Scope 的链式组合
- 常用 Scope 模板库（Paginate/WhereLike/DateBetween/StatusIn）
- Scope vs 直接写 Where 的选择场景

### 第 20 章：查询总结与对比表
- 所有查询方法速查表
- 方法选择决策树
- 查询性能基准线
- 常见查询模式汇总
- GORM vs 原生 SQL 代码量对比

---

## Phase 5：关联关系

### 第 21 章：关联关系概述与 BelongsTo
- GORM 四种关联类型全景图
- 外键（ForeignKey）与引用（References）配置
- `BelongsTo`：属于关系（User 属于 Company）
- 默认外键规则（`类型名 + ID`）
- 创建/更新带 BelongsTo 的记录
- 常见错误：忘记设置外键、外键类型不匹配

### 第 22 章：HasOne —— 一对一关系
- `HasOne`：拥有一对一关系（User 拥有一张 CreditCard）
- HasOne 标签配置
- 创建/查询 HasOne 关联
- HasOne vs BelongsTo 的区别与选择
- 完整一对一实战示例

### 第 23 章：HasMany —— 一对多关系
- `HasMany`：拥有一对多关系（User 拥有多个 Order）
- HasMany 标签配置与外键设置
- 创建/查询 HasMany 关联
- 一对多关系的增删改操作
- 完整一对多实战示例（用户与订单）

### 第 24 章：Many2Many —— 多对多关系
- `Many2Many`：多对多关系（Student 选修多门 Course）
- 中间表（Join Table）的概念与命名
- Many2Many 标签配置
- 自定义中间表 / 中间表携带额外字段
- 创建/查询多对多关联
- 完整多对多实战示例（学生选课系统）

### 第 25 章：Preload 预加载 —— 解决 N+1 问题
- 什么是 N+1 查询问题（生动图解）
- `Preload()` 预加载关联数据
- 条件预加载 / 嵌套预加载 / 自定义预加载
- `Joins` 预加载：LEFT JOIN 单次查询
- Preload 与 Joins 的区别与选择
- 常见错误：Preload 不生效、循环 Preload

### 第 26 章：Association 模式 —— 关联的增删改查
- `Association()` 关联模式
- `Append()` / `Delete()` / `Replace()` / `Clear()` / `Count()` / `Find()`
- Association 与直接操作的区别
- 批量操作关联的最佳实践
- 常见错误：Association 和 Preload 混淆

---

## Phase 6：事务、钩子与高级特性

### 第 27 章：事务（Transaction）详解
- 什么是事务，GORM 中事务的意义
- `db.Transaction()` 自动事务（推荐）
- 手动事务：`Begin()` / `Commit()` / `Rollback()`
- 嵌套事务与 SavePoint
- 事务隔离级别设置
- 悲观锁与乐观锁
- `SkipDefaultTransaction` 全局配置
- 实战：转账操作（A扣钱、B加钱，任一失败则回滚）
- 常见错误：忘记回滚、事务中 panic

### 第 28 章：钩子（Hooks）详解
- GORM 生命周期与钩子执行顺序图解
- 创建/更新/删除/查询/保存钩子全览
- 钩子函数签名与返回值
- 钩子中修改字段值
- 钩子中使用 DB（小心死循环）
- 跳过钩子的方法
- 实战：自动生成 UUID、密码自动加密、自动更新时间戳
- 常见错误：钩子循环调用、忘记返回 error

### 第 29 章：自定义数据类型
- Go 类型 ↔ MySQL 类型的完整映射表
- `Scanner` / `Valuer` 接口详解
- 自定义 JSON 类型 / 枚举类型 / 加密类型
- `serializer` 标签（json/unixtime/gob）
- 自定义 Serializer 接口
- 实战：实现一个自动加密解密的字段类型

### 第 30 章：Context 与会话管理
- `db.WithContext(ctx)`：传递 context
- Context 超时控制 / 取消传播
- `db.Session()` 创建新的会话配置
- Session 可配置项全解
- `db.Debug()` 打印 SQL
- Session 的线程安全性
- Context 传递的最佳实践

### 第 31 章：错误处理与调试
- GORM 错误类型总览（10+ 种内置错误）
- `ErrRecordNotFound` / `ErrDuplicatedKey` / `ErrInvalidTransaction` 等
- `errors.Is()` 判断错误类型
- `db.Error` / `result.Error` 的区别
- 调试三板斧：Debug() / DryRun / Logger
- 常见错误：忽略 Error、错误比较用 == 而不是 errors.Is

---

## Phase 7：性能优化、安全与最佳实践

### 第 32 章：连接池与数据库配置优化
- `sql.DB` 连接池原理
- 四个核心参数详解（MaxOpenConns/MaxIdleConns/ConnMaxLifetime/ConnMaxIdleTime）
- 如何根据业务确定连接池参数
- 连接池状态监控（Stats() 方法）
- 连接泄漏排查
- 实战：配置一个生产级连接池

### 第 33 章：批量操作与性能优化
- `CreateInBatches()` 分批插入
- `FindInBatches()` 分批查询
- 批量更新 vs 逐条更新的性能对比
- `PrepareStmt` 预编译缓存
- `SkipDefaultTransaction` 性能提升
- `IndexHint()` 索引提示
- N+1 问题的识别与解决
- 实战：批量导入 10 万条数据的优化过程

### 第 34 章：SQL 注入防护与安全实践
- SQL 注入原理回顾
- GORM 如何防止 SQL 注入（参数化查询 `?`）
- ⚠️ 不安全代码示例（字符串拼接 SQL）
- ✅ 安全代码示例（始终使用占位符）
- Raw/Exec 中的安全问题
- 密码存储（bcrypt，绝不明文存储）
- 敏感字段脱敏
- 数据库账号最小权限原则

### 第 35 章：测试与 Mock
- 为什么 GORM 代码需要测试
- 使用 SQLite 内存数据库进行单元测试
- `gorm.Open(sqlite.Open(":memory:"))` 测试环境
- 测试辅助函数：setupTestDB() / teardownTestDB()
- 每个测试用例独立数据库
- `go-sqlmock` 介绍
- GORM + sqlmock 实战示例
- CRUD 操作的测试用例编写

---

## Phase 8：Gen 代码生成器、插件与扩展

### 第 36 章：GORM Gen —— 代码生成器入门
- 什么是 GORM Gen，为什么要用它
- Gen vs 手写代码的对比
- 安装 Gen 工具
- `gen.NewGenerator(cfg)` 配置生成器
- 生成 Model：从数据库表反向生成结构体
- 生成 Query：类型安全的 DAO
- `ApplyBasic()` / `ApplyInterface()` 配置
- 生成的代码结构解析

### 第 37 章：GORM Gen —— 高级用法
- Gen 生成的查询方法详解
- 动态条件查询（Eq/Neq/Gt/Gte/Lt/Lte/In/Like 等）
- Gen 的关联操作 / 事务支持
- Gen + 自定义 SQL
- Gen 的单元测试
- 实战：用 Gen 重写一个 CRUD 服务

### 第 38 章：插件系统与高级扩展
- GORM 插件接口 `gorm.Plugin`
- 内置插件：DBResolver（读写分离）
- 读写分离配置实战（主库写入，从库读取）
- 多数据库源配置
- 分库分表插件介绍（Sharding）
- Prometheus 监控指标暴露
- 自定义 Plugin 开发（如：自动填充创建人/修改人）
- Callback 系统简介

---

## Phase 9：实战项目

### 第 39 章：实战项目一 —— RESTful API 用户管理系统
- 项目初始化与目录结构
- 模型定义（User + Profile 一对一）
- 路由设计（Gin + GORM）
- 用户注册（Create + 密码加密钩子 + 参数校验）
- 用户登录（查询 + bcrypt 验证）
- 用户列表分页查询（Scopes + 条件筛选）
- 用户详情查询（Preload Profile）
- 用户更新 / 软删除
- 统一错误处理中间件
- 项目完整可运行代码

### 第 40 章：实战项目二 —— 博客系统
- 需求分析与 ER 图设计
- 模型设计（User + Article + Comment + Tag 多对多）
- 文章 CRUD（创建/编辑/删除/列表）
- 评论系统（HasMany 一对多）
- 标签系统（Many2Many 多对多）
- 文章列表 + 分页 + 标签筛选
- 文章详情 + 评论预加载 + 作者信息
- 事务实现：发表文章 + 关联标签
- 软删除文章与评论

### 第 41 章：实战项目三 —— 电商订单系统
- 需求分析与库表设计（用户/商品/订单/订单明细）
- 模型设计（User + Product + Order + OrderItem）
- 核心业务：下单流程（事务保证库存扣减+订单创建）
- 并发控制：悲观锁防止超卖
- 订单查询（多条件筛选+分页+排序）
- 订单统计（聚合查询+GroupBy+时间范围）
- 连接池配置与压测
- 完整的项目代码与测试

---

## Phase 10：总结与面试

### 第 42 章：GORM 面试宝典与总复习
- GORM 知识点体系化回顾
- 高频面试题 30 题：
  - GORM 零值陷阱是什么？如何避免？
  - N+1 问题是什么？如何用 Preload 解决？
  - Save vs Update vs Updates 的区别？
  - 软删除原理？如何查询已删除记录？
  - 事务如何正确使用？
  - GORM 钩子执行顺序？
  - 如何优化 GORM 的批量插入性能？
  - First vs Take vs Last 的区别？
  - 结构体条件查询的零值问题？
- 面试场景题（博客系统设计 / 转账事务 / 千万级分页优化）
- GORM vs Ent vs sqlx vs sqlc 选型指南
- 进阶学习路线图

---

## 附录

### 附录 A：GORM 常用标签速查表
- 所有 gorm 标签分类汇总
- 所有数据类型映射表（Go ↔ MySQL）
- 关联标签速查表
- 常用链式方法速查表

### 附录 B：GORM 常见错误排查指南
- 连接错误（MySQL 5.7 vs 8.0 差异、SSL 问题）
- 迁移错误（类型不匹配、列已存在）
- 查询错误（零值/空值/类型不匹配）
- 关联错误（外键类型不匹配、Preload 不生效）
- 事务错误（嵌套事务、事务过期）
- 性能问题排查（慢查询、连接泄漏）

### 附录 C：项目环境搭建 FAQ
- Go 安装 FAQ（Windows/Mac/Linux）
- GORM 安装 FAQ（网络问题、版本选择）
- MySQL 连接 FAQ（DSN 格式、权限配置）
- 开发工具推荐（VS Code + Go 插件 + MySQL 插件）

---

> **最后提醒**：GORM 技能不是"看"出来的，是"敲"出来的。请务必跟着教程敲每一行代码，务必开启 SQL 日志观察生成的 SQL，遇到错误不要慌，仔细阅读错误信息，它通常会告诉你问题出在哪里。祝你学有所成！
