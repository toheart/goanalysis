# 🔍 GoAnalysis

<p align="right">
  <a href="README.ZH.md">中文版</a> |
  <a href="README.md">English</a>
</p>

<div align="center">
  <h1>GoAnalysis</h1>
  <h3>Go函数追踪与可视化工具</h3>
  
  ![许可证](https://img.shields.io/badge/License-MIT-blue.svg)
  ![版本](https://img.shields.io/badge/Version-v1.1.4-brightgreen.svg)
  ![语言](https://img.shields.io/badge/Language-Go%20|%20Vue3-yellow.svg)
</div>

## 🌟 项目概述

专业的Go函数追踪分析工具，具备先进的可视化功能。使用 **Kratos** 后端和 **Vue3** 前端构建。

## 🚀 核心功能

- **🔍 函数追踪** - 实时捕获goroutine执行路径
- **📊 可视化** - 交互式Mermaid流程图和热力图
- **📈 性能分析** - 瓶颈识别和分析
- **🔄 Git集成** - GitLab MR变更分析
- **🌐 Web界面** - 现代化Vue3界面

## 🛠️ 技术栈

- **后端**: Kratos, gRPC, SQLite
- **前端**: Vue3, Bootstrap, ECharts
- **可视化**: Mermaid.js, D3.js

## 🚀 快速开始

### 使用预编译二进制文件

1. 从 [GitHub Releases](https://github.com/toheart/goanalysis/releases) 下载
2. 解压并运行:
   ```bash
   # Linux - 使用默认配置启动
   ./goanalysis-linux-v* server
   
   # Windows - 使用默认配置启动
   goanalysis-windows-v*.exe server
   ```
3. 打开 http://localhost:8001

### 从源码构建

```bash
git clone https://github.com/toheart/goanalysis.git
cd goanalysis
make init
make sync-frontend
make build
./bin/goanalysis server
```

## ⚙️ 配置说明

### 🎯 推荐方式：命令行参数（无需配置文件）

GoAnalysis 现在支持通过命令行参数直接配置，无需创建配置文件：

```bash
# 使用默认配置启动
./goanalysis server

# 自定义端口和日志级别
./goanalysis server --http-addr=0.0.0.0:8080 --log-level=info

# 自定义数据库路径
./goanalysis server --db-path=./my-database.db

# 设置GitLab配置
./goanalysis server --gitlab-token="your-token" --gitlab-url="https://gitlab.com/api/v4"
```

### 📋 所有可用参数

```bash
# 查看所有可用参数
./goanalysis server --help
```

**服务器配置：**
- `--http-addr` - HTTP服务地址 (默认: 0.0.0.0:8001)
- `--grpc-addr` - gRPC服务地址 (默认: 0.0.0.0:9000)
- `--http-timeout` - HTTP超时时间 (默认: 1s)
- `--grpc-timeout` - gRPC超时时间 (默认: 1s)

**日志配置：**
- `--log-level` - 日志级别 (默认: debug)
- `--log-file` - 日志文件路径 (默认: ./logs/app.log)
- `--log-console` - 是否输出到控制台 (默认: true)

**GitLab配置：**
- `--gitlab-token` - GitLab访问令牌
- `--gitlab-url` - GitLab API地址
- `--gitlab-clone-dir` - 克隆目录 (默认: ./data)

**OpenAI配置：**
- `--openai-api-key` - OpenAI API密钥
- `--openai-api-base` - OpenAI API地址
- `--openai-model` - OpenAI模型名称

**存储路径：**
- `--static-store-path` - 静态存储路径 (默认: ./data/static)
- `--runtime-store-path` - 运行时存储路径 (默认: ./data/runtime)
- `--file-storage-path` - 文件存储路径 (默认: ./data/files)

**数据配置：**
- `--db-path` - 数据库路径 (默认: ./goanalysis.db)

### 🔐 环境变量支持

敏感信息也可以通过环境变量设置：

```bash
export GITLAB_TOKEN="your-gitlab-token"
export GITLAB_API_URL="https://gitlab.com/api/v4"
export OPENAI_API_KEY="your-openai-key"
export OPENAI_API_BASE="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-3.5-turbo"

./goanalysis server
```

### 📄 传统方式：配置文件

仍然支持传统的配置文件方式：

```bash
# 生成默认配置文件
./goanalysis config

# 使用配置文件启动
./goanalysis server --conf=configs/config.yaml
```

配置文件示例：

```yaml
server:
  http:
    addr: 0.0.0.0:8001
  grpc:
    addr: 0.0.0.0:9000

logger:
  level: debug
  file_path: ./logs/app.log
  console: true

biz:
  gitlab:
    token: "${GITLAB_TOKEN}"
    url: "${GITLAB_API_URL}"
    clone_dir: ./data
  openai:
    api_key: "${OPENAI_API_KEY}"
    api_base: "${OPENAI_API_BASE}"
    model: "${OPENAI_MODEL}"

data:
  dbpath: ./goanalysis.db
```

### 🔄 配置优先级

配置优先级从高到低：
1. **命令行参数** - 最高优先级
2. **环境变量** - 用于敏感信息
3. **配置文件** - 传统方式
4. **默认值** - 最低优先级


## 🔧 使用方法

### 基础追踪
```bash
# 使用默认配置启动
./goanalysis server

# 自定义配置启动
./goanalysis server --http-addr=0.0.0.0:8080 --log-level=info

# 代码重写
./goanalysis rewrite -d /path/to/project
```

### Git分析
```bash
# 设置GitLab配置
export GITLAB_TOKEN="your-token"
export GITLAB_API_URL="https://gitlab.com/api/v4"

# 启动服务
./goanalysis server
```

## 📂 项目结构

```