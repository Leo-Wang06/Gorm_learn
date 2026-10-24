# 附录A：GORM 常用标签速查表

## 字段定义标签

| 标签 | 用途 | 示例 |
|------|------|------|
| `column:name` | 自定义列名 | `gorm:"column:user_name"` |
| `type:sql_type` | 指定数据库类型 | `gorm:"type:varchar(200)"` |
| `size:n` | 列大小 | `gorm:"size:255"` |
| `precision:n` | 小数总位数 | `gorm:"precision:10"` |
| `scale:n` | 小数位数 | `gorm:"scale:2"` |
| `serializer:json` | JSON序列化 | `gorm:"serializer:json"` |
| `serializer:unixtime` | Unix时间戳 | `gorm:"serializer:unixtime"` |
| `serializer:gob` | Gob序列化 | `gorm:"serializer:gob"` |
| `embedded` | 嵌入结构体 | `gorm:"embedded"` |
| `embeddedPrefix:p` | 嵌入字段前缀 | `gorm:"embeddedPrefix:audit_"` |

## 约束与索引标签

| 标签 | 用途 | 示例 |
|------|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `autoIncrement` | 自增 | `gorm:"autoIncrement"` |
| `not null` | 非空 | `gorm:"not null"` |
| `default:val` | 默认值 | `gorm:"default:18"` |
| `unique` | 唯一约束 | `gorm:"unique"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `uniqueIndex:name` | 命名唯一索引 | `gorm:"uniqueIndex:idx_email"` |
| `index` | 普通索引 | `gorm:"index"` |
| `index:name` | 命名索引 | `gorm:"index:idx_name"` |
| `check:expr` | CHECK约束 | `gorm:"check:age > 0"` |
| `comment:text` | 列注释 | `gorm:"comment:用户名"` |

## 读写控制标签

| 标签 | 含义 |
|------|------|
| `<-` | 可读可写（默认） |
| `<-:false` | 禁止写入 |
| `<-:create` | 只允许创建时写入 |
| `<-:update` | 只允许更新时写入 |
| `->` | 只读 |
| `->:false` | 禁止读取 |
| `-` | 完全忽略 |
| `-:all` | 所有操作忽略 |
| `-:migration` | 迁移忽略 |

## 关联标签

| 标签 | 用途 |
|------|------|
| `foreignKey:Field` | 指定外键字段 |
| `references:Field` | 指定引用字段 |
| `many2many:table` | 多对多中间表名 |
| `joinForeignKey:Field` | 中间表中指向当前表的外键 |
| `joinReferences:Field` | 中间表中指向关联表的外键 |
| `polymorphic:Field` | 多态关联 |
| `polymorphicValue:val` | 多态类型值 |
| `constraint:OnUpdate:CASCADE,...` | 外键约束行为 |

## Go类型 ↔ MySQL类型映射

| Go 类型 | 默认 MySQL 类型 |
|---------|----------------|
| `int`, `uint` | BIGINT |
| `int8` | TINYINT |
| `int16` | SMALLINT |
| `int32` | INT |
| `int64` | BIGINT |
| `float32` | FLOAT |
| `float64` | DOUBLE |
| `string` | VARCHAR(255) |
| `bool` | TINYINT(1) |
| `time.Time` | DATETIME(3) |
| `[]byte` | LONGBLOB |
| `*time.Time` | DATETIME(3) NULL |

## 常用链式方法速查

| 方法 | 说明 |
|------|------|
| `db.Where()` | 条件查询 |
| `db.Not()` | 否定条件 |
| `db.Or()` | OR条件 |
| `db.Order()` | 排序 |
| `db.Limit()` | 限制行数 |
| `db.Offset()` | 偏移 |
| `db.Select()` | 选择字段 |
| `db.Omit()` | 排除字段 |
| `db.Group()` | 分组 |
| `db.Having()` | 分组过滤 |
| `db.Joins()` | 连接查询 |
| `db.Preload()` | 预加载关联 |
| `db.Scopes()` | 作用域 |
| `db.Debug()` | 打印SQL |
| `db.Raw()` | 原生查询 |
| `db.Exec()` | 原生执行 |
| `db.Scan()` | 扫描结果 |
| `db.Pluck()` | 提取单列 |
| `db.Count()` | 计数 |
| `db.Distinct()` | 去重 |
