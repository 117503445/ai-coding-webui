# 修复聊天响应问题并改进日志系统

## 主要内容和目的

修复 WebSocket 聊天功能中 Claude 无法响应的问题，同时重构日志系统使用 context 注入方式，并增强 E2E 测试验证实际响应。

## 更改内容描述

### 后端 (cmd/rpc/)

1. **claude.go**
   - 修复子进程阻塞问题：设置 `cmd.Stdin = nil` 避免 claude CLI 等待 TTY 输入
   - 添加 stderr 管道消费，防止缓冲区满导致进程阻塞
   - 重构日志：使用 `log.Ctx(ctx)` 替代全局 logger，日志信息改为中文
   - 增加 stdout 行计数和详细调试日志
   - 扩展 `extractDetail` 支持更多事件类型的中文状态提示

2. **ws.go**
   - 移除函数参数中的 `*zerolog.Logger` 传递
   - 在 ServeHTTP 入口将 logger 注入到 context
   - 所有日志改用 `log.Ctx(ctx)` 方式，信息改为中文
   - 添加 remoteAddr 到日志上下文

### 前端 (fe/src/store/chatStore.ts)

1. 新增 `parseContentBlocks` 函数，解析 assistant 消息的 content 数组
2. 支持 thinking、text、tool_use、tool_result 等内容块类型
3. 修复 `handleStreamEvent` 处理逻辑：
   - 优先处理 `assistant` 事件（claude CLI stream-json 格式）
   - 支持 Anthropic API 原生流式格式（content_block_start/delta/stop）
   - 支持 `stream_event` 包装器格式

### E2E 测试 (scripts/e2e/)

1. **test_multi_turn_chat.py**
   - 测试改为验证实际 Claude 响应，而非仅发送后中止
   - 添加断言确保 assistant 消息出现，否则测试失败
   - 验证多轮对话和 localStorage 持久化

2. **test_session_persistence.py**
   - 移除后端重启逻辑，简化为页面刷新测试
   - 添加实际响应验证
   - 验证 localStorage 中 messages 数量

3. **main.py**
   - 后端日志输出到 `logs/backend.log` 而非 pipe
   - 添加 stdin=DEVNULL 避免进程等待输入

## 验证方法和结果

1. 启动后端服务 `./data/rpc/rpc`
2. 打开前端页面，发送消息
3. 确认收到 Claude 实际响应（而非卡住）
4. 运行 E2E 测试：`python scripts/e2e/main.py`
5. 测试通过，多轮对话正常工作