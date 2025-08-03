# 🔍 GoAnalysis

<p align="right">
  <a href="README.ZH.md">中文版</a> |
  <a href="README.md">English</a>
</p>

<div align="center">
  <h1>GoAnalysis</h1>
  <h3>Go Function Tracing & Visualization Tool</h3>
  
  ![License](https://img.shields.io/badge/License-MIT-blue.svg)
  ![Version](https://img.shields.io/badge/Version-v1.1.4-brightgreen.svg)
  ![Go](https://img.shields.io/badge/Language-Go%20|%20Vue3-yellow.svg)
</div>

## 🌟 Overview

Professional Go function tracing analysis tool with advanced visualization. Built with **Kratos** backend and **Vue3** frontend.

## 🚀 Features

- **🔍 Function Tracing** - Real-time goroutine execution capture
- **📊 Visualization** - Interactive Mermaid flowcharts and heatmaps  
- **📈 Performance** - Bottleneck identification and analysis
- **🔄 Git Integration** - GitLab MR change analysis
- **🌐 Web UI** - Modern Vue3 interface

## 🛠️ Tech Stack

- **Backend**: Kratos, gRPC, SQLite
- **Frontend**: Vue3, Bootstrap, ECharts
- **Visualization**: Mermaid.js, D3.js

## 🚀 Quick Start

### Using Pre-built Binaries

1. Download from [GitHub Releases](https://github.com/toheart/goanalysis/releases)
2. Extract and run:
   ```bash
   # Linux - Start with default configuration
   ./goanalysis-linux-v* server
   
   # Windows - Start with default configuration
   goanalysis-windows-v*.exe server
   ```
3. Open http://localhost:8001

### Building from Source

```bash
git clone https://github.com/toheart/goanalysis.git
cd goanalysis
make init
make sync-frontend  
make build
./bin/goanalysis server
```

## ⚙️ Configuration

### 🎯 Recommended: Command Line Arguments (No Config File Required)

GoAnalysis now supports direct configuration via command line arguments without creating config files:

```bash
# Start with default configuration
./goanalysis server

# Customize port and log level
./goanalysis server --http-addr=0.0.0.0:8080 --log-level=info

# Customize database path
./goanalysis server --db-path=./my-database.db

# Set GitLab configuration
./goanalysis server --gitlab-token="your-token" --gitlab-url="https://gitlab.com/api/v4"
```

### 📋 All Available Parameters

```bash
# View all available parameters
./goanalysis server --help
```

**Server Configuration:**
- `--http-addr` - HTTP server address (default: 0.0.0.0:8001)
- `--grpc-addr` - gRPC server address (default: 0.0.0.0:9000)
- `--http-timeout` - HTTP timeout (default: 1s)
- `--grpc-timeout` - gRPC timeout (default: 1s)

**Logging Configuration:**
- `--log-level` - Log level (default: debug)
- `--log-file` - Log file path (default: ./logs/app.log)
- `--log-console` - Output to console (default: true)

**GitLab Configuration:**
- `--gitlab-token` - GitLab access token
- `--gitlab-url` - GitLab API URL
- `--gitlab-clone-dir` - Clone directory (default: ./data)

**OpenAI Configuration:**
- `--openai-api-key` - OpenAI API key
- `--openai-api-base` - OpenAI API base URL
- `--openai-model` - OpenAI model name

**Storage Paths:**
- `--static-store-path` - Static storage path (default: ./data/static)
- `--runtime-store-path` - Runtime storage path (default: ./data/runtime)
- `--file-storage-path` - File storage path (default: ./data/files)

**Data Configuration:**
- `--db-path` - Database path (default: ./goanalysis.db)

### 🔐 Environment Variables Support

Sensitive information can also be set via environment variables:

```bash
export GITLAB_TOKEN="your-gitlab-token"
export GITLAB_API_URL="https://gitlab.com/api/v4"
export OPENAI_API_KEY="your-openai-key"
export OPENAI_API_BASE="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-3.5-turbo"

./goanalysis server
```

### 📄 Traditional: Configuration File

Still supports traditional configuration file approach:

```bash
# Generate default configuration file
./goanalysis config

# Start with configuration file
./goanalysis server --conf=configs/config.yaml
```

Configuration file example:

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

### 🔄 Configuration Priority

Configuration priority from high to low:
1. **Command Line Arguments** - Highest priority
2. **Environment Variables** - For sensitive information
3. **Configuration File** - Traditional approach
4. **Default Values** - Lowest priority



## 🔧 Usage

### Basic Tracing
```bash
# Start with default configuration
./goanalysis server

# Start with custom configuration
./goanalysis server --http-addr=0.0.0.0:8080 --log-level=info

# Code rewriting
./goanalysis rewrite -d /path/to/project
```

### Git Analysis  
```bash
# Set GitLab configuration
export GITLAB_TOKEN="your-token"
export GITLAB_API_URL="https://gitlab.com/api/v4"

# Start server
./goanalysis server
```

## 📂 Project Structure

```
├── api/           # API definitions
├── cmd/           # CLI commands  
├── internal/      # Core logic
├── web/           # Frontend files
└── configs/       # Configuration
```

## 🏗️ Deployment

### Docker
```bash
docker run -p 8000:8000 -p 9000:9000 \
  ghcr.io/toheart/goanalysis:latest
```

### Build
```bash
make package-linux
make package-windows
```

## 🔧 Troubleshooting

| Issue | Solution |
|-------|----------|
| Port in use | `lsof -i :8000; kill -9 <PID>` |
| DB locked | `rm -f goanalysis.db-*` |
| Frontend missing | `make sync-frontend` |

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Make changes with tests
4. Submit pull request

Follow [Conventional Commits](https://www.conventionalcommits.org/).

## 📜 Releases

| Version | Date | Changes |
|---------|------|---------|
| v1.1.4 | 2024-12-16 | GitLab integration |
| v1.1.0 | 2024-12-01 | Vue3 upgrade |
| v1.0.0 | 2024-11-15 | First stable |

## 📄 License

MIT License - see [LICENSE](LICENSE) file.

## 📞 Support

- **🐛 Issues**: [GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/toheart/goanalysis/discussions)
- **📖 Docs**: [Wiki](https://github.com/toheart/goanalysis/wiki)
- **📱 WeChat**: Follow "小唐的技术日志" for updates

<div align="center">
  <h4>📱 Follow WeChat</h4>
  <p><strong>小唐的技术日志</strong></p>
  <img src="docs/images/wechat-qr.jpg" alt="WeChat QR Code" width="200"/>
  <p><i>Scan for latest updates</i></p>
</div>

---

<div align="center">
  <p><strong>GoAnalysis</strong> - Empowering Go developers</p>
  <p>⭐ Star us on GitHub!</p>
</div>
