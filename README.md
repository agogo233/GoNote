# GoNote

[![GitHub Stars](https://img.shields.io/github/stars/gamosoft/gonote?style=flat)](https://github.com/gamosoft/gonote)
[![Build](https://img.shields.io/github/actions/workflow/status/gamosoft/gonote/docker-publish.yml)](https://github.com/gamosoft/gonote/actions)
[![Latest Version](https://img.shields.io/github/v/tag/gamosoft/gonote)](https://github.com/gamosoft/gonote/releases)
[![License](https://img.shields.io/github/license/gamosoft/gonote)](LICENSE)

> Your Self-Hosted Knowledge Base
>
> [📖 中文文档](#中文文档) | [📚 Documentation](#documentation)

---

## What is GoNote?

GoNote is a **lightweight, self-hosted note-taking application** that puts you in complete control of your knowledge base. Write, organize, and discover your notes with a beautiful, modern interface—all running on your own server.

**GoNote** 是一个**轻量级、自托管的笔记应用**，让您完全掌控自己的知识库。使用美观、现代的界面编写、组织和发现笔记——全部运行在您自己的服务器上。

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Backend** | Go 1.24+ with Fiber |
| **Frontend** | Vanilla JS + Alpine.js |
| **Storage** | Plain Markdown files |

### 技术栈

| 组件 | 技术 |
|------|------|
| **后端** | Go 1.24+ with Fiber |
| **前端** | Vanilla JS + Alpine.js |
| **存储** | 纯 Markdown 文件 |

---

## Who is it for?

- **Privacy-conscious users** who want complete control over their data
- **Developers** who prefer markdown and local file storage
- **Knowledge workers** building a personal wiki or second brain
- **Teams** looking for a self-hosted alternative to commercial apps
- **Anyone** who values simplicity, speed, and ownership

### 适合谁用？

- **注重隐私的用户**——完全掌控自己的数据
- **开发者**——偏好 Markdown 和本地文件存储
- **知识工作者**——构建个人维基或第二大脑
- **团队**——寻找自托管替代商业应用
- **任何人**——看重简单、快速和所有权

---

## Quick Start

### Docker (推荐)

这是最快上手的方式，5 分钟即可运行。

#### 快速设置

**Linux / macOS:**

```bash
mkdir -p gonote/data && cd gonote
docker run -d --name gonote -p 9000:9000 \
  -v $(pwd)/data:/app/data \
  ghcr.io/gamosoft/gonote:latest
```

**Windows (PowerShell):**

```powershell
mkdir gonote\data; cd gonote
docker run -d --name gonote -p 9000:9000 `
  -v ${PWD}/data:/app/data `
  ghcr.io/gamosoft/gonote:latest
```

打开 **http://localhost:9000** 即可访问。

> 💡 **数据持久化**：您的笔记保存在宿主机的 `./data/` 目录。重启容器后数据不会丢失。
> 
> - 宿主机路径：`./data/notes/`（项目根目录下的 data 目录）
> - 容器内路径：`/app/data/notes/`
> - 映射关系：`-v $(pwd)/data:/app/data`

---

#### 使用 Docker Compose

提供两个 docker-compose 配置文件：

| 文件 | 用途 |
|------|------|
| `docker/compose/production.yml` | **推荐** - 使用 GitHub Container Registry 的预构建镜像 |
| `docker/compose/development.yml` | 开发用 - 从本地源码构建 |

**选项 1：预构建镜像（最快）**

```bash
mkdir -p gonote/data && cd gonote
curl -O https://raw.githubusercontent.com/gamosoft/gonote/main/docker/compose/production.yml
docker-compose -f docker/compose/production.yml up -d
```

**选项 2：从源码构建（开发）**

```bash
git clone https://github.com/gamosoft/gonote.git
cd gonote
docker-compose -f docker/compose/development.yml up -d
```

---

### 本地运行（无需 Docker）

适合开发环境或偏好直接运行。

**要求：**
- Go 1.24 或更高版本

```bash
# 克隆仓库
git clone https://github.com/gamosoft/gonote.git
cd gonote

# 从项目根运行应用（重要！）
go run go/cmd/server/main.go --config go/config.yaml

# 访问 http://localhost:9000
```

> ⚠️ **重要提示**：必须从**项目根目录**运行，不要进入 `go/` 目录。数据目录 `./data/` 是相对于工作目录的。

#### 从源码构建

```bash
cd gonote/go

# 下载依赖
go mod download

# 构建二进制文件
go build -o gonote ./cmd/server

# 从项目根目录运行
cd ..
./go/gonote --config go/config.yaml
```

---

### Advanced Docker Setup

The Docker image includes bundled configuration, themes, and locales. To customize, you need to:

1. **Map volumes** in your docker-compose or docker run command
2. **Provide content** — files/folders must exist with valid content (empty = app might break!)

#### Volume Configuration

| Volume | Purpose | Bundled? |
|--------|---------|----------|
| `data/` | Your notes (must create) | No |
| `config.yaml` | App settings | Yes |
| `shared/themes/` | Built-in themes | Yes |
| `shared/locales/` | Built-in translations | Yes |

```yaml
# docker-compose.yml 示例
version: '3'
services:
  gonote:
    image: ghcr.io/gamosoft/gonote:latest
    ports:
      - "9000:9000"
    volumes:
      - ./data:/app/data              # 您的笔记（必须）
      - ./config.yaml:/app/config.yaml  # 自定义配置（可选）
      - ./shared/themes:/app/shared/themes  # 自定义主题（可选）
      - ./locales:/app/locales        # 自定义语言包（可选）
```

---

### Dashboard Integration

<a href="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons@master/svg/gonote.svg" target="_blank">
  <img src="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons@master/svg/gonote.svg" alt="GoNote Icon" width="64" height="64">
</a>

An official icon for GoNote is now available on [Dashboard Icons](https://dashboardicons.com/icons/gonote)! Use it in your self-hosted dashboards like Homepage, Homarr, Dashy, Heimdall, etc...

#### 仪表盘集成

<a href="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons@master/svg/gonote.svg" target="_blank">
  <img src="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons@master/svg/gonote.svg" alt="GoNote Icon" width="64" height="64">
</a>

GoNote 官方图标现已在 [Dashboard Icons](https://dashboardicons.com/icons/gonote) 上线！可用于 Homepage、Homarr、Dashy、Heimdall 等自托管仪表盘。

---

## Documentation

Want to learn more? Check out our comprehensive documentation.

### English Documentation

- **[FEATURES.md](project-docs/user-guide/FEATURES.md)** — Complete feature list and keyboard shortcuts
- **[THEMES.md](project-docs/user-guide/THEMES.md)** — Theme customization and creating custom themes
- **[TAGS.md](project-docs/user-guide/TAGS.md)** — Organize notes with tags and combined filtering
- **[TEMPLATES.md](project-docs/user-guide/TEMPLATES.md)** — Create notes from reusable templates with dynamic placeholders
- **[MATHJAX.md](project-docs/user-guide/MATHJAX.md)** — LaTeX/math notation examples and syntax reference
- **[MERMAID.md](project-docs/user-guide/MERMAID.md)** — Diagram creation with Mermaid (flowcharts, sequence diagrams, and more)
- **[SHARING.md](project-docs/user-guide/SHARING.md)** — Share notes with tokens

- **[API.md](project-docs/developer-guide/API.md)** — REST API documentation and examples
- **[ENVIRONMENT_VARIABLES.md](project-docs/developer-guide/ENVIRONMENT_VARIABLES.md)** — Configure settings via environment variables
- **[AUTHENTICATION.md](project-docs/security/AUTHENTICATION.md)** — Enable password protection for your instance
- **[SECURITY.md](project-docs/security/SECURITY.md)** — Security guide and best practices
- **[TESTING.md](project-docs/developer-guide/TESTING.md)** — How to run tests and contribute tests

---

### 中文文档

- **[功能列表](project-docs/user-guide/FEATURES_CN.md)** — 完整功能列表和键盘快捷键
- **[主题定制](project-docs/user-guide/THEMES_CN.md)** — 主题自定义和创建自定义主题
- **[标签系统](project-docs/user-guide/TAGS_CN.md)** — 使用标签和组合过滤组织笔记
- **[笔记模板](project-docs/user-guide/TEMPLATES_CN.md)** — 使用可重用模板创建笔记
- **[数学公式](project-docs/user-guide/MATHJAX_CN.md)** — LaTeX/MathJax 示例和语法参考
- **[Mermaid 图表](project-docs/user-guide/MERMAID_CN.md)** — 使用 Mermaid 创建图表
- **[笔记分享](project-docs/user-guide/SHARING_CN.md)** — 使用分享令牌分享笔记

- **[API 文档](project-docs/developer-guide/API_CN.md)** — REST API 文档和示例
- **[环境变量](project-docs/developer-guide/ENVIRONMENT_VARIABLES_CN.md)** — 通过环境变量配置设置
- **[Docker 部署](project-docs/developer-guide/DOCKER_DEPLOYMENT_CN.md)** — Docker 部署完整指南（路径映射、故障排查）
- **[认证说明](project-docs/security/AUTHENTICATION_CN.md)** — 启用实例的密码保护
- **[安全指南](project-docs/security/SECURITY_CN.md)** — 安全指南和最佳实践
- **[测试指南](project-docs/developer-guide/TESTING_CN.md)** — 如何运行测试和贡献测试

---

## Multi-Language Support

GoNote supports multiple languages! Currently available:

- **English (en-US)** — Default language
- **Simplified Chinese (zh-CN)** — 简体中文

**To change language:** Go to Settings (gear icon) → Language dropdown.

**To add your own language:** See the [Contributing Guidelines](CONTRIBUTING.md#translating-documentation) for instructions on creating translation files.

**Docker users:** Mount your custom locales folder to add or override translations:

```yaml
volumes:
  - ./locales:/app/locales  # Custom translations
```

**Pro Tip:** If you clone this repository, you can mount the `project-docs/` folder to view these docs inside the app:

```yaml
# In your docker-compose.yml
volumes:
  - ./data:/app/data              # Your personal notes
  - ./project-docs:/app/data/docs:ro  # Mount docs inside data folder (read-only)
```

Then access them at `http://localhost:9000` — the docs will appear as a `docs/` folder in the file browser!

---

### 多语言支持

GoNote 支持多种语言！当前可用：

- **英语 (en-US)** — 默认语言
- **简体中文 (zh-CN)** — 中文界面

**切换语言：** 前往设置（齿轮图标）→ 语言下拉菜单。

**添加您自己的语言：** 请参阅 [贡献指南](CONTRIBUTING.md#translating-documentation) 获取创建翻译文件的说明。

**Docker 用户：** 挂载自定义语言文件夹以添加或覆盖翻译：

```yaml
volumes:
  - ./locales:/app/locales  # 自定义翻译
```

**小贴士：** 如果您克隆此仓库，可以挂载 `project-docs/` 文件夹在应用内查看文档：

```yaml
# 在您的 docker-compose.yml 中
volumes:
  - ./data:/app/data              # 您的个人笔记
  - ./project-docs:/app/data/docs:ro  # 挂载文档到数据文件夹（只读）
```

然后在 `http://localhost:9000` 访问——文档会以 `docs/` 文件夹形式出现在文件浏览器中！

---

## Security

GoNote is designed for **self-hosted, private use**. It's perfect for running on your local machine or home network.

### 🚨 **Before Exposing to the Internet**

If you plan to access GoNote from outside your local network, **complete this security checklist**:

1. ✅ **Change default password** — from `admin` to a strong, unique password
2. ✅ **Generate a random secret key** — for session encryption in `config.yaml`
3. ✅ **Enable authentication** — set `authentication.enabled: true`
4. ✅ **Enable rate limiting** — set `rate_limit.enabled: true`
5. ✅ **Configure CORS** — specify allowed origins (never use `*` in production)
6. ✅ **Use HTTPS** — run behind reverse proxy (nginx/Caddy) with SSL/TLS
7. ✅ **Enable secure cookies** — set `secure_cookie: true` when using HTTPS
8. ✅ **Update default secret_key** — generate a new random value

See **[SECURITY.md](project-docs/security/SECURITY.md)** for complete security guide.

---

### 安全说明

GoNote 专为**自托管、私密使用**而设计。适合在本地机器或家庭网络上运行。

#### 🚨 **暴露到互联网之前**

如果您计划从本地网络外部访问 GoNote，**请完成以下安全配置**：

1. ✅ **修改默认密码** — 从 `admin` 改为强密码
2. ✅ **生成随机密钥** — 用于会话加密（在 `config.yaml` 中配置）
3. ✅ **启用认证** — 设置 `authentication.enabled: true`
4. ✅ **启用限流** — 设置 `rate_limit.enabled: true`
5. ✅ **配置 CORS** — 指定允许的源（生产环境切勿使用 `*`）
6. ✅ **使用 HTTPS** — 使用反向代理（nginx/Caddy）配置 SSL/TLS
7. ✅ **启用安全 Cookie** — 使用 HTTPS 时设置 `secure_cookie: true`
8. ✅ **更新默认密钥** — 生成新的随机值替换默认值

完整安全指南请参阅 **[SECURITY_CN.md](project-docs/security/SECURITY_CN.md)**。

---

### Network Security Recommendation

- **Do NOT expose directly to the internet** without additional security measures
- **Use reverse proxy** (nginx, Caddy, Traefik) with HTTPS for production
- **Keep on local network** or use VPN for remote access
- By default, the app listens on `0.0.0.0:9000` (all network interfaces)

### 网络安全建议

- **不要在没有额外安全措施的情况下直接暴露到互联网**
- **使用反向代理**（nginx、Caddy、Traefik）并启用 HTTPS
- **保持在本地网络** 或使用 VPN 进行远程访问
- 默认情况下，应用监听 `0.0.0.0:9000`（所有网络接口）

---

## Why GoNote?

### vs. Commercial Apps (Notion, Evernote, Obsidian Sync)

| Feature | GoNote | Commercial Apps |
|---------|------------|-----------------|
| **Cost** | 100% Free | 💰 $xxx/month/year |
| **Privacy** | 🔒 Your server, your data | ☁️ Their servers, their terms |
| **Speed** | ⚡ Lightning fast | 📶 Depends on internet |
| **Offline** | ✅ Always works | ⚠️ Limited or requires sync |
| **Customization** | 🔧 Full control | 🔒 Limited options |
| **No Lock-in** | 📄 Plain markdown files | 🔒 Proprietary formats |

---

### Key Benefits

- **Total Privacy** — Your notes never leave your server
- **Optional Authentication** — Simple password protection for self-hosted deployments
- **Zero Cost** — No subscriptions, no hidden fees
- **Fast & Lightweight** — Instant search and navigation
- **Beautiful Themes** — Multiple themes, easy to customize

- **Responsive** — Works on desktop, tablet, and mobile
- **Simple Storage** — Plain markdown files in folders
- **Math Support** — LaTeX/MathJax for beautiful equations
- **HTML Export** — Share notes as standalone HTML files
- **Graph View** — Interactive visualization of connected notes
- **Favorites** — Star your most-used notes for instant access
- **Outline Panel** — Navigate headings with click-to-jump TOC

---

### 为什么选择 GoNote？

#### 对比商业应用（Notion、Evernote、Obsidian Sync）

| 特性 | GoNote | 商业应用 |
|------|--------|----------|
| **成本** | 100% 免费 | 💰 每月/每年 $xxx |
| **隐私** | 🔒 您的服务器，您的数据 | ☁️ 他们的服务器，他们的条款 |
| **速度** | ⚡ 闪电般快速 | 📶 取决于网络 |
| **离线** | ✅ 始终可用 | ⚠️ 有限或需同步 |
| **定制** | 🔧 完全控制 | 🔒 有限选项 |
| **无锁定** | 📄 纯 Markdown 文件 | 🔒 专有格式 |

---

### 主要优势

- **完全隐私**——您的笔记永远不会离开您的服务器
- **可选认证**——简单的密码保护，适合自托管部署
- **零成本**——无订阅费，无隐藏费用
- **快速轻量**——即时搜索和导航
- **精美主题**——多种主题，易于定制

- **响应式**——适用于桌面、平板和手机
- **简单存储**——文件夹中的纯 Markdown 文件
- **数学支持**——LaTeX/MathJax 渲染精美公式
- **HTML 导出**——将笔记分享为独立 HTML 文件
- **图谱视图**——交互式可视化笔记关联
- **收藏夹**——星标常用笔记，快速访问
- **大纲面板**——点击跳转导航标题

---

## Contributing

We welcome contributions! Before submitting a pull request:

1. Read our [Contributing Guidelines](CONTRIBUTING.md)
2. Open an issue first to discuss major features or significant changes
3. Ensure your code follows the project's style and philosophy

See [DEVELOPMENT.md](project-docs/developer-guide/DEVELOPMENT_CN.md) for detailed development setup.

### 贡献指南

欢迎贡献！提交拉取请求之前：

1. 阅读我们的 [贡献指南](CONTRIBUTING.md)
2. 先提交 issue 讨论主要功能或重大更改
3. 确保您的代码符合项目的风格与理念

详细开发环境设置请参阅 [开发指南](project-docs/developer-guide/DEVELOPMENT_CN.md)。

---

## License

MIT License — Free to use, modify, and distribute. See [LICENSE](LICENSE) for details.

## 许可证

MIT 许可证 — 可自由使用、修改和分发。详情请参阅 [LICENSE](LICENSE) 文件。

---

<p align="center">
  Made with ❤️ for the self-hosting community<br>
  倾情打造，为自托管社区服务
</p>
