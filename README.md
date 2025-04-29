# 🔍 FuncTrace Analyzer

<p align="right">
  <a href="README.ZH.md">中文版</a> |
  <a href="README.md">English Version</a>
</p>

<div align="center">
  <h1>FuncTrace Analyzer</h1>
  <h3>Go Function Tracing Analysis & Visualization Expert System</h3>
  <p><strong>Current Version: v1.1.4</strong></p>

  ![License](https://img.shields.io/badge/License-MIT-blue.svg)
  ![Version](https://img.shields.io/badge/Version-v1.0.0-brightgreen.svg)
  ![Status](https://img.shields.io/badge/Status-Developing-orange.svg)
  ![Language](https://img.shields.io/badge/Language-Golang%20|%20Vue-yellow.svg)
</div>

## 🌟 Project Overview

**FuncTrace Analyzer** is a professional Go function tracing analysis tool that helps developers deeply understand function call relationships and performance bottlenecks through visualization technologies. The system uses the Kratos microservices framework for the backend and Vue.js for the frontend, providing a complete solution from data collection to 3D visualization.

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

## 🛠️ Technology Stack

| Domain            | Technologies               |
|-------------------|----------------------------|
| **Backend**       | Kratos (Microservices)     |
| **Frontend**      | Vue3 + Composition API     |
| **Visualization** | Mermaid.js + ECharts       |
| **Storage**       | SQLite + WAL Mode + Ent    |
| **Search**        | fuse.js fuzzy search       |
| **Deployment**    | Docker + Kubernetes-ready  |

## 📂 Project Structure

```
├── api                 # API definitions (protobuf)
├── cmd                 # Main applications
├── configs             # Configuration files
├── internal            # Private application code
│   ├── biz             # Business logic
│   ├── data            # Data processing and storage (Ent)
│   ├── server          # Server implementations
│   └── service         # Service implementations
├── third_party         # Third party dependencies
└── README.md           # This file
```

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

- **Description**: SQLite data storage/query using Ent
- **Details**:
  - Type-safe database operations
  - CRUD operations for trace data

### 5. CORS Support

- **Description**: Cross-origin resource sharing
- **Details**: CORS middleware configuration

## 🚀 Quick Start

### Prerequisites

- Go 1.19+
- Node.js 16+
- SQLite3 3.36+

### Backend Setup

```bash
# Clone repository
git clone https://github.com/toheart/goanalysis.git

# start server 
go run . server

# Instrumentation
go run . rewrite -d <path-to>
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

## 🔧 Kratos Features Utilized

- **Service Discovery**: Built-in service registry and discovery
- **Error Handling**: Structured error handling and recovery
- **Logging & Tracing**: Comprehensive logging and distributed tracing
- **Project Structure**: Following Kratos recommended conventions

## GitHub Actions Pipeline and Docker Images  

This project has configured a GitHub Actions pipeline for automatically building and publishing Docker images and packages.

### Automatic Build Process

When code is pushed to the `main` branch or a new tag (in the format of `v*`, such as `v1.0.0`) is created, the build process is automatically triggered:

1. Checkout code
2. Set up Go environment
3. Retrieve version information
4. Sync frontend code (from the latest release version of https://github.com/toheart/goanalysis-web)
5. Build application
6. Package Linux and Windows versions
7. Build and push Docker image (only when pushing to a branch or tag)
8. Create GitHub Release (only when creating a tag)

### Frontend Version Synchronization

The system will automatically fetch the latest release version from the https://github.com/toheart/goanalysis-web repository for building:

1. Retrieve the latest release version information via GitHub API
2. Download the corresponding release package or source code
3. If the release package contains a compiled dist directory, use it directly
4. If only the source code is available, it will be compiled automatically
5. The Release notes will include the frontend version information used

## 📜 Version History

| Version | Date       | Milestone                   |
| :------ | :--------- | :-------------------------- |
| v1.0.0  | 2025-03-09 | Official release            |
| v0.9.0  | 2025-02-25 | Distributed tracing support |
| v0.8.0  | 2025-02-18 | Parameter heatmap analysis  |

## 📞 Contact

- **Maintainer**: [toheart](https://github.com/toheart)
- **Issues**: [GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **WeChat**: [小唐的技术日志](https://mp.weixin.qq.com/)

<div align="center">
	<p><strong>FuncTrace Analyzer</strong> - Powered by Kratos+Vue Tech Stack</p> 
	<p><i>📌 Last Updated: 2025-03-09 CST</i></p>
	<hr>
</div>
