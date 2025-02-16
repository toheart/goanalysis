# 项目 README

## 项目概述

该项目是一个基于 Go 和 Vue.js 的分析工具，主要用于跟踪和分析函数调用的性能。它提供了一个前端界面，用户可以通过输入函数名称来查询相关的 goroutine，并查看详细的跟踪信息和调用图。

## 技术栈

- **后端**: Go (使用 Kratos 框架)
- **前端**: Vue.js
- **数据库**: SQLite
- **样式**: Bootstrap

## 功能

### 1. Trace Data Viewer

- **功能描述**: 用户可以输入函数名称，系统会展示与该函数相关的 goroutine。
- **组件**: `TraceViewer.vue`
- **实现细节**:
  - 使用输入框和下拉列表动态过滤函数名称。
  - 通过 API 请求获取相关的 GIDs。
  - 使用 `fuse.js` 实现模糊搜索功能。

### 2. Trace Details

- **功能描述**: 显示特定 GID 的详细跟踪信息。
- **组件**: `TraceDetails.vue`
- **实现细节**:
  - 通过 GID 获取跟踪数据并展示。
  - 提供查看参数的功能，用户可以点击按钮查看特定函数的参数。

### 3. Mermaid Viewer

- **功能描述**: 以图形化方式展示函数调用关系。
- **组件**: `MermaidViewer.vue`
- **实现细节**:
  - 使用 Mermaid.js 渲染函数调用图。
  - 支持缩放和拖动功能。

### 4. 数据库操作

- **功能描述**: 通过 SQLite 数据库存储和查询跟踪数据。
- **实现细节**:
  - 使用 `Data` 结构体封装数据库操作。
  - 提供获取所有 GIDs、函数名称和根据函数名称查询 GIDs 的功能。

### 5. CORS 支持

- **功能描述**: 解决跨域请求问题。
- **实现细节**: 在 Go 服务器中配置 CORS 以允许来自不同源的请求。

## API 端点

- `GET /api/gids`: 获取所有 GIDs。
- `GET /api/functions`: 获取所有函数名称。
- `POST /api/gids/function`: 根据函数名称获取相关 GIDs。
- `GET /api/traces/{gid}`: 获取特定 GID 的跟踪数据。
- `GET /api/params/{id}`: 获取特定 ID 的参数数据。
- `GET /api/traces/{gid}/mermaid`: 获取特定 GID 的 Mermaid 图数据。

## 项目结构

```
.
├── cmd
│   └── server
│      └── server.go          # 服务器启动入口
│   └── rewrite.go           # 重写命令
│   └── main.go              # 程序入口
├── internal
│   ├── data
│   │   └── data.go            # 数据库操作
│   ├── service
│   │   └── analysis.go        # 业务逻辑
│   └── server
│       └── server.go          # 服务器配置
├── functrace
│   └── trace.go               # 函数跟踪实现
├── static
│   └── analysis
│       ├── src
│       │   ├── components
│       │   │   ├── MermaidViewer.vue
│       │   │   ├── TraceDetails.vue
│       │   │   └── TraceViewer.vue
│       │   ├── App.vue
│       │   └── main.js
│       └── vue.config.js       # Vue 配置
└── api
    └── analysis
        └── v1
            ├── analysis.proto   # gRPC 接口定义
            └── analysis_grpc.pb.go # gRPC 生成的代码
```

## 安装与运行

### 1. 后端

- 确保安装 Go 环境。
- 在项目根目录下运行以下命令启动服务器：

```bash
go run cmd/server/server.go
```

### 2. 前端

- 确保安装 Node.js 和 npm。
- 在 `frontWeb/` 目录下运行以下命令安装依赖并启动前端：

```bash
npm install
npm run serve
```

## 贡献

欢迎任何形式的贡献！请提交问题或拉取请求。

## 许可证

该项目遵循 MIT 许可证。请查看 LICENSE 文件以获取更多信息。