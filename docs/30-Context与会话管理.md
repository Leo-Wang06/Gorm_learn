# 第30章：Context 与会话（Session）管理

## 本章目标
学完本章后，你将能够：
1. 使用 `WithContext` 传递 Context（超时、取消、traceID）
2. 使用 `Session` 创建独立的会话配置
3. 使用 `Debug()` 快速调试 SQL
4. 理解 Session 的线程安全性

## 前置知识
- 需要先学习：第04章（配置与Logger）
- 需要了解：Go 的 `context` 包基础

---

## 30.1 `db.WithContext(ctx)`

Context 是 Go 的标配——传递超时、取消信号和请求级数据：

```go
// 创建一个带超时的 Context
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

// 将 Context 传给 GORM
var users []User
err := db.WithContext(ctx).Find(&users).Error
// 如果查询超过 2 秒，SQL 会被取消

if err != nil {
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("查询超时！")
	}
}
```

### 从 HTTP 请求传递 Context

```go
func HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	// r.Context() 包含请求超时、取消、trace 信息
	var users []User
	err := db.WithContext(r.Context()).Find(&users).Error
	// 如果 HTTP 客户端断开连接，查询自动取消
}
```

---

## 30.2 `db.Session()`：会话配置

Session 是 GORM v2 的核心创新——在特定操作上覆盖全局配置：

```go
sessionDB := db.Session(&gorm.Session{
	NewDB:                true,   // 创建一个独立的 *gorm.DB
	SkipDefaultTransaction: true, // 跳过默认事务
	PrepareStmt:          true,   // 预编译缓存
	DryRun:               false,  // 不真正执行 SQL
	SkipHooks:            false,  // 跳过钩子
	AllowGlobalUpdate:    false,  // 允许全表更新
	FullSaveAssociations: false,  // 全量保存关联
	QueryFields:          false,  // 查询时使用字段名
	Logger:               customLogger, // 覆盖 Logger
	Context:              ctx,    // 设置 Context
})
```

### 常用场景

```go
// 场景1：临时打印 SQL（调试用）
db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)}).
	Where("age > ?", 18).Find(&users)

// 场景2：预览 SQL 但不执行
db.Session(&gorm.Session{DryRun: true}).
	Create(&User{Name: "测试"})

// 场景3：全表更新（慎用）
db.Session(&gorm.Session{AllowGlobalUpdate: true}).
	Model(&User{}).Update("status", "active")
```

---

## 30.3 `db.Debug()`：快速调试

```go
// 单次调用打印 SQL（超级好用！）
db.Debug().Where("age > ?", 18).Find(&users)
// 日志会输出：
// [1.234ms] [rows:3] SELECT * FROM `users` WHERE age > 18

// Debug 本质上是 Session 的快捷方式
db.Debug() == db.Session(&gorm.Session{Logger: ...})
```

---

## 30.4 Session 的线程安全性

```go
// ✅ *gorm.DB 是并发安全的，可以在多个 goroutine 中共用
var db *gorm.DB

func handler(w http.ResponseWriter, r *http.Request) {
	db.WithContext(r.Context()).Find(&users)  // 安全
}

// ⚠️ 但 Session 修改的是副本
func handler(w http.ResponseWriter, r *http.Request) {
	// Session 返回新的 *gorm.DB，不影响原始的 db
	sessionDB := db.Session(&gorm.Session{})
	sessionDB.Where(...)  // 不影响其他 goroutine 中的 db
}
```

---

## 本章小结

- `db.WithContext(ctx)` 传递超时、取消信号
- `db.Session(config)` 创建独立配置的会话
- `db.Debug()` 是调试 SQL 的快速手段
- `*gorm.DB` 并发安全，Session 操作副本
- HTTP 请求中始终使用 `r.Context()`

## 练习题

1. 创建一个 500ms 超时的 Context，执行一个必然超时的查询，处理超时错误。
2. 使用 Session 临时开启 DryRun，预览一条 Update 的 SQL。
3. 使用 Debug() 打印一条查询的 SQL 和耗时。
4. 在 HTTP Handler 中使用 `r.Context()` 传递 Context 给 GORM。
5. （思考题）为什么 Session 要创建 `*gorm.DB` 的副本而不是直接修改原始的 db？这和并发安全有什么关系？
