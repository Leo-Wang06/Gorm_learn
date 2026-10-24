# 第34章：SQL 注入防护与安全实践

## 本章目标
学完本章后，你将能够：
1. 理解 SQL 注入的原理
2. 使用 GORM 安全地处理用户输入
3. 安全使用 `Raw` 和 `Exec`
4. 正确存储密码（bcrypt）
5. 脱敏敏感字段

## 前置知识
- 需要先学习：第11章（WHERE 查询）、第14章（Raw SQL）

---

## 34.1 SQL 注入回顾

```go
// ⚠️ 危险！SQL 注入示例
name := req.URL.Query().Get("name")
// 攻击者输入：' OR '1'='1' --

db.Raw("SELECT * FROM users WHERE name = '" + name + "'").Scan(&users)
// 实际执行：SELECT * FROM users WHERE name = '' OR '1'='1' --'
// 结果：返回所有用户！
```

---

## 34.2 GORM 的安全机制

GORM 默认使用参数化查询（占位符 `?`），自动防注入：

```go
// ✅ 安全：占位符转义
db.Where("name = ?", name).Find(&users)
// name 中的单引号被自动转义，无法注入

// ✅ 安全：结构体条件
db.Where(&User{Name: name}).Find(&users)

// ✅ 安全：Map 条件
db.Where(map[string]interface{}{"name": name}).Find(&users)
```

---

## 34.3 Raw/Exec 的安全使用

```go
// ✅ 安全：Raw 也支持占位符
db.Raw("SELECT * FROM users WHERE name = ?", name).Scan(&users)

// ✅ 安全：多参数
db.Raw("SELECT * FROM users WHERE name = ? AND status = ?", name, status).Scan(&users)

// ❌ 危险：字符串拼接
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
db.Raw(query).Scan(&users)  // 注入风险！

// ❌ 危险：用户输入拼到 ORDER BY（无法用占位符）
// 必须先白名单校验
allowedSorts := map[string]bool{"id": true, "name": true, "age": true}
if !allowedSorts[sortBy] {
	return errors.New("无效的排序字段")
}
db.Order(sortBy + " desc").Find(&users)
```

---

## 34.4 密码安全存储

```go
import "golang.org/x/crypto/bcrypt"

// 注册时加密
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// 登录时验证
func CheckPassword(db *gorm.DB, username, password string) (*User, error) {
	var user User
	db.Where("username = ?", username).First(&user)
	
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}
	return &user, nil
}
```

---

## 34.5 敏感字段脱敏

```go
type User struct {
	Password string `gorm:"->:false"`     // 不可读
	Phone    string `gorm:"->:false"`     // 不可读
	Secret   string `gorm:"-:all"`        // 完全忽略
	APIKey   string `gorm:"-:all"`        // 完全忽略
}

// 或 JSON 序列化时脱敏
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`  // 不输出
}
```

---

## 34.6 安全检查清单

- ✅ 永远使用 `?` 占位符，不拼接字符串
- ✅ 排序字段、表名、列名等无法用占位符的参数，使用白名单
- ✅ 密码使用 bcrypt 存储
- ✅ 敏感字段使用 `->:false` 或 `-:all`
- ✅ 生产环境禁用 AutoMigrate
- ✅ 数据库账号使用最小权限原则
- ✅ 连接 DSN 不硬编码密码（用环境变量）

---

## 本章小结

- GORM 参数化查询（`?`）防止 SQL 注入
- Raw/Exec 也要用占位符
- 排序字段等需白名单校验
- 密码用 bcrypt 加密存储
- 敏感字段用 `->:false` 或 `-:all`

## 练习题

1. 写一个"不安全的"查询（字符串拼接），然后用安全的参数化查询重写。
2. 实现一个用户注册/登录流程，使用 bcrypt 加密密码。
3. 配置 User 模型，使密码字段不可读取，手机号不参与 JSON 序列化。
4. （思考题）为什么 ORDER BY 和表名不能使用占位符？如果用户需要自定义排序，如何安全地实现？
