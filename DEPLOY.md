# SamaraAI 从零部署指南

本文档面向**完全没有部署经验**的同学，手把手把 SamaraAI 跑起来。  
默认场景是：**本机开发调试**（Windows / macOS / Linux 均可），文末附带**生产环境**简要说明。

---

## 一、这个项目需要什么？

SamaraAI 是一个全栈 AI 应用，由 **Go 后端** + **Vue 前端** 组成，并依赖以下中间件：

| 组件 | 用途 | 是否必须 |
|------|------|----------|
| **Go 1.24+** | 运行后端 | 必须 |
| **Node.js 18+** | 运行 / 构建前端 | 必须 |
| **MySQL 8.0+** | 存储用户、会话、消息 | 必须 |
| **Redis**（需带 **RediSearch** 模块） | 验证码、RAG 向量检索 | 必须 |
| **RabbitMQ** | 异步消息队列 | 必须 |
| **大模型 API Key**（阿里百炼等） | AI 对话、RAG 向量化 | 必须 |
| **QQ 邮箱授权码** | 注册验证码邮件 | 注册功能需要 |
| **ONNX 模型文件** | 图像识别 | 可选 |

整体架构：

```
浏览器 (Vue 前端 :8080)
        │
        ├─ 普通 API ──► Go 后端 (:9090) ──► MySQL / Redis Stack / RabbitMQ
        │
        └─ SSE 流式 ──► Go 后端 (:9090) 直连（开发环境跨域 + CORS）
                              │
                              ├── MCP 子服务 (:8081/mcp，可选)
                              └── 外部 LLM API（阿里百炼等）
```

**聊天三种模式**（前端下拉选择，对应后端 `modelType`）：

| modelType | 前端名称 | 说明 |
|-----------|----------|------|
| `1` | DeepSeek 兼容 · 百炼 | 普通 LLM 对话 |
| `2` | RAG 知识库 | 检索已上传文档后回答（需 Redis Stack） |
| `3` | MCP 工具 | Function Calling 调用外部工具（如天气） |

> 上传 RAG 文档 **不会** 自动切换模式；文档与 MCP/普通对话互不干扰，要用文档请手动选「RAG 知识库」或点「用 RAG 提问」。

---

## 二、安装基础软件

### 2.1 安装 Go

1. 打开 [https://go.dev/dl/](https://go.dev/dl/) 下载 **Go 1.24** 或更高版本。
2. 安装后打开终端，验证：

```bash
go version
# 应输出 go version go1.24.x ...
```

### 2.2 安装 Node.js

1. 打开 [https://nodejs.org/](https://nodejs.org/) 下载 **LTS 版本**（建议 18 或 20）。
2. 验证：

```bash
node -v
npm -v
```

### 2.3 安装 Git

1. 打开 [https://git-scm.com/downloads](https://git-scm.com/downloads) 下载并安装。
2. 验证：

```bash
git --version
```

### 2.4（可选）图像识别需要的 C 编译器

图像识别使用 ONNX Runtime，在 **Windows** 上需要 **gcc**（用于 cgo）：

- 安装 [MSYS2](https://www.msys2.org/)，然后在 MSYS2 终端执行：`pacman -S mingw-w64-x86_64-gcc`
- 把 `C:\msys64\mingw64\bin` 加入系统 PATH

如果暂时不用图像识别，可跳过此步。

---

## 三、安装中间件

下面提供 **Docker 一键安装**（推荐新手）和 **手动安装** 两种方式，**二选一**即可。

### 方式 A：Docker 一键安装（推荐）

先安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)。

在项目根目录新建 `docker-compose.yml`（或直接在终端逐条 `docker run`）：

```yaml
services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: "123456"
      MYSQL_DATABASE: SamaraAI
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis/redis-stack:latest
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: root
      RABBITMQ_DEFAULT_PASS: "123456"

volumes:
  mysql_data:
```

启动：

```bash
docker compose up -d
```

> **注意**：RAG 功能需要 **Redis Stack**（自带 RediSearch 向量搜索），普通 `redis:latest` 或 Windows 自带 Redis **不支持** `FT.CREATE` / `FT.SEARCH`。

**Windows 注意**：若本机已安装 Redis 服务占用了 `6379`，需先停止，再启动 Docker 容器：

```powershell
# 查看占用
netstat -ano | findstr :6379

# 若安装了 Windows Redis 服务（管理员 CMD）
net stop Redis

# 启动 Redis Stack（推荐单独容器名）
docker run -d --name samara-redis -p 6379:6379 redis/redis-stack:latest

# 验证 RediSearch
docker exec -it samara-redis redis-cli FT.INFO test
# 应返回 Unknown index name（说明 FT 命令可用），而非 unknown command
```

### 方式 B：手动安装

#### MySQL

1. 下载安装 [MySQL Community Server 8.0](https://dev.mysql.com/downloads/mysql/)。
2. 记住 root 密码。
3. 创建数据库：

```sql
CREATE DATABASE SamaraAI CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

> 表结构无需手动建，后端启动时会通过 GORM **自动建表**。

#### Redis（Redis Stack）

- Windows：用 Docker 跑 `redis/redis-stack:latest` 最省事。
- Linux：`sudo apt install redis-stack-server` 或 Docker。
- macOS：`brew install redis-stack`

验证 RediSearch 是否可用：

```bash
redis-cli FT._LIST
# 不报错即表示模块已加载
```

#### RabbitMQ

- Windows：推荐 [Docker](https://hub.docker.com/_/rabbitmq) 或安装 Erlang + RabbitMQ。
- Linux：`sudo apt install rabbitmq-server`
- macOS：`brew install rabbitmq`

默认管理界面：`http://localhost:15672`（用户名/密码见 docker-compose 配置）。

---

## 四、获取项目代码

```bash
git clone <你的仓库地址>
cd SamaraAI-v2
```

---

## 五、配置后端 `etc/samara.yaml`

### 5.1 复制配置文件

```bash
cp etc/samara.yaml.example etc/samara.yaml
```

> `etc/samara.yaml` 已在 `.gitignore` 中，**不会提交到 Git**，请放心填写密钥。

### 5.2 逐项填写

用任意文本编辑器打开 `etc/samara.yaml`，重点修改以下字段：

```yaml
# 开发环境用 dev，生产环境改为 prod
Mode: dev

Mysql:
  Host: 127.0.0.1
  Port: 3306
  User: root
  Password: "你的MySQL密码"
  DatabaseName: SamaraAI

Redis:
  Host: 127.0.0.1
  Port: 6379
  Password: ""
  Db: 0

Rabbitmq:
  Host: localhost
  Port: 5672
  Username: root
  Password: "123456"
  Vhost: /

# QQ 邮箱（用于发送注册验证码，固定走 smtp.qq.com:587）
Email:
  Email: your-email@qq.com
  Authcode: 你的QQ邮箱SMTP授权码

# 阿里百炼（通义千问）— 对话和 RAG 默认使用
OpenAI:
  ApiKey: sk-xxxxxxxx
  Model: qwen-plus
  BaseUrl: https://dashscope.aliyuncs.com/compatible-mode/v1

Rag:
  EmbeddingModel: text-embedding-v4
  ChatModelName: qwen-turbo
  DocDir: ./docs
  UploadDir: ./uploads          # 用户文档实际目录：uploads/{用户名}/
  ChunkSize: 800
  ChunkOverlap: 100
  MaxFilesPerUser: 5
  BaseUrl: https://dashscope.aliyuncs.com/compatible-mode/v1
  Dimension: 1024

Mcp:
  Enabled: true
  Addr: ":8081"
  BaseURL: http://localhost:8081/mcp
  ChatModelName: qwen-plus      # 建议使用支持工具调用的模型

Image:
  ModelPath: ./models/mobilenetv2/mobilenetv2-7.onnx
  LabelPath: ./imagenet_classes.txt
  InputH: 224
  InputW: 224
```

### 5.3 获取各项密钥

| 配置项 | 获取方式 |
|--------|----------|
| **阿里百炼 ApiKey** | 登录 [阿里云百炼控制台](https://bailian.console.aliyun.com/) → API-KEY 管理 → 创建 |
| **QQ 邮箱授权码** | QQ 邮箱 → 设置 → 账户 → POP3/SMTP → 开启服务 → 生成授权码 |
| **DeepSeek ApiKey** | [https://platform.deepseek.com/](https://platform.deepseek.com/) 注册获取（可选） |
| **百度 ApiKey** | [百度智能云](https://cloud.baidu.com/) 创建应用（可选） |

### 5.4 Mode 说明

| 值 | 含义 |
|----|------|
| `dev` | 开发模式：MySQL 打印 SQL 日志；JWT 支持 URL 参数 `?token=xxx` |
| `prod` | 生产模式：关闭上述调试行为，**部署到服务器请改为 prod** |

---

## 六、准备可选资源

### 6.1 RAG 知识库文档

RAG 文档通过**前端聊天页上传**（支持 `.txt` / `.md`），后端会：

1. 保存到 `uploads/{用户名}/`（可在 `etc/samara.yaml` 的 `Rag.UploadDir` 配置）
2. 按 **800 字切块**（可配置 `ChunkSize` / `ChunkOverlap`）并向量化
3. 写入 **Redis Stack** 向量索引（每份文件独立索引，检索时合并 Top-K）

默认每用户最多保留 **5 份**文档（`Rag.MaxFilesPerUser`），超出会自动删除最早上传的文件。

**前置条件**：Redis 必须使用 **redis-stack**（带 RediSearch），普通 Redis 不支持向量索引。

**使用步骤**：

1. 在 AI 聊天页点击 **📎** 上传 `.md` / `.txt`（上传后顶部会出现「知识库文档」栏，**不切换当前对话模式**）
2. 需要基于文档问答时：下拉选 **「RAG 知识库」**，或点击 **「用 RAG 提问」**
3. 针对文档内容提问（如「总结文档要点」）

**说明**：

- 知识库按 **用户** 隔离，保存在 `uploads/{用户名}/`，与具体聊天会话无关
- 每用户最多 **5** 份文档，超出自动删最早上传的
- 可在文档 Chip 上点 **×** 删除文档及向量索引
- 若更换 Redis 实例（如从 Windows Redis 换到 Docker Redis Stack），本地文件仍在，**首次 RAG 检索会自动重建索引**；也可删除后重新上传

**常见 RAG 报错**：

| 现象 | 原因 | 处理 |
|------|------|------|
| 上传报 `FT.INFO` / `FT.CREATE` | Redis 非 Stack | 换 `redis/redis-stack:latest` |
| 检索失败提示 Redis Stack | 索引丢失或 Redis 刚切换 | 重启后端后重试；或重新上传 |
| AI 说无法读文件 | 未选 RAG 模式 | 切换到「RAG 知识库」 |

### 6.2 图像识别模型（可选）

如需使用图像识别，先下载模型与标签文件。

#### 方式 A：一键脚本（推荐）

**Windows CMD**（推荐，不受 PowerShell 执行策略限制）：

```cmd
cd /d 项目根目录
scripts\download-models.cmd
```

仅重装 ONNX Runtime 动态库：

```cmd
scripts\download-onnxruntime.cmd
```

**Windows PowerShell**（若报「禁止运行脚本」，请改用上方 `.cmd`）：

```powershell
.\scripts\download-models.ps1
```

**Linux / macOS / Git Bash**：

```bash
bash scripts/download-models.sh
```

#### 方式 B：手动下载

| 文件 | 大小 | 官方来源 | 直接下载链接 |
|------|------|----------|--------------|
| `mobilenetv2-7.onnx` | ~14 MB | [Hugging Face ONNX Model Zoo](https://huggingface.co/onnxmodelzoo/mobilenetv2-7) | https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx |
| `imagenet_classes.txt` | ~30 KB | [PyTorch Hub](https://github.com/pytorch/hub/blob/master/imagenet_classes.txt) | https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt |

**PowerShell 手动命令**：

```powershell
mkdir models\mobilenetv2 -Force
Invoke-WebRequest -Uri "https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx" -OutFile "models\mobilenetv2\mobilenetv2-7.onnx"
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt" -OutFile "imagenet_classes.txt"
```

**Linux / macOS 手动命令**：

```bash
mkdir -p models/mobilenetv2
curl -L -o models/mobilenetv2/mobilenetv2-7.onnx \
  https://huggingface.co/onnxmodelzoo/mobilenetv2-7/resolve/main/mobilenetv2-7.onnx
curl -L -o imagenet_classes.txt \
  https://raw.githubusercontent.com/pytorch/hub/master/imagenet_classes.txt
```

> GitHub 上的 [ONNX Model Zoo 原仓库](https://github.com/onnx/models/tree/main/validated/vision/classification/mobilenet/model) 使用 Git LFS，直接 `wget` 原始链接会得到 LFS 指针而非模型文件，建议使用上方 Hugging Face 链接。

下载完成后目录结构：

```
项目根目录/
├── models/mobilenetv2/mobilenetv2-7.onnx
├── imagenet_classes.txt
└── onnxruntime.dll          # Windows，与 samara.exe 同目录
    # 或 lib/libonnxruntime.so.1.22.0   # Linux
```

> **关键**：仅有 `.onnx` 模型不够，还必须下载 **ONNX Runtime 动态库**（`onnxruntime.dll`，版本 **1.22.0**）。若 DLL 损坏或版本不对，启动会报 `image recognizer init failed`。

**仅重装 DLL（推荐）**：

```powershell
cmd /c scripts\download-onnxruntime.cmd
```

或一键下载模型 + 标签 + DLL：

```powershell
cmd /c scripts\download-models.cmd
```

在 `etc/samara.yaml` 中确认路径：

```yaml
Image:
  ModelPath: ./models/mobilenetv2/mobilenetv2-7.onnx
  LabelPath: ./imagenet_classes.txt
  # RuntimeLibPath: ./onnxruntime.dll   # 可选，留空则自动查找
  InputH: 224
  InputW: 224
```

### 6.3 MCP 工具服务（可选）

MCP 提供外部工具调用能力，当前内置：

| 工具 | 用途 |
|------|------|
| `get_weather` | 查询城市实时天气（数据源 [wttr.in](https://wttr.in)） |

默认已随主程序自动启动，在 `etc/samara.yaml` 中配置：

```yaml
Mcp:
  Enabled: true          # 设为 false 可关闭
  Addr: ":8081"          # MCP HTTP 监听地址
  BaseURL: http://localhost:8081/mcp
```

启动后端后日志应出现：

```
MCP server starting at :8081/mcp
HTTP MCP server listening on :8081/mcp
```

使用方式：AI 聊天页选择 **「MCP 工具」**，提问如「苏州今天天气怎么样」。

> 已上传 RAG 文档不影响 MCP 使用；上传文档 **不会** 自动切到 MCP 或 RAG，需手动选择模式。

MCP 使用百炼/OpenAI 兼容接口的 **原生 Function Calling**（`tools` 参数），并结合**完整多轮对话**理解上下文（例如先说「我在苏州」，再说「好的」，会自动查询苏州天气）。

`Mcp.ChatModelName` 建议使用支持工具调用的模型（默认 `qwen-plus`）。

也可单独调试 MCP 服务：

```bash
go run ./common/mcp -mode server -http-addr :8081
go run ./common/mcp -mode client -http-addr :8081 -city 北京
```

---

## 七、启动后端

在项目根目录执行：

```bash
go mod download
go build -o samara.exe .
.\samara.exe -f etc/samara.yaml
```

或使用 `go run`（开发调试）：

```bash
go run main.go -f etc/samara.yaml
```

**Windows 注意（图像识别 / CGO）**：若用户名含中文等非 ASCII 字符，`go run` / `go build` 可能报 `runtime/cgo ... can't create ... Temp` 错误。请先设置纯英文临时目录：

```powershell
$env:GOTMPDIR="D:\Gocode\tmp"
$env:GOCACHE="D:\Gocode\gocache"
New-Item -ItemType Directory -Force -Path $env:GOTMPDIR, $env:GOCACHE | Out-Null
go build -o samara.exe .
.\samara.exe -f etc/samara.yaml
```

若暂时不用图像识别，可跳过 gcc 安装；普通 AI 对话 / RAG / MCP 不依赖 CGO。

看到类似输出表示成功：

```
redis init success
rabbitmq init success
MCP server starting at :8081/mcp
AIHelperManager init success
image recognizer init success    ← 图像识别就绪（无此行则见下方排查）
Starting SamaraAI at 0.0.0.0:9090...
HTTP MCP server listening on :8081/mcp
```

后端 API 地址：`http://localhost:9090`

### 常见启动失败

| 报错 | 原因 | 解决 |
|------|------|------|
| `InitMysql error` | MySQL 连不上或库不存在 | 检查密码、端口、是否已 `CREATE DATABASE` |
| `RabbitMQ connection failed` | RabbitMQ 未启动 | 启动 RabbitMQ 或检查账号密码 |
| `redis init` 后 RAG 报错 `FT.CREATE` / `FT.INFO` | Redis 没有 RediSearch | 换用 **redis-stack**；Windows 停掉自带 Redis 释放 6379 |
| `Proxy error ECONNREFUSED`（前端） | 后端 `:9090` 未启动 | 先启动 `samara.exe`，再开前端 |
| `Failed to fetch`（流式聊天） | SSE 跨域或后端未启 | 确认 `VUE_APP_STREAM_BASE`、后端 CORS、9090 可访问 |
| `runtime/cgo ... can't create ... Temp` | Windows 用户名含中文 | 设置 `GOTMPDIR` / `GOCACHE` 到英文路径（见上文） |
| MCP 工具无响应 | MCP 服务未启动或端口被占用 | 确认 `Mcp.Enabled: true`，日志有 `8081/mcp` |
| RAG 检索失败 | Redis 换实例后索引丢失 | 重启后端触发自动 reindex，或重新上传文档 |
| 图像识别 `init failed` | 缺/错 `onnxruntime.dll` | 运行 `scripts\download-onnxruntime.cmd`，需 **1.22.0** |
| 图像识别 API 版本错误 | DLL 为 1.17 等错误版本 | 删项目根 `onnxruntime.dll` 后重装，勿用 System32 旧 DLL |
| 邮件发送失败 | QQ 授权码错误 | 重新生成授权码，不要用 QQ 密码 |

---

## 八、启动前端（开发模式）

```bash
cd vue-frontend
npm install
npm run serve
```

**Windows PowerShell** 若报「禁止运行脚本」，可任选其一：

```powershell
# 方式 1：临时改用 npm.cmd
npm.cmd run serve

# 方式 2：放宽当前用户执行策略（推荐，只需一次）
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

浏览器打开：**http://localhost:8080**

### 前端环境变量说明

开发环境已配置 `vue-frontend/.env.development`：

```
VUE_APP_STREAM_BASE=http://localhost:9090/api/v1
```

- 普通 API 请求走 `vue.config.js` 代理（`/api` → 后端 `9090`）
- **流式聊天（SSE）** 直连后端 `9090`，保持实时输出（不走代理，避免缓冲）
- 后端已为 SSE 接口注册 **OPTIONS 预检** 与 CORS，解决跨域 `Failed to fetch`
- 修改 `.env.development` 后需**重启** `npm run serve`

---

## 九、验证部署是否成功

按顺序自测：

1. **后端存活**：浏览器访问 `http://localhost:9090`（可能返回 404，说明服务在跑）
2. **注册**：打开 `http://localhost:8080` → 注册 → 邮箱收到验证码
3. **登录**：用注册的账号登录 → 进入应用中心
4. **AI 对话**：智能对话 → 新建会话 → 发送消息，应看到**流式输出**
5. **历史会话**：侧栏显示会话列表；悬停可 **删除会话**
6. **（可选）RAG**：📎 上传文档 → 选「RAG 知识库」→ 针对文档提问；文档栏可删除文件
7. **（可选）MCP**：选「MCP 工具」→ 问「上海今天天气怎么样」（与是否上传文档无关）
8. **（可选）图像识别**：先下载模型（见 6.2 节），上传图片，返回 ImageNet 分类结果

### 主要 API 一览（均需登录除注册/登录外）

| 功能 | 方法 | 路径 |
|------|------|------|
| 登录 / 注册 / 验证码 | POST | `/api/v1/user/login` 等 |
| 会话列表 | GET | `/api/v1/AI/chat/sessions` |
| 历史消息 | POST | `/api/v1/AI/chat/history` |
| 删除会话 | POST | `/api/v1/AI/chat/delete-session` |
| 流式新会话 | POST | `/api/v1/AI/chat/send-stream-new-session` |
| 流式续聊 | POST | `/api/v1/AI/chat/send-stream` |
| 文档列表 | GET | `/api/v1/file/list` |
| 上传文档 | POST | `/api/v1/file/upload` |
| 删除文档 | POST | `/api/v1/file/delete` |
| 图像识别 | POST | `/api/v1/image/recognize` |

路由源码：`internal/handler/routes.go`

---

## 十、生产环境部署（Linux 服务器）

### 10.1 编译后端

```bash
# 在开发机上交叉编译（以 Linux amd64 为例）
GOOS=linux GOARCH=amd64 go build -o samara-api main.go

# 上传到服务器，连同以下文件：
#   samara-api（二进制）
#   etc/samara.yaml（Mode 改为 prod）
#   uploads/（用户 RAG 文档，如有）
#   models/、imagenet_classes.txt、onnxruntime.dll（Linux 为 libonnxruntime.so.*，如有图像识别）
```

服务器上运行：

```bash
chmod +x samara-api
./samara-api -f etc/samara.yaml
```

建议用 **systemd** 托管（`/etc/systemd/system/samara.service`）：

```ini
[Unit]
Description=SamaraAI API
After=network.target mysql.service redis.service rabbitmq-server.service

[Service]
Type=simple
WorkingDirectory=/opt/samara
ExecStart=/opt/samara/samara-api -f /opt/samara/etc/samara.yaml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now samara
```

### 10.2 构建前端

```bash
cd vue-frontend

# 创建生产环境变量（与后端同域部署时）
echo "VUE_APP_STREAM_BASE=/api/v1" > .env.production

npm install
npm run build
```

构建产物在 `vue-frontend/dist/`，用 Nginx 托管。

### 10.3 Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    root /opt/samara/vue-frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 普通 API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:9090/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # SSE 流式接口（关闭缓冲）
    location /api/v1/AI/chat/ {
        proxy_pass http://127.0.0.1:9090/api/v1/AI/chat/;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding on;
    }
}
```

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 10.4 生产环境检查清单

- [ ] `etc/samara.yaml` 中 `Mode` 设为 `prod`
- [ ] MySQL / Redis Stack / RabbitMQ 均已启动且密码已修改
- [ ] 防火墙放行 80/443（后端 9090 建议只内网访问）
- [ ] 配置 HTTPS（推荐 Let's Encrypt + certbot）
- [ ] API Key、邮箱授权码等敏感信息不要提交到 Git

---

## 十一、端口一览

| 服务 | 默认端口 |
|------|----------|
| 后端 API | 9090 |
| 前端开发服务器 | 8080 |
| MCP HTTP 服务 | 8081 |
| MySQL | 3306 |
| Redis Stack | 6379 |
| RabbitMQ | 5672 |
| RabbitMQ 管理界面 | 15672 |

---

## 十二、功能使用速查

| 想做什么 | 操作 |
|----------|------|
| 普通聊天 | 模式选「DeepSeek 兼容 · 百炼」 |
| 总结上传的文档 | 📎 上传 → 选「RAG 知识库」或「用 RAG 提问」 |
| 查天气 | 模式选「MCP 工具」，问「xx 天气」 |
| 上传文档后继续查天气 | 上传后不切换模式，保持 MCP 即可 |
| 删对话 | 侧栏会话悬停 × |
| 删知识库文档 | 文档 Chip 上 × |
| 识图 | 应用中心 → 图像识别 → 选图 → 开始识别 |

---

## 十三、相关文档

- 项目简介与目录结构：见 [README.md](./README.md)
- 配置模板：见 [etc/samara.yaml.example](./etc/samara.yaml.example)
- 界面原型说明：见 [docs/PROTOTYPE.md](./docs/PROTOTYPE.md)（如有）
- API 路由定义：见 `internal/handler/routes.go`
- 下载脚本：`scripts/download-models.cmd`、`scripts/download-onnxruntime.cmd`

---

如有问题，欢迎在项目 Issue 中反馈。
