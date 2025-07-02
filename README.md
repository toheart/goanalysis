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
   # Linux
   ./goanalysis-linux-v* server
   
   # Windows  
   goanalysis-windows-v*.exe server
   ```
3. Open http://localhost:8000

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

Edit `configs/config.yaml`:

```yaml
server:
  http:
    addr: 0.0.0.0:8000
  grpc:
    addr: 0.0.0.0:9000

data:
  dbpath: ./goanalysis.db

biz:
  gitlab:
    token: "${GITLAB_TOKEN}"
    url: "${GITLAB_API_URL}"
```

## 📡 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/gids` | GET | Get goroutine IDs |
| `/api/functions` | GET | List traced functions |
| `/api/traces/{gid}` | GET | Get trace details |
| `/api/traces/{gid}/mermaid` | GET | Get diagram data |

## 🔧 Usage

### Basic Tracing
```bash
./goanalysis server
./goanalysis rewrite -d /path/to/project
```

### Git Analysis  
```bash
export GITLAB_TOKEN="your-token"
./goanalysis gitanalysis --project=123 --mr=45
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
