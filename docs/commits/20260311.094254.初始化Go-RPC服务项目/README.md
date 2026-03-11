# 初始化 Go RPC 服务项目

## 主要内容和目的

本项目初始化了一个基于 Go 语言的 RPC 服务框架，使用 connect-rpc 协议，为后续的 AI Coding WebUI 项目奠定基础。

## 更改内容描述

### 1. 项目初始化
- 创建 Go module (`github.com/117503445/ai-coding-webui`)
- 配置依赖：connect-rpc、zerolog、uuid、cors、protobuf

### 2. RPC 服务实现 (`cmd/rpc/`)
- `main.go`: 服务入口，支持环境变量配置端口
- `server.go`: HTTP 服务器启动和监听
- `handler.go`: TemplateService 处理器实现
- `context.go`: 请求上下文工具函数

### 3. Protobuf 定义 (`pkg/rpc/`)
- `template.proto`: 定义 TemplateService 及 Healthz 接口
- 生成代码：`template.pb.go`、`template.connect.go`

### 4. 构建脚本 (`scripts/`)
- Taskfile 任务定义：构建、运行、格式化、代码生成
- Go 脚本工具：支持构建和格式化操作

### 5. 构建信息 (`internal/buildinfo/`)
- 支持编译时注入版本、Git 信息等

### 6. 配置文件
- `buf.yaml` / `buf.gen.yaml`: Buf protobuf 工具配置
- `Taskfile.yml`: Task 任务运行器配置
- `.gitignore`: Git 忽略规则更新

## 验证方法和结果

### 编译验证
```bash
task build:bin
```

### 运行验证
```bash
task run:rpc-dev
```
服务启动后可通过 `http://localhost:8080/healthz` 访问健康检查接口。

### 代码生成验证
```bash
task gen:gen-rpc
```
确认 protobuf 代码生成正常。