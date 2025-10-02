# genGormStruct

基于 [GORM Gen](https://gorm.io/gen/) 的数据库模型代码生成工具，支持从 MySQL 数据库自动生成 GORM v2 模型代码。

## 说明

- 注意: 官方已经提供了 `gentool` 工具，建议使用官方工具。
- 参考官方文档: https://gorm.io/gen/gen_tool.html
```bash
go install gorm.io/gen/tools/gentool@latest
```

## 功能特性

- 自动生成 GORM v2 模型代码
- 支持 YAML 配置文件
- 命令行 flag 与配置文件灵活组合
- 自动跳过临时表（`_tmp`、`_bak` 后缀）
- 为每个表生成独立的模型文件
- 自动生成单元测试代码

## 安装

```bash
go build -o genGormStruct .
```

## 使用方法

### 1. 生成默认配置文件

```bash
./genGormStruct init-config
```

执行后会生成 `config.yaml` 文件，包含所有默认配置项。

### 2. 使用配置文件运行

```bash
./genGormStruct -c config.yaml
```

### 3. 使用命令行参数运行

```bash
# 基础参数方式
./genGormStruct -host 127.0.0.1 -port 3306 -user root -password secret -dbname mydb

# 或使用完整 DSN
./genGormStruct -dsn "user:password@tcp(host:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

### 4. 配置文件 + 命令行覆盖

```bash
# 使用配置文件，但覆盖主机和密码
./genGormStruct -c config.yaml -host 192.168.1.100 -password "newpassword"
```

## 配置优先级

配置优先级从高到低：

1. **命令行 flag**（如 `-host`, `-port`, `-password`, `-dsn` 等）
2. **配置文件**（`-c` 指定的 YAML 文件）
3. **内置默认值**

## 配置文件说明

### 配置文件结构

```yaml
# genGormStruct 默认配置文件
# 命令行 flag 优先级高于配置文件

database:
  host: "127.0.0.1"      # 数据库主机地址
  port: "3306"           # 数据库端口
  user: "root"           # 数据库用户名
  password: "p@ssw0rd"   # 数据库密码
  dbname: "dbname"       # 数据库名称
  # dsn: ""              # 自定义 DSN（如果设置，将忽略上述配置）

output:
  dir: "./outputs"       # 输出目录
  pkgname: "gorm2model"  # 生成的包名
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `database.host` | MySQL 主机地址 | `127.0.0.1` |
| `database.port` | MySQL 端口 | `3306` |
| `database.user` | MySQL 用户名 | `root` |
| `database.password` | MySQL 密码 | `p@ssw0rd` |
| `database.dbname` | 数据库名称 | `dbname` |
| `database.dsn` | 自定义 DSN（优先级最高） | `""` |
| `output.dir` | 代码输出目录 | `./outputs` |
| `output.pkgname` | 生成的 Go 包名 | `gorm2model` |

## 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-c` | 指定配置文件路径 | `-c config.yaml` |
| `-host` | 数据库主机地址 | `-host 127.0.0.1` |
| `-port` | 数据库端口 | `-port 3306` |
| `-user` | 数据库用户名 | `-user root` |
| `-password` | 数据库密码 | `-password secret` |
| `-dbname` | 数据库名称 | `-dbname mydb` |
| `-dsn` | 完整数据库连接串 | `-dsn "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"` |
| `-output` | 输出目录 | `-output ./models` |
| `-pkgname` | 生成的包名 | `-pkgname mymodel` |
| `-h` | 显示帮助信息 | `-h` |

## 使用示例

### 示例 1：快速开始（使用默认值）

```bash
# 使用默认配置连接本地 MySQL
./genGormStruct
```

### 示例 2：使用配置文件

```bash
# 第一步：生成配置文件
./genGormStruct init-config

# 第二步：编辑 config.yaml，填写实际的数据库信息
# vim config.yaml

# 第三步：运行
./genGormStruct -c config.yaml
```

### 示例 3：命令行直接指定

```bash
./genGormStruct \
  -host 192.168.1.100 \
  -port 3306 \
  -user admin \
  -password "MyP@ssw0rd" \
  -dbname production_db \
  -output ./models \
  -pkgname prodmodel
```

### 示例 4：使用 DSN（最灵活）

```bash
./genGormStruct -dsn "admin:secret@tcp(192.168.1.100:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
```

### 示例 5：混合使用（配置 + 覆盖）

```bash
# 使用配置文件的 most 设置，但覆盖主机和输出目录
./genGormStruct -c config.yaml -host 10.0.0.5 -output ./tmp_models
```

## 输出结构

生成后的代码结构：

```
outputs/
└── gorm2model/
    ├── gen.go              # 生成的查询接口
    ├── tablename1.gen.go   # 表1的模型
    ├── tablename2.gen.go   # 表2的模型
    └── ...
```

## 注意事项

1. **DSN 优先级**：如果设置了 `-dsn` 或配置文件中的 `database.dsn`，其他数据库连接参数将被忽略
2. **表过滤**：自动跳过包含 `.` 的表名或以 `_tmp`、`_bak` 结尾的表
3. **重复运行**：可以多次运行，生成的文件会被覆盖
4. **配置文件生成**：`init-config` 命令不会覆盖已存在的 `config.yaml` 文件

## 依赖

- Go 1.24+
- MySQL 5.7+
- GORM v2

## 技术栈

- [GORM Gen](https://gorm.io/gen/) - 代码生成工具
- [GORM](https://gorm.io/) - Go ORM 库
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 解析
