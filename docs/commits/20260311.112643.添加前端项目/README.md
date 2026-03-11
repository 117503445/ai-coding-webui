# 添加前端项目

## 主要内容和目的

初始化前端项目，使用 pnpm + Vite + TypeScript + React + Tailwind CSS + shadcn/ui 技术栈，设计风格为浅色、科技感。

## 更改内容描述

### 新增文件

1. **fe/** - 前端项目目录
   - `package.json` - 项目依赖配置
   - `vite.config.ts` - Vite 构建配置
   - `tsconfig.json` - TypeScript 配置
   - `src/App.tsx` - 主应用组件
   - `src/main.tsx` - 入口文件
   - `src/index.css` - 全局样式（科技感主题）
   - `src/components/ui/button.tsx` - shadcn/ui Button 组件
   - `src/lib/utils.ts` - 工具函数

2. **scripts/tasks/fe/Taskfile.yml** - 前端任务配置
   - `fe-dev` - 开发模式运行前端服务
   - `fe-build` - 构建前端生产版本
   - `fe-preview` - 预览前端生产版本

### 修改文件

1. **Taskfile.yml** - 添加 fe 任务目录引用
2. **scripts/tasks/base/Taskfile.yml** - 添加中文描述
3. **scripts/tasks/build/Taskfile.yml** - 添加中文描述
4. **scripts/tasks/run/Taskfile.yml** - 添加中文描述
5. **scripts/tasks/gen/Taskfile.yml** - 添加中文描述
6. **scripts/tasks/format/Taskfile.yml** - 添加中文描述

## 验证方法和结果

```bash
cd fe && pnpm dev
```

访问 http://localhost:5173 可看到科技感主题的前端页面。
