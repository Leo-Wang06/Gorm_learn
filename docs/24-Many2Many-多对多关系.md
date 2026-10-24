# 第24章：Many2Many —— 多对多关系

## 本章目标
学完本章后，你将能够：
1. 理解多对多关系和中间表的概念
2. 定义 Many2Many 关联
3. 自定义中间表名称和结构
4. 在中间表上存储额外字段
5. 操作多对多关联（选课、标签等场景）

## 前置知识
- 需要先学习：第21-23章（BelongsTo、HasOne、HasMany）
- 需要了解：MySQL 多对多关系的中间表设计

---

## 24.1 Many2Many：多对多关系

多对多关系需要一个**中间表**来存储两边的关联：

```
┌──────────┐         ┌──────────────┐         ┌──────────┐
│ Student  │         │student_course │         │  Course  │
│  id: 1   │ ───────>│ student_id:1  │<─────── │  id: 1   │
│  name:张三│         │ course_id: 1  │         │  name:数学│
└──────────┘         │ student_id:1  │         └──────────┘
                     │ course_id: 2  │
┌──────────┐         │ student_id:2  │         ┌──────────┐
│ Student  │         │ course_id: 1  │         │  Course  │
│  id: 2   │ ───────>│ student_id:2  │<─────── │  id: 2   │
│  name:李四│         │ course_id: 3  │         │  name:英语│
└──────────┘         └──────────────┘         └──────────┘
```

解读：
- 张三选修了数学和英语
- 李四选修了数学和物理
- 数学有 2 个学生选修
- 英语有 1 个学生选修

### 基础示例：Student 选修 Course

```go
type Student struct {
	ID      uint     `gorm:"primaryKey"`
	Name    string
	Courses []Course `gorm:"many2many:student_courses"` // 多对多
}

type Course struct {
	ID       uint     `gorm:"primaryKey"`
	Name     string
	Students []Student `gorm:"many2many:student_courses"` // 反向多对多
}
```

生成的表结构：
```sql
-- 学生表
CREATE TABLE students (id BIGINT PRIMARY KEY, name VARCHAR(255));

-- 课程表
CREATE TABLE courses (id BIGINT PRIMARY KEY, name VARCHAR(255));

-- 中间表（自动创建，不需要定义 Go 结构体）
CREATE TABLE student_courses (
    student_id BIGINT,
    course_id  BIGINT,
    PRIMARY KEY (student_id, course_id),  -- 复合主键
    FOREIGN KEY (student_id) REFERENCES students(id),
    FOREIGN KEY (course_id) REFERENCES courses(id)
);
```

---

## 24.2 Many2Many 标签配置

```go
type Student struct {
	ID      uint
	Name    string
	Courses []Course `gorm:"many2many:student_courses;foreignKey:ID;joinForeignKey:StudentID;References:ID;joinReferences:CourseID"`
}
```

标签参数解析：

| 标签 | 含义 | 默认值 |
|------|------|--------|
| `many2many:table` | 中间表名 | 按两表名字母序排列组合 |
| `foreignKey` | 当前表的主键 | `ID` |
| `joinForeignKey` | 中间表中指向当前表的外键 | `{当前表名}ID` |
| `References` | 关联表的主键 | `ID` |
| `joinReferences` | 中间表中指向关联表的外键 | `{关联表名}ID` |

简化写法（使用默认约定）：
```go
type Student struct {
	Courses []Course `gorm:"many2many:student_courses"`
}
// 所有外键和引用都按约定自动推断
```

---

## 24.3 创建多对多关联

### 创建学生并选课

```go
// 先创建课程
math := Course{Name: "高等数学"}
english := Course{Name: "大学英语"}
physics := Course{Name: "大学物理"}
db.Create(&math)
db.Create(&english)
db.Create(&physics)

// 创建学生并选课
student := Student{
	Name: "张三",
	Courses: []Course{math, english},  // 选两门课
}
db.Select("Courses").Create(&student)
// GORM 自动在中间表插入两条记录：
// INSERT INTO student_courses (student_id, course_id) VALUES (1, 1), (1, 2)
```

### 给已有学生添加课程

```go
var student Student
db.First(&student, 1)

// 方式一：Association Append
db.Model(&student).Association("Courses").Append(&physics)

// 方式二：直接操作（用关联 ID）
// GORM 会自动更新中间表
```

---

## 24.4 查询多对多关联

```go
// 查询学生及其选修课程
var students []Student
db.Preload("Courses").Find(&students)
// SQL: SELECT * FROM `students`
// SQL: SELECT * FROM `courses`
//      JOIN `student_courses` ON `student_courses`.`course_id` = `courses`.`id`
//      WHERE `student_courses`.`student_id` IN (1,2,3)

for _, s := range students {
	fmt.Printf("%s 选修了:\n", s.Name)
	for _, c := range s.Courses {
		fmt.Printf("  - %s\n", c.Name)
	}
}

// 查询选修了某门课的学生
var course Course
db.Preload("Students").First(&course, 1)
fmt.Printf("《%s》的学生:\n", course.Name)
for _, s := range course.Students {
	fmt.Printf("  - %s\n", s.Name)
}
```

---

## 24.5 中间表携带额外字段

默认中间表只有两个外键。如果需要额外信息（如选课时间、成绩），需要自定义中间表：

```go
// 定义中间表模型
type StudentCourse struct {
	StudentID uint    `gorm:"primaryKey"`
	CourseID  uint    `gorm:"primaryKey"`
	Score     float64 // 成绩
	EnrolledAt time.Time  // 选课时间
}

type Student struct {
	ID      uint
	Name    string
	Courses []Course `gorm:"many2many:student_courses"`
}

type Course struct {
	ID       uint
	Name     string
	Students []Student `gorm:"many2many:student_courses"`
}

// 注册自定义中间表
db.SetupJoinTable(&Student{}, "Courses", &StudentCourse{})
```

### 操作带额外字段的中间表

```go
// 添加选课记录（带成绩）
studentCourse := StudentCourse{
	StudentID: 1,
	CourseID:  1,
	Score:     92.5,
	EnrolledAt: time.Now(),
}
db.Create(&studentCourse)

// 查询成绩
var records []StudentCourse
db.Where("student_id = ?", 1).Find(&records)
for _, r := range records {
	fmt.Printf("课程%d: %.1f分\n", r.CourseID, r.Score)
}
```

---

## 24.6 完整示例：学生选课系统

```go
package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Student struct {
	ID      uint     `gorm:"primaryKey"`
	Name    string   `gorm:"type:varchar(100)"`
	Courses []Course `gorm:"many2many:student_courses"`
}

type Course struct {
	ID       uint     `gorm:"primaryKey"`
	Name     string   `gorm:"type:varchar(200)"`
	Credits  int      `gorm:"default:3"`
	Students []Student `gorm:"many2many:student_courses"`
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	db.AutoMigrate(&Student{}, &Course{})

	// 1. 创建课程
	courses := []Course{
		{Name: "高等数学", Credits: 4},
		{Name: "大学英语", Credits: 3},
		{Name: "大学物理", Credits: 4},
		{Name: "计算机基础", Credits: 3},
	}
	db.Create(&courses)

	// 2. 创建学生并选课
	students := []Student{
		{Name: "张三", Courses: []Course{courses[0], courses[1], courses[3]}},
		{Name: "李四", Courses: []Course{courses[0], courses[2]}},
		{Name: "王五", Courses: []Course{courses[1], courses[3]}},
	}
	for i := range students {
		db.Select("Courses").Create(&students[i])
	}

	// 3. 查询每个学生选课情况
	var allStudents []Student
	db.Preload("Courses").Find(&allStudents)
	for _, s := range allStudents {
		totalCredits := 0
		for _, c := range s.Courses {
			totalCredits += c.Credits
		}
		fmt.Printf("%s 选修 %d 门课, 共 %d 学分\n",
			s.Name, len(s.Courses), totalCredits)
	}

	// 4. 查询每门课的选课人数
	for _, c := range courses {
		var count int64
		db.Model(&c).Association("Students").Count()
		_ = count
	}

	// 5. 张三退选"大学英语"
	var zhangsan Student
	db.First(&zhangsan, "name = ?", "张三")
	db.Model(&zhangsan).Association("Courses").Delete(&courses[1])
	// 中间表记录被删除，但课程本身不受影响
}
```

---

## 常见错误

### 错误1：中间表名忘记统一

```go
// ❌ 两边定义不同中间表名
type Student struct {
	Courses []Course `gorm:"many2many:student_courses"`
}
type Course struct {
	Students []Student `gorm:"many2many:course_student"`  // 不同的表名！
}

// ✅ 两边必须一致
type Student struct {
	Courses []Course `gorm:"many2many:student_courses"`
}
type Course struct {
	Students []Student `gorm:"many2many:student_courses"`  // 相同！
}
```

### 错误2：Create 时忘记 `Select("Courses")`

```go
student := Student{
	Name:    "张三",
	Courses: []Course{{Name: "数学"}, {Name: "英语"}},
}
db.Create(&student)
// ❌ 学生创建了，选课记录没插入！

// ✅ 加 Select
db.Select("Courses").Create(&student)
```

### 错误3：误以为 Delete 会删除关联对象

```go
// 从中间表删除关联（课程还在，学生也还在）
db.Model(&student).Association("Courses").Delete(&course)
// 只删中间表记录，不删课程

// 删除课程本身（如果中间表有外键约束，可能会失败或级联删除）
db.Delete(&course)
```

---

## 本章小结

- Many2Many 通过中间表实现，中间表包含两个外键
- 标签：`gorm:"many2many:中间表名;joinForeignKey:...;joinReferences:..."`
- 创建关联用 `db.Select("Courses").Create(&student)`
- 操作关联用 `Association().Append/Delete/Replace/Clear`
- 中间表有额外字段时，手动定义中间表模型 + `SetupJoinTable`
- 两边 many2many 的中间表名要一致

## 练习题

1. 定义 `Article` 和 `Tag` 模型，使用 Many2Many 关联（一篇文章可有多个标签，一个标签可属于多篇文章）。
2. 创建 3 篇文章和 5 个标签，为每篇文章添加 2-3 个标签。
3. 查询带有标签的文章列表，打印每篇文章的标签。
4. 查询带有特定标签的所有文章。
5. 添加中间表字段：如果标签关联需要记录"添加时间"，怎么实现？
6. （思考题）多对多关系在业务中的常见场景有哪些？举 3 个例子，并说说它们是否需要中间表额外字段。
