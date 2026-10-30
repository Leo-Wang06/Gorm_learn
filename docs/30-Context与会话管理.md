# 第30章：Context 与会话（Session）管理

## 本章目标
学完本章后，你将能够：
1. 使用 `WithContext` 传递 Context（超时、取消、traceID）
2. 使用 `Session` 创建独立的会话配置
3. 使用 `Debug()` 快速调试 SQL
4. 理解 Session 的线程安全性

## 前置知识
- 需要先学习：第04章（配置与Logger）
- 需要了解：Go 的 `context` 包基础、HTTP Handler 和 GORM 链式查询
- `WithContext` 只能传递取消/截止时间等信号；数据库驱动及服务器是否及时响应取消，还取决于其实现。

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
	// r.Context() 会在请求结束/客户端断开时取消；trace 等值由上游约定注入。
	var users []User
	if err := db.WithContext(r.Context()).Find(&users).Error; err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request canceled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	// 写回响应...
}
```

---

## 30.2 `db.Session()`：会话配置

Session 用于在特定操作链上覆盖配置或创建新的查询状态；它不会创建新的数据库连接池：

```go
sessionDB := db.Session(&gorm.Session{
	NewDB:                 true,          // 新建不带既有条件的查询状态；保留底层连接池
	SkipDefaultTransaction: true,         // 写操作跳过 GORM 默认事务（需理解一致性风险）
	PrepareStmt:           true,          // 启用预编译语句缓存
	DryRun:                false,         // true 才只构造 SQL、不执行
	SkipHooks:             false,         // true 才跳过钩子
	AllowGlobalUpdate:     false,         // true 才允许无条件批量更新/删除
	FullSaveAssociations:  false,         // true 才完整保存关联字段
	QueryFields:           false,         // true 才按字段名显式选择列
	Logger:                customLogger,  // 覆盖当前会话 Logger
	Context:               ctx,           // 设置当前会话的 Context
})
```

### 常用场景

```go
// 场景1：临时打印 SQL（调试用）
db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)}).
	Where("age > ?", 18).Find(&users)

// 场景2：预览 SQL 但不执行；使用 Statement 查看生成的 SQL 和绑定变量
stmt := db.Session(&gorm.Session{DryRun: true}).
	Create(&User{Name: "测试"}).Statement
fmt.Println(stmt.SQL.String(), stmt.Vars)

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

// Debug 本质上是临时提高当前链的日志级别；不要用指针相等判断两个 *gorm.DB。
// 可理解为：db.Debug() 等价于为这条查询链使用 Info 级别 Logger。
```

---

## 30.4 Session 的线程安全性

```go
// 已初始化的根 *gorm.DB 可由多个 goroutine 共用；每个请求使用自己的目标变量和 Context。
func handler(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := db.WithContext(r.Context()).Find(&users).Error; err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
}

// Session 产生新的会话配置/查询状态；链式方法返回新 DB 状态，避免并发请求共享可变的查询链。
sessionDB := db.Session(&gorm.Session{DryRun: true})
_ = sessionDB
```

---

## 本章小结

- `db.WithContext(ctx)` 传递超时、取消信号
- `db.Session(config)` 为查询链设置独立配置/状态，但不创建新的连接池
- `db.Debug()` 是临时提高日志级别的调试手段
- 已初始化的根 `*gorm.DB` 可并发复用；不要跨请求共享并修改同一条查询链
- HTTP 请求中始终使用 `r.Context()`

## 练习题

1. 创建一个 500ms 超时的 Context，执行一个必然超时的查询，处理超时错误。
2. 使用 Session 临时开启 DryRun，预览一条 Update 的 SQL。
3. 使用 Debug() 打印一条查询的 SQL 和耗时。
4. 在 HTTP Handler 中使用 `r.Context()` 传递 Context 给 GORM。
5. （思考题）为什么 Session 要创建 `*gorm.DB` 的副本而不是直接修改原始的 db？这和并发安全有什么关系？
