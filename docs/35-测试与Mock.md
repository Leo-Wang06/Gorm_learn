# 第35章：测试与 Mock

## 本章目标
学完本章后，你将能够：
1. 使用 SQLite 内存数据库进行单元测试
2. 编写测试辅助函数
3. 使用事务回滚避免测试数据污染
4. 了解 go-sqlmock 的基本用法
5. 为 CRUD 操作编写测试用例

## 前置知识
- 需要先学习：第09-13章（CRUD）
- 需要了解：Go 的 `testing` 包基础

---

## 35.1 使用 SQLite 内存数据库测试

```go
import (
	"testing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	// 自动建表
	db.AutoMigrate(&User{})
	return db
}
```

---

## 35.2 CRUD 测试用例

```go
func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	user := User{Name: "张三", Age: 25}
	result := db.Create(&user)

	if result.Error != nil {
		t.Errorf("创建用户失败: %v", result.Error)
	}
	if user.ID != 1 {
		t.Errorf("期望 ID=1, 得到 ID=%d", user.ID)
	}
}

func TestFindUser(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&User{Name: "张三", Age: 25})

	var user User
	result := db.First(&user, 1)

	if result.Error != nil {
		t.Errorf("查询用户失败: %v", result.Error)
	}
	if user.Name != "张三" {
		t.Errorf("期望 Name='张三', 得到 '%s'", user.Name)
	}
}
```

---

## 35.3 使用事务回滚隔离测试

```go
func TestWithTransactionRollback(t *testing.T) {
	db := setupTestDB(t)

	// 在事务中执行测试
	tx := db.Begin()
	defer tx.Rollback()  // 测试结束自动回滚，不留垃圾数据

	// 测试操作...
	tx.Create(&User{Name: "测试用户"})
	var count int64
	tx.Model(&User{}).Count(&count)
	if count != 1 {
		t.Errorf("期望 1 个用户, 得到 %d", count)
	}
	// 事务回滚后数据库中不会有"测试用户"
}
```

---

## 35.4 go-sqlmock 简介

```go
import (
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
)

func TestWithSqlmock(t *testing.T) {
	// 创建 mock 数据库连接
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	db, _ := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	// 设置期望的 SQL
	mock.ExpectQuery("SELECT .* FROM `users`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "张三", 25))

	// 执行查询
	var user User
	db.First(&user)

	// 验证期望
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("期望未满足: %v", err)
	}
}
```

---

## 35.5 测试辅助工具

```go
// 创建测试数据
func createTestUser(db *gorm.DB) *User {
	user := &User{Name: "测试用户", Age: 25}
	db.Create(user)
	return user
}

// 表驱动测试
func TestFindUserByAge(t *testing.T) {
	db := setupTestDB(t)
	db.Create(&User{Name: "张三", Age: 18})
	db.Create(&User{Name: "李四", Age: 25})
	db.Create(&User{Name: "王五", Age: 30})

	tests := []struct {
		name    string
		minAge  int
		wantLen int
	}{
		{"全部成年人", 18, 3},
		{"25岁以上", 25, 2},
		{"30岁以上", 30, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var users []User
			db.Where("age >= ?", tt.minAge).Find(&users)
			if len(users) != tt.wantLen {
				t.Errorf("期望 %d 条, 得到 %d 条", tt.wantLen, len(users))
			}
		})
	}
}
```

---

## 本章小结

- SQLite `:memory:` 是测试首选（快速、隔离、无需真实 MySQL）
- 事务回滚避免测试间数据污染
- 表驱动测试覆盖多种场景
- sqlmock 用于不依赖真实数据库的单元测试

## 练习题

1. 用 SQLite 内存数据库写一个 Create 操作的测试用例。
2. 用事务回滚方式写一个 Delete 操作的测试用例。
3. 写一个表驱动测试，覆盖 First/Find/Distinct 多种查询场景。
4. （思考题）SQLite 内存数据库测试的局限性是什么？什么情况下必须用真实 MySQL 测试？
