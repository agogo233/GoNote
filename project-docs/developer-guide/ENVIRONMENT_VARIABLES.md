# 🔧 环境变量配置

GoNote 支持通过环境变量覆盖配置文件中的设置，使您能够在不同部署环境（本地、预发布、生产）中灵活配置应用行为。

---

## 📋 可用环境变量总览

### 应用配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `LOG_ENABLED` | boolean | `true` | 启用/禁用所有日志输出 |

### 服务器配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `HOST` | string | `0.0.0.0` | 服务器绑定地址 |
| `PORT` | integer | `9000` | HTTP 端口 |
| `ALLOWED_ORIGINS` | 逗号分隔的字符串 | `*` | CORS 允许的源（列表） |
| `DEBUG` | boolean | `false` | 启用调试模式（生产环境不推荐） |
| `RELOAD` | boolean | `false` | 启用文件热重载（仅开发环境） |

> **ALLOWED_ORIGINS 示例**：
> ```bash
> # 允许多个源
> export ALLOWED_ORIGINS="http://localhost:8000,https://yourdomain.com"
> # 或使用逗号分隔
> ALLOWED_ORIGINS="http://localhost:8000,https://*.yourdomain.com"
> ```

### 存储配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `STORAGE_NOTES_DIR` | string | `./data/notes` | 笔记数据目录路径（相对路径基于工作目录） |

### 搜索配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `SEARCH_ENABLED` | boolean | `true` | 启用/禁用搜索功能 |

### 认证配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `AUTHENTICATION_ENABLED` | boolean | `false` | 启用/禁用密码认证 |
| `AUTHENTICATION_PASSWORD` | string | `admin` | 明文密码（启动时自动哈希处理） |
| `AUTHENTICATION_PASSWORD_HASH` | string | - | 预哈希的 bcrypt 密码（高级用户） |
| `AUTHENTICATION_SECRET_KEY` | string | 配置文件中的值 | 会话加密密钥 |
| `AUTHENTICATION_SECURE_COOKIE` | boolean | `false` | 强制启用安全 Cookie |
| `AUTHENTICATION_SESSION_MAX_AGE` | integer | `604800` | 会话有效期（秒），默认 7 天 |

> **密码优先级**：如果同时设置 `AUTHENTICATION_PASSWORD` 和 `AUTHENTICATION_PASSWORD_HASH`，明文密码优先。

### 缓存配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `CACHE_TTL` | integer | `60` | 缓存过期时间（秒） |
| `CACHE_CAPACITY` | integer | `1000` | 缓存最大条目数 |
| `CACHE_SCAN_INTERVAL` | integer | `30` | 后台扫描间隔（秒） |

### 限流配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `RATE_LIMIT_ENABLED` | boolean | `false` | 启用全局限流 |
| `RATE_LIMIT_MAX` | integer | `30` | 每个时间窗口的最大请求数 |
| `RATE_LIMIT_WINDOW` | integer | `1` | 时间窗口时长（秒） |

### 文件上传配置

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `UPLOAD_MAX_FILE_SIZE_MB` | integer | `50` | 单个文件最大大小（MB） |
| `UPLOAD_MAX_BODY_SIZE_MB` | integer | `100` | 请求体最大大小（MB） |
| `UPLOAD_ALLOWED_TYPES` | 逗号分隔的字符串 | (所有类型) | 允许的 MIME 类型列表 |

> **支持的 MIME 类型**：
> - 图片：`image/jpeg`, `image/png`, `image/gif`, `image/webp`
> - 音频：`audio/mpeg`, `audio/wav`, `audio/ogg`, `audio/mp4`
> - 视频：`video/mp4`, `video/webm`, `video/quicktime`, `video/x-msvideo`
> - 文档：`application/pdf`

### 演示模式

| 变量 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `DEMO_MODE` | boolean | `false` | 启用演示模式（启用限流和其他演示限制） |
| `ALREADY_DONATED` | boolean | `false` | 在设置面板中隐藏支持按钮 |

> ⚠️ **免责声明**：此标志无验证。传说设置为 `true` 而未捐赠会导致下一次 `git push` 静默失败。仅一次。在最关键的时候。捐赠？[☕ 请我喝咖啡](https://ko-fi.com/gamosoft) - 只需 30 秒，让我开心一整天！

---

## 🔒 HTTPS 自动检测和安全 Cookie

当未显式设置 `secure_cookie: true` 时，系统会自动检测 HTTPS 并启用安全 Cookie：

| 检测方式 | 环境变量 | 触发值 |
|----------|----------|--------|
| HTTPS 标志 | `HTTPS` | `true`, `1`, 或 `on` |
| 反向代理 | `X_FORWARDED_PROTO` | `https` |
| 允许的源 | `allowed_origins` 配置 | 包含 `https://` URL |

**示例：PaaS 部署（Render、Railway 等）**
```bash
# 大多数 PaaS 平台会自动设置 HTTPS=true
# 无需额外配置！
```

**示例：反向代理（Nginx、Caddy 等）**
```bash
# 设置转发的协议
export X_FORWARDED_PROTO=https
```

**示例：Docker Compose**
```yaml
environment:
  - X_FORWARDED_PROTO=https
```

> **安全注意**：安全 Cookie 仅通过 HTTPS 连接发送。如果您使用终止 SSL 的反向代理，请确保代理转发正确的协议头。

---

## 🎯 配置优先级

配置按以下顺序加载（后设置的覆盖先前的）：

1. **`config.yaml`** - 默认配置文件
2. **命令行参数** - 使用 `--config` 指定配置文件路径
3. **环境变量** - 运行时覆盖
4. **内置默认值** - 最低优先级（配置结构中的零值或 applyDefaults 设置的默认值）

> **提示**：环境变量会覆盖 `config.yaml` 中的对应设置。这意味着您可以保留 config.yaml 中的默认值，通过环境变量在部署时动态调整配置。

---

## 📂 配置文件说明

虽然大部分配置可通过环境变量设置，但某些复杂结构或文件路径仍需在 `config.yaml` 中配置。以下是完整的配置结构说明：

### 应用配置

```yaml
app:
  name: "GoNote"          # 应用名称（无可覆盖的环境变量）
  # 版本号从 VERSION 文件自动读取
```

### 日志配置

```yaml
log:
  enabled: true            # 覆盖变量：LOG_ENABLED
  # 日志目录（无环境变量，仅配置文件）
  log_dir: "./logs"
```

### 服务器配置

```yaml
server:
  host: "0.0.0.0"          # 覆盖变量：HOST
  port: 9000               # 覆盖变量：PORT
  allowed_origins: ["*"]   # 覆盖变量：ALLOWED_ORIGINS（逗号分隔）
  debug: false             # 覆盖变量：DEBUG
  reload: false            # 覆盖变量：RELOAD（热重载）
```

### 存储配置

```yaml
storage:
  notes_dir: "./data/notes"     # 覆盖变量：STORAGE_NOTES_DIR
  # 以下路径目前仅支持 config.yaml 配置
  # 没有对应的环境变量
  cache_dir: "./data/cache"
  temp_dir: "./data/temp"
  backup_dir: "./backups"
```

### 搜索配置

```yaml
search:
  enabled: true            # 覆盖变量：SEARCH_ENABLED
  # 索引缓存目录基于 notes_dir 自动生成: ./data/cache/search
```

### 认证配置

```yaml
authentication:
  enabled: false           # 覆盖变量：AUTHENTICATION_ENABLED
  secret_key: "change_this_to_a_random_secret_key_in_production"  # 覆盖变量：AUTHENTICATION_SECRET_KEY
  password: "admin"        # 覆盖变量：AUTHENTICATION_PASSWORD（明文）
  password_hash: ""        # 覆盖变量：AUTHENTICATION_PASSWORD_HASH（预哈希）
  session_max_age: 604800  # 覆盖变量：AUTHENTICATION_SESSION_MAX_AGE（秒）
  secure_cookie: false     # 覆盖变量：AUTHENTICATION_SECURE_COOKIE（自动 HTTPS 检测）
```

### 缓存配置

```yaml
cache:
  ttl: 60                  # 覆盖变量：CACHE_TTL（秒）
  capacity: 1000           # 覆盖变量：CACHE_CAPACITY
  scan_interval: 30        # 覆盖变量：CACHE_SCAN_INTERVAL（秒）
```

### 限流配置

```yaml
rate_limit:
  enabled: false           # 覆盖变量：RATE_LIMIT_ENABLED
  max_requests: 30         # 覆盖变量：RATE_LIMIT_MAX
  window_seconds: 1        # 覆盖变量：RATE_LIMIT_WINDOW
```

### 文件上传配置

```yaml
upload:
  max_file_size_mb: 50     # 覆盖变量：UPLOAD_MAX_FILE_SIZE_MB
  max_body_size_mb: 100    # 覆盖变量：UPLOAD_MAX_BODY_SIZE_MB
  allowed_types: []        # 覆盖变量：UPLOAD_ALLOWED_TYPES（空 = 允许所有）
  # 支持的 MIME 类型示例：
  # - image/jpeg, image/png, image/gif, image/webp
  # - audio/mpeg, audio/wav, audio/ogg, audio/mp4
  # - video/mp4, video/webm, video/quicktime, video/x-msvideo
  # - application/pdf
```

---

## 📚 相关文档

- **认证配置**：[../security/AUTHENTICATION.md](../security/AUTHENTICATION.md)
- **API 文档**：[API.md](API.md)（包括Rate Limiting说明）
- **配置文件**：`config.yaml`（位于项目根目录的 `go/config.yaml`）

---

## 💡 实用建议

- **开发环境**：使用 `config.yaml` 本地配置，避免污染环境变量
- **生产环境**：使用环境变量或 Docker secrets 管理敏感信息（如 `AUTHENTICATION_SECRET_KEY`、`AUTHENTICATION_PASSWORD`）
- **Docker 部署**：在 `docker-compose.yml` 的 `environment` 部分设置环境变量
- **PaaS 平台**：使用平台的"环境变量"配置面板（如 Render、Railway、Heroku）
- **安全性**：永远不要将敏感配置提交到版本控制，使用 `.env` 文件（添加到 `.gitignore`）或平台的密钥管理功能

---

**🎯 最佳实践**：使用环境变量管理**部署相关**的设置，使用 `config.yaml` 管理**应用默认**设置。这样可以让配置保持灵活且易于维护！
