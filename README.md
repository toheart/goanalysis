# 🔍 FuncTrace Analyzer

<p align="right">
  <a href="README.ZH.md">中文版</a> |
  <a href="README.md">English Version</a>
</p>

<div align="center">
  <h1>FuncTrace Analyzer</h1>
  <h3>Go Function Tracing Analysis & Visualization Expert System</h3>
  <p><strong>Current Version: v1.0.0</strong></p>

  ![License](https://img.shields.io/badge/License-MIT-blue.svg)
  ![Version](https://img.shields.io/badge/Version-v1.0.0-brightgreen.svg)
  ![Status](https://img.shields.io/badge/Status-Developing-orange.svg)
  ![Language](https://img.shields.io/badge/Language-Golang%20|%20Vue-yellow.svg)
</div>

## 🌟 Project Overview

**FuncTrace Analyzer** is a professional Go function tracing analysis tool that helps developers deeply understand function call relationships and performance bottlenecks through visualization technologies. The system combines the efficient Kratos framework backend with a dynamic Vue.js frontend, providing a complete solution from data collection to 3D visualization.

### 🚀 Core Features

- **Intelligent Function Tracing** - Real-time goroutine execution path capture
- **Multi-dimensional Analysis** - Time dimension, call depth, resource consumption analysis
- **Interactive Visualization** - Dynamic zoomable Mermaid flowcharts + parameter heatmaps
- **Smart Diagnostics** - Performance bottleneck prediction based on historical data
- **Cross-platform Support** - Lightweight SQLite storage solution

### 🎯 Design Goals

1. **Low-overhead Monitoring** - Under 5% performance overhead
2. **Zero-Intrusive Integration** - No code modification required
3. **Millisecond Response** - Fast query for 10M+ call chains
4. **Production-ready** - Rigorously stress-tested

## 🛠️ Technology Stack

| Domain            | Technologies               |
|-------------------|----------------------------|
| **Backend**       | Kratos (Microservices)     |
| **Frontend**      | Vue3 + Composition API     |
| **Visualization** | Mermaid.js + ECharts       |
| **Storage**       | SQLite + WAL Mode          |
| **Search**        | fuse.js fuzzy search       |
| **Deployment**    | Docker + Kubernetes-ready  |

## 🧩 Feature Modules

### 1. Smart Trace Viewer

- **Description**: Search and display goroutines related to specific functions
- **Component**: `TraceViewer.vue`
- **Details**:
  - Dynamic filtering with input and dropdown
  - API integration for GID retrieval
  - Fuzzy search using `fuse.js`

### 2. 3D Call Graph Visualization

- **Description**: Detailed trace analysis for specific GIDs
- **Component**: `TraceDetails.vue`
- **Details**:
  - Parameter inspection capabilities
  - Interactive timeline navigation

### 3. Parameter Heatmap Analysis

- **Description**: Visualize function call relationships
- **Component**: `MermaidViewer.vue`
- **Details**:
  - Mermaid.js rendering
  - Zoom/drag support

### 4. Database Operations

- **Description**: SQLite data storage/query
- **Details**:
  - `Data` struct encapsulation
  - CRUD operations for trace data

### 5. CORS Support

- **Description**: Cross-origin resource sharing
- **Details**: CORS middleware configuration

## ⚙️ System Architecture

```
.
├── cmd
│   └── server
│      └── server.go          # Server entry
│   └── rewrite.go           # Rewrite logic
│   └── main.go              # Main entry
├── internal
│   ├── data
│   │   └── data.go            # Database operations
│   ├── service
│   │   └── analysis.go        # Business logic
│   └── server
│       └── server.go          # Server config
├── functrace
│   └── trace.go               # Tracing implementation
├── static
│   └── analysis
│       ├── src
│       │   ├── components
│       │   │   ├── MermaidViewer.vue
│       │   │   ├── TraceDetails.vue
│       │   │   └── TraceViewer.vue
│       │   ├── App.vue
│       │   └── main.js
│       └── vue.config.js       # Vue config
└── api
    └── analysis
        └── v1
            ├── analysis.proto   # gRPC proto
            └── analysis_grpc.pb.go # Generated code
```

## 🚀 Quick Start

### Prerequisites

- Go 1.19+
- Node.js 16+
- SQLite3 3.36+

### Backend Setup

```bash
# Clone repository
git clone https://github.com/toheart/goanalysis.git

# Start server
go run cmd/server/server.go
```

### Frontend Setup

```bash
cd frontWeb

# Install dependencies
npm install

# Start development server
npm run serve
```

## 📡 API Reference

| Endpoint                    | Method | Description              |
| :-------------------------- | :----- | :----------------------- |
| `/api/gids`                 | GET    | Get all GIDs             |
| `/api/functions`            | GET    | List all functions       |
| `/api/gids/function`        | POST   | Find GIDs by function    |
| `/api/traces/{gid}`         | GET    | Get trace by GID         |
| `/api/params/{id}`          | GET    | Get parameters by ID     |
| `/api/traces/{gid}/mermaid` | GET    | Get Mermaid diagram data |

## 🤝 Contributing

We follow [Gitflow workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow):

1. Create feature branch: `git checkout -b feature/your-feature`
2. Commit atomic changes (follow Conventional Commits)
3. Write unit tests (≥80% coverage)
4. Update documentation
5. Create PR to `develop` branch

## 📜 Version History

| Version | Date       | Milestone                   |
| :------ | :--------- | :-------------------------- |
| v1.0.0  | 2025-03-09 | Official release            |
| v0.9.0  | 2025-02-25 | Distributed tracing support |
| v0.8.0  | 2025-02-18 | Parameter heatmap analysis  |

## 📞 Contact

- **Maintainer**: [toheart](https://github.com/toheart)
- **Issues**: [GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **WeChat**: [小唐云原生](https://mp.weixin.qq.com/)

------

<div align="center">
	<p><strong>FuncTrace Analyzer</strong> - Powered by Go+Vue Tech Stack</p> 
	<p><i>📌 Last Updated: 2025-03-09 CST</i></p>
	<hr>
</div>