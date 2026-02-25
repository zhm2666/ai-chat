# 模型输出协议分析文档

## 一、当前协议分析

### 1.1 协议类型

本项目当前使用的是 **HTTP Chunked Transfer Encoding + 自定义 JSON 行分隔协议**，不是标准的 SSE（Server-Sent Events）。

### 1.2 协议工作原理

#### 后端实现（ai-chat-backend）

位置：`ai-chat-backend/pkg/controllers/chat.go`

```125:176:ai-chat-backend/pkg/controllers/chat.go
func (chat *ChatController) ChatProcess(ctx *gin.Context) {
    // ... 省略部分代码 ...
    
    // 设置 Content-Type 为 application/octet-stream（二进制流）
    ctx.Header("Content-type", "application/octet-stream")
    
    for {
        // 从 gRPC 流接收响应
        rsp, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            return
        }
        // ... 处理响应 ...
        
        // 序列化为 JSON
        bts, err := json.Marshal(result)
        
        // 使用换行符 \n 作为分隔符
        if !firstChunk {
            ctx.Writer.Write([]byte("\n"))
        }
        
        // 写入响应并立即刷新
        ctx.Writer.Write(bts)
        ctx.Writer.Flush()
    }
}
```

#### 前端解析实现（ai-chat-web）

位置：`ai-chat-web/src/views/chat/index.vue`

```116:126:ai-chat-web/src/views/chat/index.vue
onDownloadProgress: ({ event }) => {
  const xhr = event.target
  const { responseText } = xhr
  // 找到最后一个换行符的位置
  const lastIndex = responseText.lastIndexOf('\n', responseText.length - 2)
  let chunk = responseText
  if (lastIndex !== -1)
    chunk = responseText.substring(lastIndex)
  // 解析 JSON
  const data = JSON.parse(chunk)
  // 更新聊天消息
  updateChat(...)
}
```

### 1.3 当前协议特点

| 特性 | 说明 |
|------|------|
| **传输方式** | HTTP Chunked Transfer Encoding |
| **数据格式** | JSON 字符串 |
| **分隔符** | 换行符 `\n` |
| **Content-Type** | `application/octet-stream` |
| **无标准事件** | 非标准 SSE，无 `data:` 前缀 |
| **无事件名** | 无法区分不同类型的事件 |
| **实现复杂度** | 需要前端手动解析 `\n` 分隔的 JSON |

### 1.4 数据流示例

```
# 后端发送（每行一个 JSON 对象）
{"id":"chat-001","text":"","delta":"你好","detail":{...}}
{"id":"chat-001","text":"你好","delta":"！","detail":{...}}
{"id":"chat-001","text":"你好！","delta":"有什么","detail":{...}}
{"id":"chat-001","text":"你好！有什么","delta":"可以","detail":{...}}
{"id":"chat-001","text":"你好！有什么可以","delta":"帮助","detail":{...}}
{"id":"chat-001","text":"你好！有什么可以帮助","delta":"你的","detail":{...}}
{"id":"chat-001","text":"你好！有什么可以帮助你的","delta":"？","detail":{...}}
{"id":"chat-001","text":"你好！有什么可以帮助你的？","delta":"","detail":{...},"finish_reason":"stop"}
```

---

## 二、SSE 协议介绍

### 2.1 什么是 SSE

**Server-Sent Events (SSE)** 是 HTML5 引入的一种服务器推送技术，允许服务器通过 HTTP 协议向客户端推送数据。

### 2.2 SSE 特点

| 特性 | 说明 |
|------|------|
| **标准协议** | W3C 标准，自动重连 |
| **Content-Type** | `text/event-stream` |
| **数据格式** | `data: 消息内容\n\n` |
| **事件支持** | 可通过 `event:` 定义事件类型 |
| **单向通信** | 服务器 → 客户端 |
| **简单易用** | 原生 EventSource API 支持 |

### 2.3 SSE 数据格式

```sse
# 普通消息
data: {"message": "Hello"}

# 多行数据
data: {"line1": "Hello"}
data: {"line2": "World"}

# 命名事件
event: chat
data: {"message": "Hello"}

# 结束消息
event: done
data: {"status": "complete"}
```

---

## 三、改造方案

### 3.1 改造目标

将当前的 **自定义 JSON 行分隔协议** 改为标准的 **SSE 协议**。

### 3.2 改造内容

#### 3.2.1 后端改造（ai-chat-backend）

**文件位置**：`ai-chat-backend/pkg/controllers/chat.go`

**修改点 1：修改 Content-Type**

```go
// 修改前
ctx.Header "application/oct("Content-type",et-stream")

// 修改后
ctx.Header("Content-Type", "text/event-stream")
ctx.Header("Cache-Control", "no-cache")
ctx.Header("Connection", "keep-alive")
ctx.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲
```

**修改点 2：修改响应格式**

```go
// 修改前
if !firstChunk {
    ctx.Writer.Write([]byte("\n"))
}
ctx.Writer.Write(bts)
ctx.Writer.Flush()

// 修改后 - 使用 SSE 格式
// 发送文本内容
ctx.Writer.Write([]byte("data: "))
ctx.Writer.Write(bts)
ctx.Writer.Write([]byte("\n\n"))
ctx.Writer.Flush()
```

**完整代码示例**

```go
func (chat *ChatController) ChatProcessSSE(ctx *gin.Context) {
    // ... 省略：获取请求参数、创建 gRPC 客户端等 ...
    
    // 设置 SSE 响应头
    ctx.Header("Content-Type", "text/event-stream")
    ctx.Header("Cache-Control", "no-cache")
    ctx.Header("Connection", "keep-alive")
    ctx.Header("X-Accel-Buffering", "no")
    ctx.Header("Access-Control-Allow-Origin", "*")
    
    stream, err := aiChatServiceClient.ChatCompletionStream(ctx1, in)
    if err != nil {
        // 发送错误事件
        errorData, _ := json.Marshal(gin.H{
            "error": err.Error(),
        })
        ctx.Writer.Write([]byte("event: error\n"))
        ctx.Writer.Write([]byte("data: " + string(errorData) + "\n\n"))
        ctx.Writer.Flush()
        return
    }
    defer stream.CloseSend()
    
    // 发送开始事件（可选）
    startData, _ := json.Marshal(gin.H{
        "id":      in.Id,
        "created": time.Now().Unix(),
    })
    ctx.Writer.Write([]byte("event: start\n"))
    ctx.Writer.Write([]byte("data: " + string(startData) + "\n\n"))
    ctx.Writer.Flush()
    
    for {
        rsp, err := stream.Recv()
        if errors.Is(err, io.EOF) {
            // 发送结束事件
            ctx.Writer.Write([]byte("event: done\n"))
            ctx.Writer.Write([]byte("data: {\"status\":\"complete\"}\n\n"))
            ctx.Writer.Flush()
            return
        }
        if err != nil {
            // 发送错误事件
            errorData, _ := json.Marshal(gin.H{
                "error": err.Error(),
            })
            ctx.Writer.Write([]byte("event: error\n"))
            ctx.Writer.Write([]byte("data: " + string(errorData) + "\n\n"))
            ctx.Writer.Flush()
            return
        }
        
        // 构建响应数据
        result := &ChatMessage{
            ID:     rsp.Id,
            Delta:  rsp.Choices[0].Delta.Content,
            Detail: rsp,
        }
        if len(rsp.Choices[0].Delta.Content) > 0 {
            result.Text += rsp.Choices[0].Delta.Content
        }
        
        // 检查是否结束
        finishReason := rsp.Choices[0].FinishReason
        if finishReason == "stop" {
            result.FinishReason = finishReason
        }
        
        // 发送消息事件（SSE 格式）
        bts, _ := json.Marshal(result)
        ctx.Writer.Write([]byte("data: " + string(bts) + "\n\n"))
        
        // 如果是最后一条，发送结束事件
        if finishReason == "stop" {
            ctx.Writer.Write([]byte("event: done\n"))
            ctx.Writer.Write([]byte("data: {\"status\":\"complete\"}\n\n"))
        }
        
        ctx.Writer.Flush()
    }
}
```

**路由配置**

```go
// 在 routers 中添加新路由
r.POST("/chat-process-sse", chat.ChatProcessSSE)
```

#### 3.2.2 前端改造（ai-chat-web）

**方案一：使用 EventSource API（原生推荐）**

```typescript
// src/utils/sse.ts - SSE 工具类
export class SSERenderer {
  private eventSource: EventSource | null = null
  private onMessage: (data: any) => void
  private onError: (error: any) => void
  private onDone: () => void

  constructor(
    onMessage: (data: any) => void,
    onError: (error: any) => void,
    onDone: () => void
  ) {
    this.onMessage = onMessage
    this.onError = onError
    this.onDone = onDone
  }

  connect(url: string, body: any, signal: AbortSignal) {
    // EventSource 只支持 GET 请求，需要将 body 转为 URL 参数
    // 或者使用 fetch + ReadableStream
    
    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
      signal,
    }).then(response => {
      const reader = response.body?.getReader()
      const decoder = new TextDecoder()

      const read = () => {
        reader?.read().then(({ done, value }) => {
          if (done) {
            this.onDone()
            return
          }

          const chunk = decoder.decode(value)
          // 解析 SSE 格式
          this.parseSSEMessage(chunk)
          read()
        })
      }

      read()
    }).catch(this.onError)
  }

  private parseSSEMessage(chunk: string) {
    // 按行分割
    const lines = chunk.split('\n')
    let eventType = 'message'
    let data = ''

    for (const line of lines) {
      if (line.startsWith('event:')) {
        eventType = line.slice(6).trim()
      } else if (line.startsWith('data:')) {
        data = line.slice(5).trim()
      }
    }

    if (data) {
      const parsed = JSON.parse(data)
      
      switch (eventType) {
        case 'start':
          // 开始事件
          break
        case 'done':
          // 完成事件
          this.onDone()
          break
        case 'error':
          // 错误事件
          this.onError(parsed)
          break
        default:
          // 普通消息事件
          this.onMessage(parsed)
      }
    }
  }

  close() {
    this.eventSource?.close()
  }
}
```

**方案二：使用 fetch + ReadableStream（推荐，支持 POST）**

在 `src/api/index.ts` 中添加 SSE 请求方法：

```typescript:1:60:ai-chat-web\src\api\index.ts
import axios from 'axios'

// ... 现有代码 ...

// 新增 SSE 请求方法
export function fetchChatAPISSE<T = any>(
  url: string,
  data: any,
  options?: {
    signal?: AbortSignal
    onMessage?: (data: T) => void
    onError?: (error: any) => void
    onDone?: () => void
  }
): Promise<void> {
  return new Promise((resolve, reject) => {
    const { signal, onMessage, onError, onDone } = options || {}

    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
      signal,
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        const reader = response.body?.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        const read = () => {
          reader?.read().then(({ done, value }) => {
            if (done) {
              // 处理剩余缓冲数据
              if (buffer.trim()) {
                const parsed = parseSSELine(buffer)
                if (parsed && parsed.event === 'message') {
                  onMessage?.(parsed.data)
                }
              }
              onDone?.()
              resolve()
              return
            }

            // 解码并追加到缓冲区
            buffer += decoder.decode(value, { stream: true })

            // 按双换行符分割（每个 SSE 消息以 \n\n 结束）
            const messages = buffer.split('\n\n')
            buffer = messages.pop() || ''

            for (const message of messages) {
              const parsed = parseSSELine(message)
              if (!parsed) continue

              switch (parsed.event) {
                case 'start':
                  // 开始事件
                  break
                case 'done':
                  onDone?.()
                  resolve()
                  return
                case 'error':
                  onError?.(parsed.data)
                  break
                case 'message':
                default:
                  onMessage?.(parsed.data)
              }
            }

            read()
          })
        }

        read()
      })
      .catch(err => {
        if (err.name === 'AbortError') {
          onDone?.()
          resolve()
        } else {
          onError?.(err)
          reject(err)
        }
      })
  })
}

// 解析单条 SSE 消息
function parseSSELine(line: string): { event: string; data: any } | null {
  let event = 'message'
  let dataStr = ''

  for (const l of line.split('\n')) {
    if (l.startsWith('event:')) {
      event = l.slice(6).trim()
    } else if (l.startsWith('data:')) {
      dataStr = l.slice(5).trim()
    }
  }

  if (!dataStr) return null

  try {
    return {
      event,
      data: JSON.parse(dataStr),
    }
  } catch {
    return null
  }
}
```

**在聊天页面中使用 SSE**

修改 `src/views/chat/index.vue`：

```typescript
// 修改前
import { fetchChatAPIProcess } from '@/api'

// 修改后
import { fetchChatAPIProcess, fetchChatAPISSE } from '@/api'

// 使用 SSE 的聊天函数
const fetchChatAPIOnce = async () => {
  // ... 现有代码 ...

  // 使用 SSE
  await fetchChatAPISSE<Chat.ConversationResponse>(
    '/api/chat-process-sse',
    {
      prompt: message,
      options,
      signal: controller.signal,
    },
    {
      signal: controller.signal,
      onMessage: (data) => {
        // 更新聊天消息
        updateChat(+uuid, dataSources.value.length - 1, {
          dateTime: new Date().toLocaleString(),
          text: lastText + data.text ?? '',
          inversion: false,
          error: false,
          loading: false,
          conversationOptions: {
            conversationId: data.conversationId,
            parentMessageId: data.id,
          },
          requestOptions: { prompt: message, options: { ...options } },
        })
      },
      onError: (error) => {
        console.error('SSE Error:', error)
      },
      onDone: () => {
        // 聊天完成
      },
    }
  )
}
```

---

## 四、路由规划

### 4.1 兼容方案（推荐）

为了保证向后兼容，建议同时支持两种协议：

| 路由 | 协议 | 说明 |
|------|------|------|
| `/api/chat-process` | 原协议 | 保持向后兼容 |
| `/api/chat-process-sse` | SSE | 新增 SSE 支持 |
| `/api/chat` | 原协议 | 保持向后兼容 |
| `/api/chat-sse` | SSE | 新增 SSE 支持 |

### 4.2 完整切换方案

如果需要完全切换到 SSE：

1. 修改后端路由：`/api/chat-process` 直接使用 SSE
2. 修改前端 API 调用：使用新的 SSE 方法
3. 移除旧的 `onDownloadProgress` 相关代码

---

## 五、Nginx 配置（如使用 Nginx 代理）

如果后端服务部署在 Nginx 后面，需要禁用 Nginx 的缓冲：

### 5.1 方案一：Nginx 配置

```nginx
location /api/ {
    # 禁用代理缓冲
    proxy_buffering off;
    proxy_cache off;
    
    # 设置正确的响应头
    proxy_set_header Connection "";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    
    # 禁用缓存
    proxy_set_header Cache-Control "no-cache";
    
    # 设置超时
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
    
    proxy_pass http://backend;
}
```

### 5.2 方案二：使用 X-Accel-Buffering 头

后端设置响应头后，Nginx 会自动识别：

```go
ctx.Header("X-Accel-Buffering", "no")
```

---

## 六、改造后的数据流示例

### 6.1 SSE 格式响应

```sse
event: start
data: {"id":"chat-001","created":1708838400}

data: {"id":"chat-001","text":"","delta":"你好","detail":{...}}

data: {"id":"chat-001","text":"你好","delta":"！","detail":{...}}

data: {"id":"chat-001","text":"你好！","delta":"有什么","detail":{...}}

data: {"id":"chat-001","text":"你好！有什么","delta":"可以","detail":{...}}

data: {"id":"chat-001","text":"你好！有什么可以","delta":"帮助","detail":{...}}

data: {"id":"chat-001","text":"你好！有什么可以帮助你的","delta":"？","detail":{...}}

data: {"id":"chat-001","text":"你好！有什么可以帮助你的？","delta":"","detail":{...},"finish_reason":"stop"}

event: done
data: {"status":"complete"}
```

### 6.2 事件类型说明

| 事件名 | 说明 | 数据格式 |
|--------|------|----------|
| `start` | 开始生成响应 | `{id, created}` |
| `message` | 普通消息块（默认） | `{id, text, delta, detail, ...}` |
| `done` | 生成完成 | `{status: "complete"}` |
| `error` | 发生错误 | `{error: "错误信息"}` |

---

## 七、改造优势

### 7.1 使用 SSE 的优势

| 优势 | 说明 |
|------|------|
| **标准协议** | W3C 标准，浏览器原生支持 |
| **自动重连** | 连接断开后自动重连 |
| **事件支持** | 可区分不同类型的事件 |
| **简单易用** | EventSource API 简单易用 |
| **兼容性** | 兼容所有现代浏览器 |
| **调试友好** | 易于在浏览器开发者工具中查看 |

### 7.2 与原方案对比

| 特性 | 原方案（Chunked+JSON） | SSE 方案 |
|------|----------------------|----------|
| 标准性 | 自定义协议 | W3C 标准 |
| 浏览器支持 | 需要 XMLHttpRequest | 原生 EventSource |
| 事件类型 | 无 | 支持命名事件 |
| 自动重连 | 无 | 支持 |
| 调试难度 | 较难 | 容易 |
| Nginx 兼容 | 需要特殊配置 | 需要禁用缓冲 |

---

## 八、注意事项

### 8.1 潜在问题

1. **Nginx 缓冲**：Nginx 默认会缓冲 SSE 响应，需要禁用
2. **代理超时**：长时间连接可能触发代理超时，需要配置超时时间
3. **移动端电池**：SSE 保持长连接，可能影响移动端电池寿命
4. **负载均衡**：需要确保请求路由到同一服务器（使用 sticky session）

### 8.2 性能优化建议

1. **消息批量**：在高并发场景下，可以批量发送多条消息减少网络往返
2. **压缩传输**：启用 gzip 压缩减少传输数据量
3. **心跳机制**：定期发送心跳消息保持连接活跃

---

## 九、完整代码改动清单

### 9.1 后端改动（ai-chat-backend）

| 文件 | 改动说明 |
|------|----------|
| `pkg/controllers/chat.go` | 新增 SSE 版本的处理函数 |
| `routers/routers.go` | 注册新的 SSE 路由 |

### 9.2 前端改动（ai-chat-web）

| 文件 | 改动说明 |
|------|----------|
| `src/api/index.ts` | 新增 `fetchChatAPISSE` 函数 |
| `src/views/chat/index.vue` | 可选：切换使用 SSE 接口 |

---

## 十、总结

本项目当前使用的是 **HTTP Chunked + 自定义 JSON 行分隔协议**，通过手动解析 `\n` 分隔的 JSON 来实现流式响应。

改造为 **SSE 协议** 后，可以：

- 使用标准协议，更易于调试和维护
- 支持事件类型区分（如开始、消息、结束、错误）
- 获得浏览器原生的自动重连支持
- 代码更加清晰和现代化

**推荐采用兼容方案**：保留原有接口，新增 SSE 接口，逐步迁移。

---

*文档生成时间：2026-02-25*
