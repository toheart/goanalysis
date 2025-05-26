# 🔍 FuncTrace Analyzer

<p align="right">
  <a href="README.ZH.md">中文版</a> |
  <a href="README.md">English Version</a>
</p>

<div align="center">
  <h1>FuncTrace 分析器</h1>
  <h3>Go函数追踪分析与可视化专家系统</h3>
  <p><strong>当前版本: v1.1.4</strong></p>

  ![许可证](https://img.shields.io/badge/License-MIT-blue.svg)
  ![版本](https://img.shields.io/badge/Version-v1.0.0-brightgreen.svg)
  ![状态](https://img.shields.io/badge/Status-开发中-orange.svg)
  ![语言](https://img.shields.io/badge/Language-Golang%20|%20Vue-yellow.svg)
</div>

## 🌟 项目概述

**goanalysis**是一款专业的Go函数追踪分析工具，通过可视化技术帮助开发者深入理解函数调用关系和性能瓶颈。系统使用Kratos微服务框架作为后端，Vue.js作为前端，提供从数据采集到3D可视化的完整解决方案。

### 🚀 核心功能

- **智能函数追踪** - 实时捕获goroutine执行路径
- **多维度分析** - 时间维度、调用深度、资源消耗分析
- **交互式可视化** - 动态可缩放Mermaid流程图 + 参数热力图
- **智能诊断** - 基于历史数据的性能瓶颈预测
- **跨平台支持** - 轻量级SQLite存储解决方案

### 🎯 设计目标

1. **低开销监控** - 性能开销低于5%
2. **零侵入集成** - 无需修改代码
3. **毫秒级响应** - 快速查询1000万+调用链

## 🛠️ 技术栈

| 领域              | 技术                       |
|-------------------|----------------------------|
| **后端**          | Kratos (微服务)            |
| **前端**          | Vue3 + Composition API     |
| **可视化**        | Mermaid.js + ECharts       |
| **存储**          | SQLite + WAL模式 + Ent     |
| **搜索**          | fuse.js模糊搜索            |
| **部署**          | Docker + Kubernetes就绪    |

## 📂 项目结构

```
├── api                 # API定义（protobuf）
├── cmd                 # 主应用程序
├── configs             # 配置文件
├── internal            # 私有应用代码
│   ├── biz             # 业务逻辑
│   ├── data            # 数据处理和存储（Ent）
│   ├── server          # 服务器实现
│   └── service         # 服务实现
├── third_party         # 第三方依赖
└── README.md           # 本文件
```

## 🧩 功能模块

### 1. 智能追踪查看器

- **描述**: 搜索并显示与特定函数相关的goroutines
- **组件**: `TraceViewer.vue`
- **详情**:
  - 动态过滤与下拉菜单
  - GID检索API集成
  - 使用`fuse.js`模糊搜索

### 2. 3D调用图可视化

- **描述**: 特定GID的详细追踪分析
- **组件**: `TraceDetails.vue`
- **详情**:
  - 参数检查功能
  - 交互式时间线导航

### 3. 参数热力图分析

- **描述**: 可视化函数调用关系
- **组件**: `MermaidViewer.vue`
- **详情**:
  - Mermaid.js渲染
  - 缩放/拖拽支持

### 4. 数据库操作

- **描述**: 使用Ent进行SQLite数据存储/查询
- **详情**:
  - 类型安全的数据库操作
  - 追踪数据的CRUD操作

### 5. CORS支持

- **描述**: 跨域资源共享
- **详情**: CORS中间件配置

## 🚀 快速开始

### 先决条件

- Go 1.19+
- Node.js 16+
- SQLite3 3.36+

### 后端设置

```bash
# 克隆仓库
git clone https://github.com/toheart/goanalysis.git

# 启动服务器
go run . server

# 插桩操作
go run . rewrite -d <path-to>
```

## 📡 API参考

| 端点                       | 方法   | 描述                     |
| :------------------------- | :----- | :----------------------- |
| `/api/gids`                | GET    | 获取所有GID              |
| `/api/functions`           | GET    | 列出所有函数             |
| `/api/gids/function`       | POST   | 按函数查找GID            |
| `/api/traces/{gid}`        | GET    | 按GID获取追踪            |
| `/api/params/{id}`         | GET    | 按ID获取参数             |
| `/api/traces/{gid}/mermaid`| GET    | 获取Mermaid图表数据      |

## 🤝 贡献

我们遵循[Gitflow工作流](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow):

1. 创建功能分支: `git checkout -b feature/your-feature`
2. 提交原子更改(遵循约定式提交)
3. 编写单元测试(≥80%覆盖率)
4. 更新文档
5. 创建PR至`develop`分支

## 🔧 Kratos特性应用

- **服务发现**: 内置服务注册与发现
- **错误处理**: 结构化错误处理与恢复
- **日志与追踪**: 全面的日志和分布式追踪
- **项目结构**: 遵循Kratos推荐的项目结构规范

## GitHub Actions流水线和Docker镜像

本项目配置了GitHub Actions流水线，用于自动构建和发布Docker镜像和软件包。

### 自动构建流程

当代码推送到`main`分支或创建新标签(格式为`v*`，如`v1.0.0`)时，构建流程自动触发:

1. 检出代码
2. 设置Go环境
3. 获取版本信息
4. 同步前端代码(从https://github.com/toheart/goanalysis-web的最新版本)
5. 构建应用
6. 打包Linux和Windows版本
7. 构建并推送Docker镜像(仅当推送到分支或标签时)
8. 创建GitHub发布(仅当创建标签时)

### 前端版本同步

系统将自动从https://github.com/toheart/goanalysis-web仓库获取最新发布版本进行构建:

1. 通过GitHub API获取最新发布版本信息
2. 下载相应的发布包或源代码
3. 如果发布包包含已编译的dist目录，则直接使用
4. 如果只有源代码可用，则自动编译
5. 发布说明将包含所使用的前端版本信息

## 📜 版本历史

| 版本    | 日期       | 里程碑                     |
| :------ | :--------- | :------------------------ |
| v1.0.0  | 2025-03-09 | 正式发布                  |
| v0.9.0  | 2025-02-25 | 分布式追踪支持            |
| v0.8.0  | 2025-02-18 | 参数热力图分析            |

## 📞 联系方式

- **维护者**: [toheart](https://github.com/toheart)
- **问题**: [GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **微信**: [小唐的技术日志](https://mp.weixin.qq.com/)

## Git改动分析

新增的Git改动分析功能可以分析GitLab上的Merge Request变更，识别出被修改的Go函数。详细说明请参考[Git分析功能文档](docs/git_analysis.md)。

主要特性：
- 自动拉取GitLab仓库和MR变更
- 使用AST解析识别变更的函数
- 支持JSON和表格格式输出结果
- 基于配置文件的灵活配置

使用示例：
```bash
# 配置文件中设置GitLab Token和URL
goanalysis gitanalysis --project=123 --mr=45 --output=./analysis-result.json
```

<div align="center">
	<p><strong>FuncTrace 分析器</strong> - 由Kratos+Vue技术栈驱动</p> 
	<p><i>📌 最后更新: 2025-03-09 CST</i></p>
	<hr>
</div>