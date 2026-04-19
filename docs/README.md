# FX Portfolio 文档中心

本文档中心包含 FX Portfolio 项目的所有技术文档。

## 文档导航

### 📊 业务逻辑
- [calculations.md](./calculations.md) - 财务指标计算说明
  - 成本计算（加权平均法）
  - 盈亏指标（浮动/实现/总盈亏）
  - 收益率计算
  - 数据模型说明

### ⚙️ 配置指南
- [configuration.md](./configuration.md) - 配置文件说明
  - 配置项详解
  - 环境变量
  - 配置优先级
  - 部署建议

### 🔧 开发指南
- [architecture.md](./architecture.md) - 系统架构设计
  - 技术栈
  - 目录结构
  - 数据流
  - 模块依赖

### 📚 API 文档
- [api.md](./api.md) - 接口文档
  - 认证接口
  - 资产组合接口
  - 价格接口
  - 数据导出/导入接口

---

## 快速开始

```bash
# 1. 克隆项目
git clone <repository>

# 2. 复制配置文件
cp config.yaml.example config.yaml

# 3. 编辑配置
vim config.yaml

# 4. 运行
go run main.go
```

## 环境要求

- Go 1.22+
- MySQL 8.0+
- Node.js 20+ (前端开发)

---

*文档版本: 1.0*
*最后更新: 2026年4月19日*
