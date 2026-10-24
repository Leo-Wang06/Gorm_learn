# 附录B：GORM 常见错误排查指南

## 连接错误

| 错误 | 现象 | 原因 | 解决 |
|------|------|------|------|
| Error 1045 | Access denied | 用户名/密码错误 | 验证 MySQL 凭据 |
| Error 1049 | Unknown database | 数据库不存在 | 创建数据库 |
| connection refused | dial tcp: connect: connection refused | MySQL 未启动/端口错误 | 启动 MySQL，检查端口 |
| TLS error | TLS handshake timeout | SSL配置问题 | DSN加 `tls=skip-verify` 或配置SSL |
| `parseTime` | 时间字段是字符串 | 忘了 `parseTime=True` | DSN加 `parseTime=True` |

## 迁移错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 列没有被删除 | AutoMigrate 不做减法 | 用 Migrator().DropColumn 手动删 |
| 类型修改失败 | MySQL有限制 | 检查类型兼容性，手动 ALTER |
| Error 1067 | default值无效 | 检查default值格式（字符串要加引号） |

## 查询错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 零值条件不生效 | struct零值被跳过 | 用map或字符串条件 |
| ErrRecordNotFound | First找不到记录 | 用errors.Is判断，正常处理 |
| Preload 不生效 | 字段名拼错/未定义关联 | 检查关联名和关联定义 |
| N+1 查询 | 循环中查询关联 | 用 Preload 或 Joins |

## 更新错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 零值字段不更新 | struct Updates跳过零值 | 用 map 或 Select |
| Save 覆盖字段 | Save 全字段更新 | 改用 Update/Updates |
| ErrMissingWhereClause | 批量更新没加WHERE | 加Where或用AllowGlobalUpdate |
| Hook不触发 | 用了UpdateColumn | 改用Update（如需触发Hook） |

## 关联错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 外键未被识别 | 字段名不匹配约定 | 显式指定 foreignKey |
| 关联创建未写入 | 忘记 Select("Relation") | 创建时加 Select |
| 中间表名不一致 | 两边 many2many 表名不同 | 保持两边一致 |
| Association 不生效 | 忘记 Model() | 加 db.Model(&record) |

## 事务错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 事务无效 | 闭包中用db而非tx | 闭包中必须用tx |
| 手动事务未Rollback | 忘记Rollback | 用defer或改用Transaction |
| 事务超时 | 事务执行时间过长 | 拆分事务或优化SQL |

## 性能问题

| 问题 | 排查方向 |
|------|---------|
| 查询慢 | 检查索引、EXPLAIN、慢查询日志 |
| 连接数飙升 | 检查连接池配置（MaxOpenConns） |
| 连接泄漏 | 检查 rows.Close()、事务提交 |
| 内存高 | 检查大结果集是否分批处理 |
| N+1 | 用 Debug() 查看SQL条数 |
