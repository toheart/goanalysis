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
   # Linux
   ./goanalysis-linux-v* server
   
   # Windows
   goanalysis-windows-v*.exe server
   ```
3. 打开 http://localhost:8000

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

编辑 `configs/config.yaml`:

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

## 📡 API接口

| 接口地址 | 方法 | 描述 |
|----------|------|------|
| `/api/gids` | GET | 获取goroutine ID |
| `/api/functions` | GET | 列出追踪函数 |
| `/api/traces/{gid}` | GET | 获取追踪详情 |
| `/api/traces/{gid}/mermaid` | GET | 获取图表数据 |

## 🔧 使用方法

### 基础追踪
```bash
./goanalysis server
./goanalysis rewrite -d /path/to/project
```

### Git分析
```bash
export GITLAB_TOKEN="your-token"
./goanalysis gitanalysis --project=123 --mr=45
```

## 📂 项目结构

```
├── api/           # API定义
├── cmd/           # CLI命令
├── internal/      # 核心逻辑
├── web/           # 前端文件
└── configs/       # 配置文件
```

## 🏗️ 部署

### Docker
```bash
docker run -p 8000:8000 -p 9000:9000 \
  ghcr.io/toheart/goanalysis:latest
```

### 构建
```bash
make package-linux
make package-windows
```

## 🔧 故障排除

| 问题 | 解决方案 |
|------|----------|
| 端口被占用 | `lsof -i :8000; kill -9 <PID>` |
| 数据库锁定 | `rm -f goanalysis.db-*` |
| 前端缺失 | `make sync-frontend` |

## 🤝 贡献指南

1. Fork仓库
2. 创建功能分支
3. 提交更改和测试
4. 提交Pull Request

遵循 [约定式提交](https://www.conventionalcommits.org/zh-hans/)。

## 📜 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| v1.1.4 | 2024-12-16 | GitLab集成 |
| v1.1.0 | 2024-12-01 | Vue3升级 |
| v1.0.0 | 2024-11-15 | 首个稳定版 |

## 📄 许可证

MIT许可证 - 查看 [LICENSE](LICENSE) 文件。

## 📞 支持

- **🐛 问题**: [GitHub Issues](https://github.com/toheart/goanalysis/issues)
- **💬 讨论**: [GitHub Discussions](https://github.com/toheart/goanalysis/discussions)
- **📖 文档**: [Wiki](https://github.com/toheart/goanalysis/wiki)
- **📱 微信**: 关注"小唐的技术日志"获取更新

<div align="center">
  <h4>📱 关注微信公众号</h4>
  <p><strong>小唐的技术日志</strong></p>
  <img src="docs/images/wechat-qr.jpg" alt="微信公众号二维码" width="200"/>
  <p><i>扫描获取最新资讯</i></p>
</div>

---

<div align="center">
  <p><strong>GoAnalysis</strong> - 为Go开发者赋能</p>
  <p>⭐ 在GitHub上给我们一个Star！</p>
</div>