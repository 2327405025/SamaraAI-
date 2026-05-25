# SamaraAI

SamaraAI 是一个基于 Go（go-zero）与 Vue 3 的全栈 AI 应用平台，提供智能对话、图像识别、会话管理与 RAG 等能力。

## 功能

- AI 多轮对话（支持流式输出）
- 图像识别
- 用户注册 / 登录 / 验证码
- 会话与历史消息管理
- Redis、RabbitMQ、MySQL 集成
- RAG 文档检索增强

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24、go-zero、GORM |
| 前端 | Vue 3、Element Plus、Vue Router |
| 中间件 | MySQL、Redis、RabbitMQ |

## 快速开始

### 1. 配置

```bash
cp etc/samara.yaml.example etc/samara.yaml
```

编辑 `etc/samara.yaml`，填写数据库、Redis、邮件、大模型 API Key 等配置。开发环境可设置 `Mode: dev` 开启 MySQL 详细日志与 URL token 鉴权。

聊天仅支持 SSE 流式接口：`/api/v1/AI/chat/send-stream`、`/api/v1/AI/chat/send-stream-new-session`（见 `api/samara.api` 顶部注释）。

### 2. 启动后端

```bash
go run main.go -f etc/samara.yaml
```

### 3. 启动前端

```bash
cd vue-frontend
npm install
npm run serve
```

## 目录结构

```
├── api/              # API 定义
├── common/           # 公共模块（AI、Redis、MQ、RAG 等）
├── etc/              # go-zero 服务配置（samara.yaml）
├── dao/              # 数据访问
├── etc/              # go-zero 服务配置
├── internal/         # HTTP handler / logic / middleware
├── service/          # 业务服务层
├── vue-frontend/     # Vue 前端
└── main.go
```

## 许可证

MIT
