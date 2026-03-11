# ai-coding-webui

基于 `connectrpc` 的 Go 后端初始化模板，包含：

- `connect` RPC 服务 (`cmd/rpc`)
- `go-task` 任务拆分 (`Taskfile.yml` + `scripts/tasks/*`)
- Go 编写的项目脚本 (`scripts/script`)
- `buf` + `protoc-gen-go` + `protoc-gen-connect-go` 的代码生成流程

## 快速开始

```bash
# 生成 proto 代码
buf generate

# 构建 rpc 二进制
cd scripts/script && go run . build

# 运行服务
./data/rpc/rpc
```

默认监听 `:8080`。
