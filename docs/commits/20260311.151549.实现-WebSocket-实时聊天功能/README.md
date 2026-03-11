# WebSocket 实时聊天功能实现

## 主要内容和目的

实现 Claude Code WebUI 的核心功能，包括：
- WebSocket 实时通信，支持流式聊天
- Claude CLI 进程管理和状态追踪
- 前端 React 组件和状态管理
- E2E 自动化测试框架

## 更改内容描述

### 后端变更

1. **Claude CLI 管理** (`cmd/rpc/claude.go`)
   - 新增 `ClaudeManager` 结构体管理 claude CLI 进程
   - 支持启动、中止、状态查询
   - 解析 stream-json 输出提取 session_id 和状态详情

2. **WebSocket 处理** (`cmd/rpc/ws.go`)
   - 实现 WebSocket 协议处理
   - 支持 chat、command、abort 三种消息类型
   - 流式转发 claude CLI 输出到前端

3. **RPC 服务扩展** (`pkg/rpc/template.proto`)
   - 新增 `GetStatus` 和 `Abort` RPC 方法
   - 服务名从 `TemplateService` 改为 `ClaudeService`

4. **前端静态服务** (`cmd/rpc/frontend_dev.go`, `frontend_prod.go`)
   - 开发模式代理到 Vite 开发服务器
   - 生产模式使用 go:embed 嵌入前端资源

### 前端变更

1. **聊天组件** (`fe/src/components/chat/`)
   - `ChatContainer`: 主容器组件
   - `MessageList/MessageItem`: 消息列表展示
   - `ThinkingBlock`: 思考过程折叠展示
   - `ToolCallBlock`: 工具调用卡片展示
   - `MarkdownRenderer`: Markdown 渲染与复制
   - `InputArea`: 输入框与斜杠命令
   - `StatusBar`: 状态栏显示连接和工作状态
   - `ConnectionOverlay`: 断线重连提示

2. **WebSocket 客户端** (`fe/src/lib/ws.ts`)
   - 指数退避重连机制
   - 消息队列管理
   - 事件订阅模式

3. **状态管理** (`fe/src/store/chatStore.ts`)
   - 使用 zustand 管理聊天状态
   - localStorage 持久化 session_id 和消息缓存

### E2E 测试

- 基于 Playwright 实现自动化测试
- 支持多轮会话测试、页面加载测试、会话持久化测试、斜杠命令测试
- 测试结果输出到 `data/e2e/cases/` 目录

### 文档

- 新增架构设计文档 `docs/design.md`，包含 ASCII 架构图

## 验证方法和结果

1. **后端测试**
   ```bash
   go build ./cmd/rpc
   ./data/rpc/rpc
   ```
   - 服务启动正常
   - WebSocket 连接正常
   - Claude CLI 调用正常

2. **前端测试**
   ```bash
   cd fe && pnpm dev
   ```
   - UI 渲染正常
   - WebSocket 连接正常
   - 流式消息展示正常

3. **E2E 测试**
   ```bash
   cd scripts/e2e && uv run main.py
   ```
   - 所有测试用例通过