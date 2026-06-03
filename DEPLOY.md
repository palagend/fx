# Vercel 部署指南

## 项目结构

```
.
├── api/                    # Vercel Serverless Functions (Go)
│   ├── *.go               # API 端点处理器
│   ├── db/                # 数据库连接 (MySQL)
│   ├── models/            # 数据模型
│   ├── utils/             # 工具函数
│   └── middleware/        # 中间件
├── web/                   # Vue 3 前端
│   ├── src/
│   └── package.json
└── vercel.json           # Vercel 配置
```

## 前置要求

1. 注册 [Vercel](https://vercel.com) 账号
2. 安装 Vercel CLI:
   ```bash
   npm i -g vercel
   ```
3. 准备 MySQL 数据库（你已有免费数据库）

## 数据库配置

你的 MySQL 数据库信息：
- Host: `mysql6.sqlpub.com`
- Port: `3311`
- User: `palagend`
- Password: `Change-Your-Password`
- Database: `palagend`
- Charset: `utf8mb4`

### 配置环境变量（二选一）

**方案 A：使用完整的 DATABASE_URL（推荐）**
```bash
vercel env add DATABASE_URL
# 输入: palagend:JxIFhCdzBLcEdSC5@tcp(mysql6.sqlpub.com:3311)/palagend?charset=utf8mb4&parseTime=True&loc=Local
```

**方案 B：使用分开的环境变量**
```bash
vercel env add DB_HOST
# 输入: mysql6.sqlpub.com

vercel env add DB_PORT
# 输入: 3311

vercel env add DB_USER
# 输入: palagend

vercel env add DB_PASSWORD
# 输入: JxIFhCdzBLcEdSC5

vercel env add DB_NAME
# 输入: palagend

vercel env add DB_CHARSET
# 输入: utf8mb4
```

### JWT 密钥
```bash
vercel env add JWT_SECRET
# 输入一个随机字符串，例如: your-random-secret-key-12345
```

## 部署步骤

### 1. 本地测试

```bash
# 安装依赖
cd api && go mod tidy && cd ..
cd web && npm install && cd ..

# 本地开发（设置环境变量）
export DB_HOST=mysql6.sqlpub.com
export DB_PORT=3311
export DB_USER=palagend
export DB_PASSWORD=JxIFhCdzBLcEdSC5
export DB_NAME=palagend
export DB_CHARSET=utf8mb4
export JWT_SECRET=your-secret-key

vercel dev
```

### 2. 部署到生产环境

```bash
# 登录 Vercel
vercel login

# 部署
vercel --prod
```

## API 端点列表

| 端点 | 方法 | 描述 | 认证 |
|------|------|------|------|
| `/api/health` | GET | 健康检查 | 否 |
| `/api/auth/register` | POST | 用户注册 | 否 |
| `/api/auth/login` | POST | 用户登录 | 否 |
| `/api/auth/refresh` | POST | 刷新令牌 | 否 |
| `/api/auth/logout` | POST | 用户登出 | 否 |
| `/api/auth/me` | GET | 获取当前用户 | 是 |
| `/api/auth/change-password` | POST | 修改密码 | 是 |
| `/api/auth/logout-all` | POST | 登出所有设备 | 是 |
| `/api/portfolio/dashboard` | GET | 获取仪表盘数据 | 是 |
| `/api/portfolio/trades` | GET | 获取交易列表 | 是 |
| `/api/portfolio/trades` | POST | 创建交易 | 是 |
| `/api/portfolio/trades/:id` | DELETE | 删除交易 | 是 |
| `/api/portfolio/trades` | DELETE | 清空所有交易 | 是 |
| `/api/portfolio/export` | GET | 导出数据 | 是 |
| `/api/portfolio/import/preview` | POST | 导入预览 | 是 |
| `/api/portfolio/import/confirm` | POST | 确认导入 | 是 |

## 数据库迁移

首次部署后需要创建数据库表。可以在本地运行迁移：

```go
package main

import (
    "api/db"
    "api/models"
)

func main() {
    database := db.GetDB()
    database.AutoMigrate(
        &models.User{},
        &models.RefreshToken{},
        &models.Trade{},
        &models.Holding{},
        &models.ExchangeRate{},
    )
}
```

或者创建一个临时的 `/api/migrate.go` 端点来执行迁移。

## 注意事项

1. **冷启动**: Vercel Serverless Functions 有冷启动时间，首次访问可能较慢
2. **数据库连接**: 使用连接池复用数据库连接，避免每次请求都新建连接
3. **执行时间限制**: 免费版函数执行时间限制为 10 秒
4. **MySQL 连接**: 确保你的 MySQL 数据库允许从 Vercel 的 IP 地址访问

## 故障排查

### 查看日志
```bash
vercel logs --all
```

### 常见问题

1. **数据库连接失败**: 
   - 检查环境变量是否正确设置
   - 确认 MySQL 数据库允许外部访问
   - 检查防火墙设置

2. **JWT 验证失败**: 
   - 确保 JWT_SECRET 已正确设置
   - 检查令牌是否过期

3. **CORS 错误**: 
   - 如果前端和后端域名不同，需要在响应头中添加 CORS 配置
