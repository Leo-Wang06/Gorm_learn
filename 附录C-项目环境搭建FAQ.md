# 附录C：项目环境搭建 FAQ

## Go 安装 FAQ

**Q：怎么安装 Go？**
> 官网 [https://go.dev/dl](https://go.dev/dl) 下载对应系统安装包。Windows用`.msi`，Mac用`.pkg`或`brew install go`，Linux解压tar.gz。

**Q：怎么验证安装成功？**
> `go version` 输出版本号即成功。`go env` 查看环境变量。

**Q：GOPATH 必须设置吗？**
> Go Modules 时代（Go 1.16+）不强制设置。项目建在任意目录，用 `go mod init` 初始化即可。

**Q：`go mod init` 有什么用？**
> 初始化 Go Module，生成 `go.mod` 文件，管理项目依赖。类比 npm init、pipenv。

## GORM 安装 FAQ

**Q：GORM 怎么安装？**
```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
```

**Q：下载慢怎么办？**
```bash
# 设置 Go 代理（国内用户推荐）
go env -w GOPROXY=https://goproxy.cn,direct
```

**Q：安装哪个版本？**
> 不指定版本就是最新版。查看版本：`grep gorm go.mod`。

**Q：v1 和 v2 怎么区分？**
> v1：`github.com/jinzhu/gorm`（已停止维护，不要用）
> v2：`gorm.io/gorm`（当前版本，本教程使用）

## MySQL 连接 FAQ

**Q：连接报错 Error 1045？**
> 用户名或密码错误。先用 `mysql -u root -p` 验证能否登录。

**Q：连接报错 Error 1049？**
> 数据库不存在：
```sql
CREATE DATABASE gorm_tutorial CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**Q：时间字段显示为字符串？**
> DSN 缺少 `parseTime=True`。完整DSN：
```
root:123456@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local
```

**Q：时区不对？**
> 加 `loc=Local` 使用系统时区，或 `loc=Asia%2FShanghai`。

**Q：SSL 连接错误？**
> 本地开发在 DSN 加 `tls=skip-verify` 或完全不加 tls 参数。

## 开发工具推荐

| 工具 | 用途 |
|------|------|
| VS Code + Go 插件 | Go 开发主 IDE（免费） |
| GoLand | JetBrains Go IDE（付费，功能强大） |
| Navicat / DBeaver | 数据库可视化工具 |
| TablePlus | Mac 端数据库工具 |
| Postman | API 测试 |
| Docker | 快速启动 MySQL、Redis 等 |

## 快速启动 MySQL（Docker）

```bash
# 一行命令启动 MySQL 8.0
docker run -d \
  --name mysql-gorm \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -e MYSQL_DATABASE=gorm_tutorial \
  -p 3306:3306 \
  mysql:8.0 \
  --default-authentication-plugin=mysql_native_password
```

## 最小可运行项目

```go
package main

import (
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100)"`
	Age  int
}

func main() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/gorm_tutorial?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	db.AutoMigrate(&User{})
	
	db.Create(&User{Name: "张三", Age: 25})
	
	var user User
	db.First(&user)
	log.Printf("用户: %s, 年龄: %d", user.Name, user.Age)
}
```

运行：
```bash
go run main.go
```
