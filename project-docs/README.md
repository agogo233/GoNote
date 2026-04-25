# GoNote 项目文档

本目录包含 GoNote 项目的完整文档。

This directory contains comprehensive documentation for the GoNote project.

## 📚 文档分类 / Documentation Categories

### 👤 用户指南 / User Guide
面向最终用户的文档，涵盖功能和使用说明。

Documentation for end users covering features and usage.

| 文档 / Document | 描述 / Description |
|----------|-------------|
| [FEATURES.md](user-guide/FEATURES.md) / [中文版](user-guide/FEATURES_CN.md) | 完整功能列表和键盘快捷键 |
| [THEMES.md](user-guide/THEMES.md) / [中文版](user-guide/THEMES_CN.md) | 主题自定义和创建自定义主题 |
| [TAGS.md](user-guide/TAGS.md) / [中文版](user-guide/TAGS_CN.md) | 使用标签和组合过滤组织笔记 |
| [TEMPLATES.md](user-guide/TEMPLATES.md) / [中文版](user-guide/TEMPLATES_CN.md) | 使用可重用模板创建笔记 |
| [MATHJAX.md](user-guide/MATHJAX.md) / [中文版](user-guide/MATHJAX_CN.md) | LaTeX/MathJax 示例和语法参考 |
| [MERMAID.md](user-guide/MERMAID.md) / [中文版](user-guide/MERMAID_CN.md) | 使用 Mermaid 创建图表 |
| [SHARING.md](user-guide/SHARING.md) / [中文版](user-guide/SHARING_CN.md) | 使用分享令牌分享笔记 |

### 🔧 开发者指南 / Developer Guide
面向开发者的技术文档。

Technical documentation for developers and contributors.

| 文档 / Document | 描述 / Description |
|----------|-------------|
| [API.md](developer-guide/API.md) / [API_CN.md](developer-guide/API_CN.md) | REST API 完整文档（含认证、限流、新端点等） |
| [ENVIRONMENT_VARIABLES.md](developer-guide/ENVIRONMENT_VARIABLES.md) / [ENVIRONMENT_VARIABLES_CN.md](developer-guide/ENVIRONMENT_VARIABLES_CN.md) | 所有环境变量配置参考 |
| [ASSETS.md](developer-guide/ASSETS.md) / [ASSETS_CN.md](developer-guide/ASSETS_CN.md) | 共享资源目录结构说明 |
| [BUILD.md](developer-guide/BUILD.md) / [BUILD_CN.md](developer-guide/BUILD_CN.md) | Tailwind CSS 构建配置 |
| [DEPLOY.md](developer-guide/DEPLOY.md) / [DEPLOY_CN.md](developer-guide/DEPLOY_CN.md) | 部署配置指南（Render、Docker、本地） |
| [TESTING.md](developer-guide/TESTING.md) / [TESTING_CN.md](developer-guide/TESTING_CN.md) | 测试套件说明（E2E + Go单元测试） |

### 🔒 安全 / Security
安全指南和最佳实践。

Security guides and best practices.

| 文档 / Document | 描述 / Description |
|----------|-------------|
| [SECURITY.md](security/SECURITY.md) | 安全指南和最佳实践（英文） |
| [SECURITY_CN.md](security/SECURITY_CN.md) | 安全指南（中文） |
| [AUTHENTICATION.md](security/AUTHENTICATION.md) | 认证设置和配置详解 |
| [AUTHENTICATION_CN.md](security/AUTHENTICATION_CN.md) | 认证说明（中文） |

### 📋 项目 / Project
项目级文档。

Project-level documentation.

| 文档 / Document | 描述 / Description |
|----------|-------------|
| [CONTRIBUTING.md](../../CONTRIBUTING.md) | 贡献指南（代码、文档、翻译） |
| [CHANGELOG.md](../../CHANGELOG.md) | 版本历史和发布说明 |

### 📝 模板 / Templates
笔记模板文件。

Template files for notes.

| 目录 / Directory | 描述 / Description |
|-----------|-------------|
| [templates/](templates/) | 预设笔记模板（日记、会议、项目规划） |

---

## 🌐 额外文档 / Additional Documentation

### 根目录文档 / Root Documentation
- [README.md](../../README.md) - 项目主入口（双语）
- [AGENTS.md](../../AGENTS.md) - OpenCode 代理开发指南
- [IMPLEMENTATION_SUMMARY.md](../../IMPLEMENTATION_SUMMARY.md) - 路径统一重构总结

### 网站资源 / Website Assets
- [docs/](../../docs/) - 网站文档和营销资源（icons、screenshot）

### Docker 相关
- [docker/README.md](../../docker/README.md) - Docker 使用说明
- [docker-compose.ghcr.yml](../../docker-compose.ghcr.yml) - 生产环境 Docker Compose
- [docker/compose/development.yml](../../docker/compose/development.yml) - 开发环境 Docker Compose
- [docker/compose/production.yml](../../docker/compose/production.yml) - 生产环境 Docker Compose（推荐）

### 部署配置
- [deploy/render.yaml](../../deploy/render.yaml) - Render.com 部署配置

---

## 📖 快速链接 / Quick Links

- **[快速开始](../../README.md#quick-start)** - 安装和设置 / Installation and setup
- **[功能列表](user-guide/FEATURES.md)** - GoNote 能做什么 / What GoNote can do
- **[主题定制](user-guide/THEMES.md)** - 自定义界面外观
- **[API 文档](developer-guide/API.md)** - REST API 完整参考
- **[环境变量](developer-guide/ENVIRONMENT_VARIABLES.md)** - 所有配置覆盖方式
- **[安全指南](security/SECURITY.md)** - 生产环境安全最佳实践
- **[贡献指南](../../CONTRIBUTING.md)** - 如何参与开发 / How to contribute
- **[版本历史](../../CHANGELOG.md)** - 变更日志和发布说明

---

## 🤝 贡献文档 / Contributing to Documentation

如果发现错误或者想改进文档，请：
1. 查看[贡献指南](../../CONTRIBUTING.md)
2. 针对重大更改先提交 issue 讨论
3. 提交包含改进的拉取请求

If you find errors or want to improve documentation, please:
1. Check the [Contributing Guidelines](../../CONTRIBUTING.md)
2. Open an issue for discussion (for major changes)
3. Submit a pull request with your improvements

---

**欢迎贡献翻译！** 所有用户指南和安全文档均有中英文版本。开发者指南正在逐步完善中文翻译。如需协助，请参考贡献指南。

**Last Updated:** 2026-04-24 (v0.25.0 - 完整文档同步更新)
