# API 文档

本文档详细说明 FX Portfolio 的后端 API 接口。

## 目录

- [认证接口](#认证接口)
- [资产组合接口](#资产组合接口)
- [价格接口](#价格接口)
- [数据导出/导入接口](#数据导出导入接口)

---

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Token (Bearer)
- **Content-Type**: `application/json`

---

## 认证接口

### 1. 用户注册

```http
POST /api/auth/register
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码 |

**响应示例:**

```json
{
  "message": "注册成功",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

### 2. 用户登录

```http
POST /api/auth/login
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码 |

**响应示例:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

### 3. 刷新 Token

```http
POST /api/auth/refresh
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | 刷新令牌 |

### 4. 退出登录

```http
POST /api/auth/logout
```

**请求头:**

```
Authorization: Bearer <token>
```

---

## 资产组合接口

### 1. 获取仪表盘数据

```http
GET /api/portfolio/dashboard
```

**请求头:**

```
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "prices": {
    "BTC": 45000.00,
    "ETH": 3000.00
  },
  "price_changes": {
    "BTC": 2.5,
    "ETH": -1.2
  },
  "portfolio": [
    {
      "symbol": "BTC",
      "amount": 0.5,
      "current_price": 45000.00,
      "avg_cost": 40000.00,
      "market_value": 22500.00,
      "cost": 20000.00,
      "profit_loss": 2500.00,
      "pl_rate": 12.5
    }
  ],
  "crypto_value": 22500.00,
  "usdt_balance": 10000.00,
  "unrealized_profit_loss": 2500.00,
  "unrealized_profit_loss_rate": 12.5,
  "realized_profit_loss": 1000.00,
  "realized_profit_loss_rate": 5.0
}
```

### 2. 获取交易记录

```http
GET /api/portfolio/trades
```

**请求头:**

```
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "trades": [
    {
      "id": 1,
      "symbol": "BTC",
      "type": "buy",
      "amount": 0.5,
      "price": 40000.00,
      "total": 20000.00,
      "created_at": "2024-01-15 10:30:00"
    }
  ]
}
```

### 3. 创建交易

```http
POST /api/portfolio/trades
```

**请求头:**

```
Authorization: Bearer <token>
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 是 | 币种代码 (BTC, ETH等) |
| type | string | 是 | 交易类型: buy, sell, recharge |
| amount | number | 是 | 数量 |
| price | number | 是 | 单价 |

**响应示例:**

```json
{
  "message": "交易创建成功",
  "trade": {
    "id": 1,
    "symbol": "BTC",
    "type": "buy",
    "amount": 0.5,
    "price": 40000.00,
    "total": 20000.00
  }
}
```

### 4. 删除交易

```http
DELETE /api/portfolio/trades/:id
```

**请求头:**

```
Authorization: Bearer <token>
```

**路径参数:**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | number | 交易ID |

### 5. 清空交易记录

```http
DELETE /api/portfolio/trades
```

**请求头:**

```
Authorization: Bearer <token>
```

---

## 价格接口

### 1. 获取单个资产价格

```http
GET /api/prices/:symbol
```

**路径参数:**

| 字段 | 类型 | 说明 |
|------|------|------|
| symbol | string | 币种代码 |

**响应示例:**

```json
{
  "symbol": "BTC",
  "price": 45000.00,
  "updated_at": 1705312200000
}
```

---

## 数据导出/导入接口

### 1. 导出数据

```http
GET /api/portfolio/export
```

**请求头:**

```
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "success": true,
  "data": {
    "version": "1.0",
    "export_time": "2024-01-15T10:30:00Z",
    "app_name": "fx-portfolio",
    "trades": [
      {
        "symbol": "BTC",
        "type": "buy",
        "amount": 0.5,
        "price": 40000.00,
        "total": 20000.00,
        "created_at": "2024-01-10T08:00:00Z"
      }
    ]
  }
}
```

### 2. 导入预览

```http
POST /api/portfolio/import/preview
```

**请求头:**

```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| data | object | 是 | 导入数据 |

**响应示例:**

```json
{
  "success": true,
  "preview": {
    "total_trades": 10,
    "new_trades": 8,
    "conflicts": 2,
    "conflict_items": [
      {
        "trade": { ... },
        "reason": "交易记录已存在"
      }
    ]
  }
}
```

### 3. 确认导入

```http
POST /api/portfolio/import/confirm
```

**请求头:**

```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求参数:**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| data | object | 是 | 导入数据 |
| conflict_strategy | string | 是 | 冲突策略: skip, overwrite |

**响应示例:**

```json
{
  "success": true,
  "imported": 8,
  "skipped": 2,
  "overwritten": 0
}
```

---

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权，Token 无效或过期 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 错误响应格式

```json
{
  "error": "错误信息"
}
```

---

*文档版本: 1.0*
*最后更新: 2026年4月19日*
