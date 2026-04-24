# zai-proxy

zai-proxy 是一个基于 Go 语言的代理服务，将 z.ai 网页聊天转换为 **OpenAI API** 和 **Anthropic Messages API** 兼容格式。用户使用自己的 z.ai token 或 `free` 匿名令牌进行调用。

## 功能特性

- OpenAI `/v1/chat/completions` 兼容
- Anthropic `/v1/messages` 兼容（支持 Claude SDK 直连）
- 支持流式 (SSE) 和非流式响应
- 支持多种 GLM 模型，Claude 模型名自动映射
- 支持思考模式 (thinking)
- 支持联网搜索模式 (search)
- 支持工具/函数调用 (function calling)
- 支持多模态图片输入
- 支持匿名 Token（免登录）
- 自动生成请求签名
- 自动跟踪前端版本号

## 快速开始

### 从源码运行

要求 Go 1.21+：

```bash
git clone https://github.com/yurika0211/zai-proxy.git
cd zai-proxy
go mod download
go run main.go
```

编译为可执行文件：

```bash
go build -o zai-proxy .
./zai-proxy
```

### Docker 部署

#### 使用预构建镜像

```bash
docker run -d --name zai-proxy -p 8000:8000 ghcr.io/yurika0211/zai-proxy:latest
```

自定义端口和日志级别：

```bash
docker run -d --name zai-proxy \
  -p 8080:8000 \
  -e LOG_LEVEL=debug \
  ghcr.io/yurika0211/zai-proxy:latest
```

> 镜像仅支持 `linux/amd64` 平台。

#### 挂载代理文件（可选）

```bash
docker run -d --name zai-proxy \
  -p 8000:8000 \
  -v ./proxies.txt:/app/proxies.txt \
  ghcr.io/yurika0211/zai-proxy:latest
```

#### 可用镜像标签

| 标签 | 说明 |
|------|------|
| `latest` | 默认分支的最新构建 |
| `v*`（如 `v1.0.0`） | 语义化版本发布 |
| `<commit-sha>` | 特定 commit 的构建 |

#### 自行构建镜像

```bash
docker build -t zai-proxy .
docker run -d --name zai-proxy -p 8000:8000 zai-proxy
```

#### Docker Compose

```yaml
services:
  zai-proxy:
    image: ghcr.io/yurika0211/zai-proxy:latest
    ports:
      - "8000:8000"
    environment:
      - LOG_LEVEL=info
    restart: unless-stopped
```

```bash
docker compose up -d
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 监听端口 | `8000` |
| `LOG_LEVEL` | 日志级别 (`debug` / `info` / `warn` / `error`) | `info` |

支持 `.env` 文件自动加载。

## 代理池（可选）

在项目根目录放置 `proxies.txt` 文件，每行一个 SOCKS5 代理，格式：

```
ip:port:username:password
ip:port
```

启动时自动加载，每次请求随机选择一个代理。不存在该文件则直连。

## 获取 z.ai Token

### 方式一：使用匿名 Token（免登录）

直接使用 `free` 作为 API key，服务会自动获取匿名 token：

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer free" \
  -H "Content-Type: application/json" \
  -d '{"model": "glm-4.7", "messages": [{"role": "user", "content": "hello"}]}'
```

### 方式二：使用个人 Token

1. 登录 https://chat.z.ai
2. 打开浏览器开发者工具 (F12)
3. 切换到 Application/Storage 标签
4. 在 Cookies 中找到 `token` 字段
5. 复制其值作为 API 调用的 Authorization

## API 端点

### `GET /v1/models`

返回可用模型列表（OpenAI 兼容格式）。

### `POST /v1/chat/completions`

OpenAI 兼容的聊天补全接口。

**认证方式：** `Authorization: Bearer <token>` 或 `x-api-key: <token>`

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_ZAI_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-4.7",
    "messages": [{"role": "user", "content": "hello"}],
    "stream": true
  }'
```

### `POST /v1/messages`

Anthropic Messages API 兼容接口，可直接使用 Anthropic SDK 连接。

**认证方式：** `x-api-key: <token>` 或 `Authorization: Bearer <token>`

```bash
curl http://localhost:8000/v1/messages \
  -H "x-api-key: free" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "hello"}],
    "stream": true
  }'
```

支持的 Anthropic 特性：
- `system` 字段（字符串或内容块数组）
- `thinking` 推理模式（`{"type": "enabled"}`）
- `tools` / `tool_choice` 工具调用
- 流式 SSE 事件（`message_start` / `content_block_delta` / `message_stop` 等）
- 非流式 JSON 响应

## 支持的模型

### GLM 模型

| 模型名称 | 上游模型 | 说明 |
|----------|----------|------|
| `glm-4.5` | 0727-360B-API | |
| `glm-4.6` | GLM-4-6-API-V1 | |
| `glm-4.7` | glm-4.7 | |
| `glm-5` | glm-5 | |
| `glm-5.1` | GLM-5.1 | 最新 |
| `glm-4.5-v` | glm-4.5v | 视觉 |
| `glm-4.6-v` | glm-4.6v | 视觉（最新） |
| `glm-4.5-air` | 0727-106B-API | 轻量 |

### Claude 模型名映射

通过 `/v1/messages` 端点使用 Claude 模型名时，会自动映射到对应的 GLM 模型：

| Claude 模型 | 映射到 | 备注 |
|-------------|--------|------|
| `claude-opus-4-6` | glm-5.1 | 自动启用 thinking |
| `claude-opus-4-5-20250514` | glm-5.1 | 自动启用 thinking |
| `claude-sonnet-4-6` | glm-4.7 | |
| `claude-sonnet-4-5-20241022` | glm-4.7 | |
| `claude-haiku-4-5` | glm-4.5-air | |
| `claude-haiku-4-5-20251001` | glm-4.5-air | |
| `claude-3-5-sonnet-20241022` | glm-4.7 | |
| `claude-3-5-haiku-20241022` | glm-4.5-air | |

> Opus 系列模型始终自动启用 thinking 模式。未识别的模型名会回退到 glm-4.7。

### 模型标签

模型名称支持以下后缀标签（可组合使用，顺序不限）：

| 标签 | 说明 |
|------|------|
| `-thinking` | 启用思考模式，响应包含 `reasoning_content` 字段 |
| `-search` | 启用联网搜索，搜索结果自动转换为 markdown 引用 |
| `-tools` | 自动注入内置工具定义，模型可进行函数调用 |

组合示例：
- `glm-4.7-thinking`
- `glm-4.7-search`
- `glm-4.7-thinking-search`
- `glm-4.7-tools`
- `glm-4.7-tools-thinking`
- `glm-5-thinking`
- `glm-5-tools`

## 使用示例

### 多模态请求

```json
{
  "model": "glm-4.6-v",
  "messages": [
    {
      "role": "user",
      "content": [
        {"type": "text", "text": "描述这张图片"},
        {"type": "image_url", "image_url": {"url": "https://example.com/image.jpg"}}
      ]
    }
  ]
}
```

支持的图片格式：
- HTTP/HTTPS URL
- Base64 编码 (`data:image/jpeg;base64,...`)

### Anthropic SDK (Python)

```python
import anthropic

client = anthropic.Anthropic(
    api_key="free",
    base_url="http://localhost:8000",
)

message = client.messages.create(
    model="claude-sonnet-4-6",
    max_tokens=1024,
    messages=[{"role": "user", "content": "hello"}],
)
print(message.content[0].text)
```

### Anthropic SDK 思考模式

```python
message = client.messages.create(
    model="claude-opus-4-6",
    max_tokens=8192,
    thinking={"type": "enabled", "budget_tokens": 4096},
    messages=[{"role": "user", "content": "解释量子纠缠"}],
)

for block in message.content:
    if block.type == "thinking":
        print("思考:", block.thinking)
    elif block.type == "text":
        print("回答:", block.text)
```

## 工具调用 (Function Calling)

### 内置工具

使用 `-tools` 后缀时，代理会自动注入以下 6 个内置工具定义：

| 工具名 | 描述 | 主要参数 |
|--------|------|----------|
| `get_current_time` | 获取当前时间 | `timezone`, `format` |
| `calculate` | 执行数学计算 | `expression` |
| `search_web` | 搜索网络信息 | `query`, `num_results` |
| `query_database` | 执行 SQL 查询 | `sql`, `database` |
| `file_operations` | 文件读写列表 | `operation`, `path`, `content` |
| `call_external_api` | 调用外部 API | `url`, `method`, `headers`, `body` |

### 基本调用

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_ZAI_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-4.7-tools",
    "messages": [{"role": "user", "content": "现在几点了？"}],
    "stream": true
  }'
```

模型会返回 `tool_calls`（`finish_reason` 为 `"tool_calls"`），由客户端自行执行工具并将结果发回。

### 多轮调用流程

```
第1轮：用户提问 → 模型返回 tool_calls
第2轮：发送工具执行结果 → 模型生成最终回答
```

```bash
curl http://localhost:8000/v1/chat/completions \
  -H "Authorization: Bearer YOUR_ZAI_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-4.7-tools",
    "messages": [
      {"role": "user", "content": "现在几点了？"},
      {"role": "assistant", "content": "", "tool_calls": [
        {"id": "call_xxx", "type": "function", "function": {"name": "get_current_time", "arguments": "{}"}}
      ]},
      {"role": "tool", "tool_call_id": "call_xxx", "content": "{\"time\": \"2026-03-14 15:30:00\"}"}
    ],
    "stream": true
  }'
```

### 自定义工具

也可以不使用 `-tools` 后缀，直接在请求中传入 `tools` 字段（标准 OpenAI 格式）：

```json
{
  "model": "glm-4.7",
  "messages": [{"role": "user", "content": "北京天气怎么样？"}],
  "tools": [{
    "type": "function",
    "function": {
      "name": "get_weather",
      "description": "获取天气信息",
      "parameters": {
        "type": "object",
        "properties": {
          "city": {"type": "string", "description": "城市名称"}
        },
        "required": ["city"]
      }
    }
  }],
  "tool_choice": "auto"
}
```

`tool_choice` 支持：
- `"auto"` — 模型自行决定是否调用工具
- `"required"` — 强制调用工具
- `"none"` — 禁用工具调用
- `{"type": "function", "function": {"name": "xxx"}}` — 强制调用指定工具

两者可混合使用：`-tools` 模型名 + 自定义 `tools` 字段。**客户端自带的同名工具优先**，不会被内置工具覆盖。

### Anthropic 格式工具调用

通过 `/v1/messages` 端点同样支持工具调用：

```bash
curl http://localhost:8000/v1/messages \
  -H "x-api-key: free" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "北京天气怎么样？"}],
    "tools": [{
      "name": "get_weather",
      "description": "获取天气信息",
      "input_schema": {
        "type": "object",
        "properties": {
          "city": {"type": "string", "description": "城市名称"}
        },
        "required": ["city"]
      }
    }]
  }'
```

`tool_choice` 映射关系：

| Anthropic | OpenAI 等价 |
|-----------|-------------|
| `{"type": "auto"}` | `"auto"` |
| `{"type": "any"}` | `"required"` |
| `{"type": "none"}` | `"none"` |
| `{"type": "tool", "name": "xxx"}` | `{"type": "function", ...}` |

## 项目结构

```
zai-proxy/
├── main.go                     # 入口，路由注册
├── internal/
│   ├── auth/                   # 认证与签名
│   │   ├── anonymous.go        # 匿名 token 获取
│   │   ├── jwt.go              # JWT 解码
│   │   └── signature.go        # HMAC-SHA256 请求签名
│   ├── config/                 # 配置加载 (.env)
│   ├── filter/                 # 响应处理过滤器
│   │   ├── prompttool.go       # 提取 prompt 注入的 tool call
│   │   ├── toolcall.go         # 解析 JSON 格式的 tool call
│   │   ├── thinking.go         # 思考/推理内容处理
│   │   └── search.go           # 搜索结果引用转换
│   ├── handler/                # HTTP 处理器
│   │   ├── chat.go             # /v1/chat/completions
│   │   ├── anthropic.go        # /v1/messages
│   │   └── models.go           # /v1/models
│   ├── logger/                 # 日志系统
│   ├── model/                  # 类型定义
│   │   ├── types.go            # OpenAI 兼容类型
│   │   ├── anthropic.go        # Anthropic API 类型
│   │   └── mapping.go          # 模型名映射与解析
│   ├── proxy/                  # SOCKS5 代理池
│   │   └── pool.go             # 代理加载与随机选择
│   ├── tools/                  # 内置工具定义
│   │   ├── builtin.go          # 6 个内置工具
│   │   └── prompt.go           # 工具系统提示词构建
│   ├── upstream/               # 上游 z.ai 客户端
│   │   ├── client.go           # 请求构建与发送
│   │   └── upload.go           # 图片上传
│   └── version/                # 前端版本号跟踪
├── scripts/
│   └── test_tool_call.sh       # 工具调用集成测试
├── Dockerfile
└── .github/workflows/          # CI/CD
```
