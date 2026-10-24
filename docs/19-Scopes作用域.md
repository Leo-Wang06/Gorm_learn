# 第19章：Scopes 作用域 —— 封装可复用查询

## 本章目标
学完本章后，你将能够：
1. 理解 Scope 的概念和签名
2. 编写可复用的 Scope 函数（分页、过滤、排序）
3. 组合多个 Scope 构建复杂查询
4. 使用闭包实现参数化 Scope
5. 根据条件动态应用 Scope
6. 建立项目级的 Scope 函数库

## 前置知识
- 需要先学习：第11-18章（查询能力）
- 需要了解：Go 的闭包和高阶函数

---

## 19.1 什么是 Scope

Scope 是一个**接受 `*gorm.DB` 并返回 `*gorm.DB` 的函数**：

```go
// Scope 的签名
type Scope func(db *gorm.DB) *gorm.DB
```

它让你把常用的查询片段封装起来，随处复用：

```go
// 定义一个简单的 Scope
func ActiveUsers(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "active")
}

// 使用
db.Scopes(ActiveUsers).Find(&users)
// SQL: SELECT * FROM `users` WHERE status='active'
```

### 为什么需要 Scope

没有 Scope 的代码：
```go
// 三个地方都要查活跃用户，每个地方都写一遍 Where
db.Where("status = ?", "active").Find(&users)     // 第1处
db.Where("status = ?", "active").First(&user, id) // 第2处
db.Where("status = ?", "active").Count(&count)    // 第3处
```

有 Scope 的代码：
```go
// 一次定义，到处复用
db.Scopes(ActiveUsers).Find(&users)
db.Scopes(ActiveUsers).First(&user, id)
db.Scopes(ActiveUsers).Count(&count)

// 如果过滤条件变了，只改 Scope 定义一处即可
```

---

## 19.2 基础 Scope 示例

### 分页 Scope

```go
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// 使用
db.Scopes(Paginate(2, 10)).Find(&users)
// SQL: SELECT * FROM `users` LIMIT 10 OFFSET 10
```

### 状态过滤 Scope

```go
func WithStatus(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db  // 空状态时不加条件
	}
}

// 使用
db.Scopes(WithStatus("active")).Find(&users)
```

### 时间范围 Scope

```go
func CreatedBetween(start, end time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at BETWEEN ? AND ?", start, end)
	}
}

// 使用
weekAgo := time.Now().Add(-7 * 24 * time.Hour)
db.Scopes(CreatedBetween(weekAgo, time.Now())).Find(&users)
```

---

## 19.3 Scope 的组合

多个 Scope 可以链式组合：

```go
// 查询最近一周创建的活跃成年用户，分页展示
db.Scopes(
	ActiveUsers,                         // status = 'active'
	AdultUsers,                          // age >= 18
	CreatedBetween(weekAgo, time.Now()), // 最近一周
	Paginate(1, 20),                     // 分页
).Order("created_at desc").Find(&users)

// SQL:
// SELECT * FROM `users`
// WHERE status='active'
//   AND age>=18
//   AND created_at BETWEEN '2024-01-08' AND '2024-01-15'
// ORDER BY created_at DESC
// LIMIT 20 OFFSET 0
```

---

## 19.4 条件 Scope（按需生效）

通过闭包参数控制 Scope 是否生效：

```go
// 只有当 keyword 非空时才添加 LIKE 条件
func WithKeyword(keyword string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if keyword != "" {
			return db.Where("name LIKE ?", "%"+keyword+"%")
		}
		return db  // 空关键词，不添加条件
	}
}

// 只有当 minAge > 0 时才添加年龄条件
func WithMinAge(minAge int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if minAge > 0 {
			return db.Where("age >= ?", minAge)
		}
		return db
	}
}
```

### 实际场景：复杂搜索 API

```go
type UserSearchParams struct {
	Keyword  string
	Status   string
	MinAge   int
	MaxAge   int
	City     string
	Page     int
	PageSize int
}

func SearchUsers(db *gorm.DB, params UserSearchParams) ([]User, int64, error) {
	var users []User
	var total int64

	query := db.Model(&User{}).Scopes(
		WithKeyword(params.Keyword),
		WithStatus(params.Status),
		WithMinAge(params.MinAge),
		WithMaxAge(params.MaxAge),
		WithCity(params.City),
	)

	// 先统计总数
	query.Count(&total)

	// 再查分页数据
	err := query.Scopes(Paginate(params.Page, params.PageSize)).
		Order("created_at desc").
		Find(&users).Error

	return users, total, err
}
```

---

## 19.5 常用 Scope 模板库

```go
// ============ 过滤类 ============

func WhereLike(column, keyword string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if keyword != "" {
			return db.Where(column+" LIKE ?", "%"+keyword+"%")
		}
		return db
	}
}

func WhereIn(column string, values []interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(values) > 0 {
			return db.Where(column+" IN ?", values)
		}
		return db
	}
}

func WhereDateBetween(column string, start, end *time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if start != nil {
			db = db.Where(column+" >= ?", *start)
		}
		if end != nil {
			db = db.Where(column+" <= ?", *end)
		}
		return db
	}
}

// ============ 排序类 ============

func OrderBy(field string, desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order(field + " desc")
		}
		return db.Order(field + " asc")
	}
}

// ============ 分页类 ============

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 { page = 1 }
		if pageSize <= 0 { pageSize = 10 }
		return db.Offset((page - 1) * pageSize).Limit(pageSize)
	}
}

// ============ 预加载类 ============

func WithPreload(associations ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, assoc := range associations {
			db = db.Preload(assoc)
		}
		return db
	}
}
```

---

## 19.6 Scope 中的 DB 操作（小心！）

Scope 不仅能加查询条件，还能做更多事：

```go
// Scope 中执行额外操作（不推荐，但有时有用）
func AuditLog(db *gorm.DB) *gorm.DB {
	// 记录查询日志
	log.Printf("执行查询: %s", db.Statement.SQL.String())
	return db
}

// 更实用的：自动添加租户隔离
func TenantScope(tenantID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}
// 所有查询自动带上租户 ID，防止数据泄漏
```

---

## 常见错误

### 错误1：Scope 改变了不该改变的东西

```go
// ❌ Scope 中使用了不安全的修改
func BadScope(db *gorm.DB) *gorm.DB {
	db.Statement.Table = "hacked"  // 改了表名！
	return db
}

// ✅ Scope 应该只用 Where/Order/Limit/Preload 等方法
func GoodScope(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "active")
}
```

### 错误2：忘记 return

```go
// ❌ Scope 函数没有返回
func ActiveUsers(db *gorm.DB) *gorm.DB {
	db.Where("status = ?", "active")  // 没有 return！
}
// Scope 不生效

// ✅ 必须 return
func ActiveUsers(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "active")
}
```

### 错误3：Scope 过于复杂

```go
// ❌ 一个 Scope 做了太多事
func EverythingScope(db *gorm.DB) *gorm.DB {
	return db.Where("...").Order("...").Limit(10).Preload("...").Joins("...")
}

// ✅ 拆分成多个单一职责的 Scope
db.Scopes(StatusFilter, DateRange, Paginate).Find(&users)
```

---

## 本章小结

- Scope 是 `func(*gorm.DB) *gorm.DB` 类型的函数
- 用闭包实现参数化的 Scope
- `db.Scopes(s1, s2, s3)` 组合多个 Scope
- Scope 可复用、可测试、可管理
- 条件 Scope：参数为空时不添加条件（`return db`）
- 最佳实践：建立项目级 `scopes/` 包

## 练习题

1. 定义一个 `WithStatus` Scope，接受状态参数，当参数非空时添加 WHERE 条件。
2. 定义 `WithDateRange` Scope，接受起始和结束时间，用于过滤 `created_at`。
3. 组合使用三个 Scope 实现：查询上周创建的、活跃的、VIP 用户，并分页展示。
4. 将第15章的分页逻辑改写为 Scope 版本。
5. 写一个 `SearchUsers` 函数，接收多个可选搜索条件，用 Scope 动态构建查询。
6. （思考题）Scope 和直接链式调用 `db.Where().Where()` 有什么区别？什么情况下 Scope 更好？
