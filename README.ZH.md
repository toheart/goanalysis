# 🔍 FuncTrace Analyzer

<p align="right">
  <a href="README.md">English Version</a> |
  <a href="README.ZH.md">中文版</a>
</p>

<div align="center">
  <h1>FuncTrace Analyzer</h1>
  <h3> Go 函数追踪分析与可视化专家系统</h3>
  <p><strong>当前版本：v1.0.0</strong></p>
  
  ![许可证](https://img.shields.io/badge/许可证-MIT-blue.svg)
  ![版本](https://img.shields.io/badge/版本-v1.0.0-brightgreen.svg)
  ![状态](https://img.shields.io/badge/状态-开发中-orange.svg)
  ![语言](https://img.shields.io/badge/语言-Golang%20|%20Vue-yellow.svg)
</div>


## 🌟 项目概览

**FuncTrace Analyzer** 是一款专业的 Go 函数追踪分析工具，通过可视化技术帮助开发者深入理解函数调用关系与性能瓶颈。系统结合 Kratos 框架的高效后端与 Vue.js 的动态前端，提供从数据采集到三维可视化的完整解决方案。

### 🚀 核心功能

- **智能函数追踪** - 实时捕获 goroutine 执行路径
- **多维数据分析** - 支持时间维度、调用深度、资源消耗等多角度分析
- **交互式可视化** - 动态可缩放的 Mermaid 流程图 + 参数热力图
- **智能诊断建议** - 基于历史数据的性能瓶颈预测
- **跨平台支持** - 全平台兼容的轻量级 SQLite 存储方案

### 🎯 设计目标

1. **性能无损监控** - 低于 5% 的性能损耗
2. **零侵入式集成** - 无需修改业务代码
3. **毫秒级响应** - 千万级调用链快速查询
4. **生产级可靠** - 经过严格压力测试验证

## 🛠️ 技术栈全景

| 领域           | 技术选型                 |
| -------------- | ------------------------ |
| **后端框架**   | Kratos (微服务架构)      |
| **前端框架**   | Vue3 + Composition API   |
| **数据可视化** | Mermaid.js + ECharts     |
| **存储引擎**   | SQLite + WAL 模式        |
| **搜索优化**   | fuse.js 模糊搜索         |
| **部署方案**   | Docker + Kubernetes 就绪 |

## 🧩 功能模块详解

### 1. 智能追踪视图 (TraceViewer)

- **功能描述**: 用户可以输入函数名称，系统会展示与该函数相关的 goroutine。
- **组件**: `TraceViewer.vue`
- **实现细节**:
  - 使用输入框和下拉列表动态过滤函数名称。
  - 通过 API 请求获取相关的 GIDs。
  - 使用 `fuse.js` 实现模糊搜索功能。

### 2. 三维调用图谱 (Mermaid Viewer)

- **功能描述**: 显示特定 GID 的详细跟踪信息。
- **组件**: `TraceDetails.vue`
- **实现细节**:
  - 通过 GID 获取跟踪数据并展示。
  - 提供查看参数的功能，用户可以点击按钮查看特定函数的参数。

### 3. 参数热力分析 (Parameter Analyzer)

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

## ⚙️ 系统架构

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

## 🚀 快速入门

### 先决条件

- Go 1.19+
- Node.js 16+
- SQLite3 3.36+

### 后端启动

- 确保安装 Go 环境。
- 在项目根目录下运行以下命令启动服务器：

```bash
go run cmd/server/server.go
```

### 前端启动

- 确保安装 Node.js 和 npm。
- 在 `frontWeb/` 目录下运行以下命令安装依赖并启动前端：

```bash
npm install
npm run serve
```

## 📡 API 参考

| 端点                     | 方法 | 描述                          |
| ------------------------ | ---- | ----------------------------- |
| `/api/gids`         | GET  | 获取所有 GIDs          |
| `/api/functions`  | GET | 获取所有函数名称 |
| `/api/gids/function`    | POST  | 根据函数名称获取相关 GIDs              |
| `/api/traces/{gid}` | GET  | 获取特定 GID 的跟踪数据        |
| `/api/params/{id}`        | GET  | 获取特定 ID 的参数数据            |
| `/api/traces/{gid}/mermaid`        | GET  | 获取特定 GID 的 Mermaid 图数据           |

## 🤝 贡献指南

我们遵循 [Gitflow 工作流](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow)，贡献流程包括：

1. 创建功能分支： `git checkout -b feature/your-feature`
2. 提交原子性变更：遵循 Conventional Commits 规范
3. 编写单元测试：覆盖率需 ≥ 80%
4. 更新文档：同步修改相关文档
5. 发起 Pull Request：至 develop 分支

## 📜 版本历史

| 版本   | 日期       | 里程碑               |
| ------ | ---------- | -------------------- |
| v1.0.0 | 2023-08-01 | 正式发布版           |
| v0.9.0 | 2023-07-25 | 增加分布式追踪支持   |
| v0.8.0 | 2023-07-18 | 实现参数热力分析功能 |

## 📞 联系我们

- **项目维护**：[toheart](https://github.com/toheart)
- **问题反馈**：[GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **微信公众号**：[小唐云原生](https://mp.weixin.qq.com/)

------

<div align="center">
  <p><strong>FuncTrace Analyzer</strong> - 由 Go+Vue 技术栈强力驱动</p>
  <p><i>📌 最后更新: 2023-08-01 CST</i></p>
  <hr>
</div>
