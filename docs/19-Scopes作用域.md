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
	// 先规范输入并限制 pageSize；实际 API 还应对超大 page 做范围/溢出校验。
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
	}
}

// 使用：分页查询配合稳定排序
db.Order("id ASC").Scopes(Paginate(2, 10)).Find(&users)
// SQL: SELECT * FROM `users` ORDER BY id ASC LIMIT 10 OFFSET 10
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

多个 Scope 可以链式组合。此处的 `ActiveUsers`、`Paginate`、`CreatedBetween` 定义见前文；下面补充 `AdultUsers`：

```go
func AdultUsers(db *gorm.DB) *gorm.DB {
	return db.Where("age >= ?", 18)
}

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

func WithMaxAge(maxAge int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if maxAge > 0 {
			return db.Where("age <= ?", maxAge)
		}
		return db
	}
}

func WithCity(city string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if city != "" {
			return db.Where("city = ?", city)
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

	baseQuery := db.Model(&User{}).Scopes(
		WithKeyword(params.Keyword),
		WithStatus(params.Status),
		WithMinAge(params.MinAge),
		WithMaxAge(params.MaxAge),
		WithCity(params.City),
	)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.Scopes(Paginate(params.Page, params.PageSize)).
		Order("created_at DESC, id DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
```

---

## 19.5 常用 Scope 模板库

```go
// ============ 过滤类 ============

// 列名不能用 ? 占位符绑定。调用这类通用 Scope 前，必须由调用方把 column 限制为可信白名单。
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

// field 必须由调用方通过白名单校验；列名不可作为普通 SQL 参数绑定。
func OrderBy(field string, desc bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := clause.OrderByColumn{
			Column: clause.Column{Name: field},
			Desc:   desc,
		}
		return db.Order(order)
	}
}

// import "gorm.io/gorm/clause"

// ============ 分页类 ============

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pageSize)
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

Scope 通常应描述查询构造，不应执行独立副作用或假设 SQL 已经生成：

```go
// 不要在 Scope 构造阶段读取 db.Statement.SQL：此时通常还没有执行 SQL。
// 审计/查询日志应使用 GORM Logger、插件或回调，在 SQL 构建/执行阶段记录。
func AuditLog(db *gorm.DB) *gorm.DB {
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
