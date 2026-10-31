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

先安装本章示例用到的驱动和 Mock 库：

```bash
go get gorm.io/driver/sqlite
go get github.com/DATA-DOG/go-sqlmock
```

```go
import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// shared cache + 保持连接，避免 SQLite 每个连接各自拥有独立的 :memory: 数据库。
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法创建测试数据库: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("无法获取底层数据库连接: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	// 自动建表；测试代码应检查迁移错误
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	return db
}
```

---

## 35.2 CRUD 测试用例

```go
func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	user := User{Name: "张三", Age: 25}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if user.ID == 0 {
		t.Fatalf("期望数据库回填非零主键")
	}
}

func TestFindUser(t *testing.T) {
	db := setupTestDB(t)
	created := User{Name: "张三", Age: 25}
	if err := db.Create(&created).Error; err != nil {
		t.Fatalf("准备测试数据失败: %v", err)
	}

	var user User
	if err := db.First(&user, created.ID).Error; err != nil {
		t.Fatalf("查询用户失败: %v", err)
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

	// 在事务中执行测试；使用同一个事务句柄完成所有断言
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("开启事务失败: %v", tx.Error)
	}
	defer tx.Rollback() // 测试结束自动回滚

	if err := tx.Create(&User{Name: "测试用户"}).Error; err != nil {
		t.Fatalf("创建测试数据失败: %v", err)
	}
	var count int64
	if err := tx.Model(&User{}).Count(&count).Error; err != nil {
		t.Fatalf("统计测试数据失败: %v", err)
	}
	if count != 1 {
		t.Errorf("期望 1 个用户, 得到 %d", count)
	}
	// 事务回滚后数据库中不会有"测试用户"。
	// 此模式适用于测试代码能在同一事务连接中运行的场景；不适合并行测试共享同一事务。
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
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败: %v", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("初始化 GORM 失败: %v", err)
	}

	// GORM 的 First 会生成带主键排序和 LIMIT 的 SQL；Regexp 字符串中的反斜杠需转义。
	mock.ExpectQuery("SELECT .* FROM `users`.*ORDER BY `users`\\.`id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "张三", 25))

	// 执行查询
	var user User
	if err := db.First(&user).Error; err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if user.Name != "张三" {
		t.Fatalf("期望 Name=张三，得到 %q", user.Name)
	}

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
func createTestUser(t *testing.T, db *gorm.DB) *User {
	t.Helper()
	user := &User{Name: "测试用户", Age: 25}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
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

- SQLite 内存数据库适合快速单元测试；使用共享内存 DSN/单连接池避免连接间看不到表
- SQLite 与 MySQL 的类型、SQL 方言、约束和事务行为仍有差异，关键集成测试要使用真实 MySQL
- 事务回滚避免测试间数据污染
- 表驱动测试覆盖多种场景
- sqlmock 用于不依赖真实数据库的单元测试

## 练习题

1. 用 SQLite 内存数据库写一个 Create 操作的测试用例。
2. 用事务回滚方式写一个 Delete 操作的测试用例。
3. 写一个表驱动测试，覆盖 First/Find/Distinct 多种查询场景。
4. （思考题）SQLite 内存数据库测试的局限性是什么？什么情况下必须用真实 MySQL 测试？
