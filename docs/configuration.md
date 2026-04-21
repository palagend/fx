# 配置指南

本文档详细说明 FX Portfolio 的配置方式和使用方法。

## 目录

- [配置文件位置](#配置文件位置)
- [配置项说明](#配置项说明)
- [环境变量](#环境变量)
- [配置优先级](#配置优先级)
- [部署场景](#部署场景)

---

## 配置文件位置

配置文件支持以下方式（按优先级排序）：

| 优先级 | 方式 | 路径 | 适用场景 |
|--------|------|------|----------|
| 1 | 环境变量 | - | 生产环境 |
| 2 | 本地配置 | `./config.yaml` | 开发环境 |
| 3 | 配置目录 | `./config/config.yaml` | 多环境配置 |
| 4 | 系统配置 | `/etc/fx/config.yaml` | 服务器部署 |

---

## 配置项说明

### 完整配置示例

#### MySQL 配置示例

```yaml
# 数据库配置
database:
  type: "mysql"          # 数据库类型: mysql, postgres, sqlite
  host: "localhost"      # 数据库主机地址
  port: 3306             # 数据库端口
  user: "fx_user"        # 数据库用户名
  password: "password"   # 数据库密码（建议使用环境变量）
  name: "fx_db"          # 数据库名
  charset: "utf8mb4"     # 字符集
  log_level: "info"      # 数据库日志级别: silent, error, warn, info
  pool:
    max_idle_conns: 10       # 最大空闲连接数
    max_open_conns: 100      # 最大打开连接数
    conn_max_lifetime: "1h"  # 连接最大生命周期
    conn_max_idle_time: "30m" # 连接最大空闲时间

# API Keys
api:
  coincap_key: ""        # CoinCap API Key，用于获取加密货币价格

# 服务器配置
server:
  port: "8080"           # 服务端口
  mode: "release"        # 运行模式: debug 或 release

# JWT 配置
jwt:
  secret: "your-secret"  # JWT 签名密钥（生产环境必须修改）
  expires_in: 24         # Token 有效期（小时）

# 日志配置
log:
  level: "info"          # 日志级别: debug, info, warn, error
  format: "text"         # 日志格式: text 或 json
```

#### PostgreSQL 配置示例

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "fx_user"
  password: "password"
  name: "fx_db"
  ssl_mode: "disable"    # SSL模式: disable, require, verify-ca, verify-full
  log_level: "info"
  pool:
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: "1h"
    conn_max_idle_time: "30m"
```

#### SQLite 配置示例

```yaml
database:
  type: "sqlite"
  path: "./fx.db"        # SQLite数据库文件路径
  # 或
  # name: "fx_db"        # 如果不指定path，则使用name作为文件名
  log_level: "info"
```

### 配置项详解

#### Database 配置

##### 通用配置项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| type | string | "mysql" | 数据库类型: `mysql`, `postgres`, `sqlite` |
| log_level | string | "" | 数据库日志级别: `silent`, `error`, `warn`, `info` |

##### MySQL 配置项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| host | string | "172.23.112.1" | 数据库主机地址 |
| port | int | 3306 | 数据库端口 |
| user | string | "admin" | 数据库用户名 |
| password | string | "ctsi@Passw0rd" | 数据库密码 |
| name | string | "insight_onchain" | 数据库名 |
| charset | string | "utf8mb4" | 字符集 |

##### PostgreSQL 配置项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| host | string | "172.23.112.1" | 数据库主机地址 |
| port | int | 5432 | 数据库端口 |
| user | string | "admin" | 数据库用户名 |
| password | string | "ctsi@Passw0rd" | 数据库密码 |
| name | string | "insight_onchain" | 数据库名 |
| ssl_mode | string | "disable" | SSL模式: `disable`, `require`, `verify-ca`, `verify-full` |

##### SQLite 配置项

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| path | string | "" | 数据库文件路径（优先级高于 name） |
| name | string | "fx.db" | 数据库文件名（当 path 为空时使用） |

> **注意**: SQLite 是嵌入式数据库，不需要 host、port、user、password 等连接信息。

##### 连接池配置 (pool)

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| max_idle_conns | int | 10 | 最大空闲连接数 |
| max_open_conns | int | 100 | 最大打开连接数 |
| conn_max_lifetime | duration | "1h" | 连接最大生命周期 |
| conn_max_idle_time | duration | "30m" | 连接最大空闲时间 |

> **提示**: 连接池配置适用于 MySQL 和 PostgreSQL，SQLite 不需要连接池配置。

#### API 配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| coincap_key | string | "" | CoinCap API Key |

> **注意**: CoinCap API Key 用于获取加密货币实时价格。如果没有配置，价格查询功能可能受限。

#### Server 配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| port | string | "8080" | HTTP 服务端口 |
| mode | string | "release" | Gin 运行模式: `debug` 或 `release` |

> **提示**: `debug` 模式会输出详细的日志信息，适合开发调试；`release` 模式性能更好，适合生产环境。

#### JWT 配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| secret | string | "your-secret-key..." | JWT 签名密钥 |
| expires_in | int | 24 | Token 有效期（小时） |

> **警告**: 生产环境**必须**修改默认的 JWT Secret，建议使用随机生成的强密码。

#### Log 配置

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| level | string | "info" | 日志级别: `debug`, `info`, `warn`, `error` |
| format | string | "text" | 日志格式: `text` 或 `json` |

---

## 环境变量

所有配置项都支持通过环境变量设置，命名规则为：`FX_` + 大写配置路径（`.` 替换为 `_`）。

### 环境变量列表

```bash
# 数据库通用配置
export FX_DATABASE_TYPE=mysql              # 数据库类型: mysql, postgres, sqlite
export FX_DATABASE_LOG_LEVEL=info          # 数据库日志级别

# MySQL / PostgreSQL 配置
export FX_DATABASE_HOST=localhost
export FX_DATABASE_PORT=3306
export FX_DATABASE_USER=fx_user
export FX_DATABASE_PASSWORD=your-password
export FX_DATABASE_NAME=fx_db

# MySQL 特有配置
export FX_DATABASE_CHARSET=utf8mb4

# PostgreSQL 特有配置
export FX_DATABASE_SSL_MODE=disable

# SQLite 特有配置
export FX_DATABASE_PATH=./fx.db

# 数据库连接池配置
export FX_DATABASE_POOL_MAX_IDLE_CONNS=10
export FX_DATABASE_POOL_MAX_OPEN_CONNS=100
export FX_DATABASE_POOL_CONN_MAX_LIFETIME=1h
export FX_DATABASE_POOL_CONN_MAX_IDLE_TIME=30m

# API Keys
export FX_API_COINCAP_KEY=your-api-key

# 服务器配置
export FX_SERVER_PORT=8080
export FX_SERVER_MODE=release

# JWT 配置
export FX_JWT_SECRET=your-secret-key
export FX_JWT_EXPIRES_IN=24

# 日志配置
export FX_LOG_LEVEL=info
export FX_LOG_FORMAT=text
```

### 生成安全的 JWT Secret

```bash
# 使用 OpenSSL 生成随机密钥
export FX_JWT_SECRET=$(openssl rand -hex 32)

echo $FX_JWT_SECRET
```

---

## 数据库选择指南

### 各数据库特点对比

| 特性 | MySQL | PostgreSQL | SQLite |
|------|-------|------------|--------|
| **部署复杂度** | 中等 | 中等 | 简单 |
| **性能** | 高 | 高 | 中等 |
| **并发能力** | 优秀 | 优秀 | 一般 |
| **数据量支持** | 大规模 | 大规模 | 中小规模 |
| **备份恢复** | 完善 | 完善 | 简单 |
| **适用场景** | 生产环境 | 生产环境 | 开发测试/单机 |

### 推荐场景

#### SQLite - 快速开始

适合场景：
- 本地开发测试
- 单机部署
- 数据量较小（< 1GB）
- 无需远程访问

优点：
- 零配置，开箱即用
- 无需安装数据库服务
- 数据存储在单个文件，便于迁移

```yaml
database:
  type: "sqlite"
  path: "./data/fx.db"
```

#### MySQL - 生产环境推荐

适合场景：
- 生产环境部署
- 高并发访问
- 需要主从复制
- 团队熟悉 MySQL

优点：
- 成熟的生态系统
- 丰富的运维工具
- 良好的性能表现

```yaml
database:
  type: "mysql"
  host: "mysql.example.com"
  port: 3306
  user: "fx_user"
  password: "${DB_PASSWORD}"
  name: "fx_db"
  charset: "utf8mb4"
```

#### PostgreSQL - 高级功能需求

适合场景：
- 需要复杂查询
- 需要 JSON/数组类型支持
- 需要地理空间数据支持
- 需要高级数据类型

优点：
- 强大的 SQL 支持
- 丰富的数据类型
- 优秀的扩展性

```yaml
database:
  type: "postgres"
  host: "postgres.example.com"
  port: 5432
  user: "fx_user"
  password: "${DB_PASSWORD}"
  name: "fx_db"
  ssl_mode: "require"
```

---

## 配置优先级

配置加载优先级（高 → 低）：

```
1. 环境变量（如 FX_DATABASE_PASSWORD）
   ↓
2. 配置文件（config.yaml）
   ↓
3. 代码默认值
```

**优先级说明：**
- 环境变量优先级最高，会覆盖配置文件中的同名配置
- 如果环境变量未设置，则使用配置文件中的值
- 如果配置文件不存在或配置项缺失，则使用代码默认值

---

## 部署场景

### 本地开发

```bash
# 1. 复制示例配置文件
cp config.yaml.example config.yaml

# 2. 编辑本地配置
vim config.yaml

# 3. 运行
go run main.go
```

### Docker 部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080

# 通过环境变量传入配置
ENV FX_SERVER_PORT=8080
ENV FX_SERVER_MODE=release

CMD ["./main"]
```

```yaml
# docker-compose.yml
version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - FX_DATABASE_HOST=mysql
      - FX_DATABASE_PASSWORD=${DB_PASSWORD}
      - FX_JWT_SECRET=${JWT_SECRET}
      - FX_API_COINCAP_KEY=${COINCAP_KEY}
```

### Kubernetes 部署

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: fx-config
type: Opaque
stringData:
  FX_DATABASE_PASSWORD: "your-db-password"
  FX_JWT_SECRET: "your-jwt-secret"
  FX_API_COINCAP_KEY: "your-api-key"
```

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fx-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fx
  template:
    metadata:
      labels:
        app: fx
    spec:
      containers:
      - name: fx
        image: fx:latest
        ports:
        - containerPort: 8080
        envFrom:
        - secretRef:
            name: fx-config
        env:
        - name: FX_SERVER_PORT
          value: "8080"
        - name: FX_SERVER_MODE
          value: "release"
```

### CI/CD 流水线

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy
        run: |
          # 设置环境变量
          export FX_DATABASE_PASSWORD="${{ secrets.DB_PASSWORD }}"
          export FX_JWT_SECRET="${{ secrets.JWT_SECRET }}"
          export FX_API_COINCAP_KEY="${{ secrets.COINCAP_KEY }}"
          
          # 启动服务
          ./main
```

---

## 配置验证

启动时会自动验证配置：

```
✓ 数据库配置验证通过
✓ JWT 配置验证通过
⚠ 警告: CoinCap API Key 未配置，价格查询功能可能受限
✓ 服务器配置验证通过
```

如果配置验证失败，程序会输出错误信息并退出。

---

## 常见问题

**Q: 配置文件和示例文件有什么区别？**
A: `config.yaml.example` 是示例文件，包含默认配置和注释，可以提交到代码仓库；`config.yaml` 是实际使用的配置文件，包含敏感信息，已添加到 `.gitignore`。

**Q: 如何在不修改配置文件的情况下修改配置？**
A: 使用环境变量覆盖，例如：`export FX_DATABASE_HOST=newhost`。

**Q: 生产环境应该如何管理敏感配置？**
A: 推荐使用环境变量或专门的密钥管理服务（如 AWS Secrets Manager、HashiCorp Vault）。

---

*文档版本: 1.0*
*最后更新: 2026年4月19日*
