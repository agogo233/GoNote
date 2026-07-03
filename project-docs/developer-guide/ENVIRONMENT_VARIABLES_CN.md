# 🔧 环境变量配置 / Environment Variables

GoNote 支持通过环境变量覆盖配置文件中的设置，使您能够在不同部署环境（本地、预发布、生产）中灵活配置应用行为。

---

## 📋 可用环境变量总览 / Available Environment Variables

### 应用配置 / Application

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `LOG_ENABLED` | 布尔 (boolean) | `true` | 启用/禁用所有日志输出 |
| `CONFIG_PATH` | 字符串 (string) | `go/config.yaml` | 配置文件路径（覆盖默认位置） |

---

### 服务器配置 / Server

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `HOST` | 字符串 (string) | `0.0.0.0` | 服务器绑定地址 |
| `PORT` | 整数 (integer) | `9000` | HTTP 端口 |
| `ALLOWED_ORIGINS` | 逗号分隔字符串 (comma-separated string) | `*` | CORS 允许的源（列表） |
| `DEBUG` | 布尔 (boolean) | `false` | 启用调试模式（生产环境不推荐） |

**ALLOWED_ORIGINS 示例：**

```bash
# 允许多个源（用逗号分隔）
export ALLOWED_ORIGINS="http://localhost:8000,https://yourdomain.com"

# 使用通配符子域名
ALLOWED_ORIGINS="https://*.yourdomain.com"
```

---

### 存储配置 / Storage

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `STORAGE_NOTES_DIR` | 字符串 (string) | `./data/notes` | 笔记数据目录路径（相对路径基于工作目录） |

**注意：** 路径应为相对路径（如 `./data/notes`），从项目根目录解析。绝对路径也可以，但会降低可移植性。

---

### 搜索配置 / Search

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `SEARCH_ENABLED` | 布尔 (boolean) | `true` | 启用/禁用搜索功能 |

---

### 认证配置 / Authentication

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `AUTHENTICATION_ENABLED` | 布尔 (boolean) | `false` | 启用/禁用密码认证 |
| `AUTHENTICATION_PASSWORD` | 字符串 (string) | `admin` | 明文密码（启动时自动哈希处理） |
| `AUTHENTICATION_PASSWORD_HASH` | 字符串 (string) | 无 | 预哈希的 bcrypt 密码（高级用户） |
| `AUTHENTICATION_SECRET_KEY` | 字符串 (string) | 配置文件中的值 | 会话加密密钥 |
| `AUTHENTICATION_SECURE_COOKIE` | 布尔 (boolean) | `false` | 强制启用安全 Cookie |
| `AUTHENTICATION_SESSION_MAX_AGE` | 整数 (integer) | `604800` | 会话有效期（秒），默认 7 天 |

> **密码优先级说明**：如果同时设置 `AUTHENTICATION_PASSWORD` 和 `AUTHENTICATION_PASSWORD_HASH`，明文密码优先。环境变量中的密码会覆盖配置文件中的设置。

---

### 缓存配置 / Cache

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `CACHE_TTL` | 整数 (integer) | `60` | 缓存过期时间（秒） |
| `CACHE_CAPACITY` | 整数 (integer) | `1000` | 缓存最大条目数 |
| `CACHE_SCAN_INTERVAL` | 整数 (integer) | `30` | 后台扫描间隔（秒） |

---

### 限流配置 / Rate Limiting

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `RATE_LIMIT_ENABLED` | 布尔 (boolean) | `false` | 启用全局限流 |
| `RATE_LIMIT_MAX` | 整数 (integer) | `30` | 每个时间窗口的最大请求数 |
| `RATE_LIMIT_WINDOW` | 整数 (integer) | `1` | 时间窗口时长（秒） |

---

### 文件上传配置 / Upload

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `UPLOAD_MAX_FILE_SIZE_MB` | 整数 (integer) | `50` | 单个文件最大大小（MB） |
| `UPLOAD_MAX_BODY_SIZE_MB` | 整数 (integer) | `100` | 请求体最大大小（MB） |
| `UPLOAD_ALLOWED_TYPES` | 逗号分隔字符串 (comma-separated string) | (所有类型) | 允许的 MIME 类型列表 |

**支持的 MIME 类型：**

- 图片 (Images)：`image/jpeg`、`image/png`、`image/gif`、`image/webp`
- 音频 (Audio)：`audio/mpeg`、`audio/wav`、`audio/ogg`、`audio/mp4`
- 视频 (Video)：`video/mp4`、`video/webm`、`video/quicktime`、`video/x-msvideo`
- 文档 (Documents)：`application/pdf`

**使用示例：**

```bash
# 仅允许图片上传
export UPLOAD_ALLOWED_TYPES="image/jpeg,image/png,image/gif,image/webp"
```

---

### 演示模式 / Demo Mode

| 环境变量 (Variable) | 类型 (Type) | 默认值 (Default) | 说明 (Description) |
|---------------------|-------------|------------------|-------------------|
| `DEMO_MODE` | 布尔 (boolean) | `false` | 启用演示模式（启用限流和其他演示限制） |
| `ALREADY_DONATED` | 布尔 (boolean) | `false` | 在设置面板中隐藏支持按钮 |

> ⚠️ **免责声明**：`ALREADY_DONATED` 标志没有验证机制。传说中的说法是，如果设置为 `true` 而未实际捐赠，下一次 `git push` 可能会静默失败。仅一次，在最关键的时候。  
> 如果您喜欢这个项目，欢迎 [☕ 请我喝杯咖啡](https://ko-fi.com/gamosoft) - 只需 30 秒，就能让我开心一整天！

---

## 🔒 HTTPS 自动检测和安全 Cookie / HTTPS Auto-Detection & Secure Cookies

当未显式设置 `secure_cookie: true` 时，系统会自动检测 HTTPS 并启用安全 Cookie：

| 检测方式 (Method) | 环境变量 (Env Variable) | 触发值 (Trigger Value) |
|-------------------|-------------------------|------------------------|
| HTTPS 标志 | `HTTPS` | `true`、`1` 或 `on` |
| 反向代理协议 | `X_FORWARDED_PROTO` | `https` |
| 允许的源 | `allowed_origins` 配置 | 包含 `https://` URL |

**示例：PaaS 部署（Render、Railway 等）**

```bash
# 大多数 PaaS 平台会自动设置 HTTPS=true
# 无需额外配置！
```

**示例：反向代理（Nginx、Caddy 等）**

```bash
# 设置转发的协议头
export X_FORWARDED_PROTO=https
```

**示例：Docker Compose**

```yaml
environment:
  - X_FORWARDED_PROTO=https
```

> **安全提示**：安全 Cookie 仅通过 HTTPS 连接发送。如果您使用终止 SSL 的反向代理，请确保代理正确转发了 `X-Forwarded-Proto` 协议头。

---

## 🎯 配置优先级 / Configuration Priority

配置按以下顺序加载（后加载的覆盖先前的）：

1. **`config.yaml`** - 默认配置文件
2. **命令行参数** - 使用 `--config` 指定配置文件路径
3. **环境变量** - 运行时覆盖
4. **内置默认值** - 最低优先级

> **提示**：环境变量会覆盖 `config.yaml` 中的对应设置。这意味着您可以保留 config.yaml 中的默认值，通过环境变量在部署时动态调整配置。

---

## 📂 配置文件结构 / Configuration File Structure

虽然大部分配置可通过环境变量设置，但某些复杂结构或文件路径仍需在 `config.yaml` 中配置。以下是完整的配置结构说明，并标注了哪些字段可以被环境变量覆盖。

### 应用配置 / Application

```yaml
app:
  name: "GoNote"          # 应用名称（无对应的环境变量）
  # 版本号从 VERSION 文件自动读取
```

### 日志配置 / Logging

```yaml
log:
  enabled: true            # 可被环境变量覆盖：LOG_ENABLED
```

### 服务器配置 / Server

```yaml
server:
  host: "0.0.0.0"          # 可被环境变量覆盖：HOST
  port: 9000               # 可被环境变量覆盖：PORT
  allowed_origins: ["*"]   # 可被环境变量覆盖：ALLOWED_ORIGINS（逗号分隔）
  debug: false             # 可被环境变量覆盖：DEBUG
```

### 存储配置 / Storage

```yaml
storage:
  notes_dir: "./data/notes"     # 可被环境变量覆盖：STORAGE_NOTES_DIR
```

### 搜索配置 / Search

```yaml
search:
  enabled: true            # 可被环境变量覆盖：SEARCH_ENABLED
  # 索引缓存目录基于 notes_dir 自动生成: ./data/cache/search
```

### 认证配置 / Authentication

```yaml
authentication:
  enabled: false           # 可被环境变量覆盖：AUTHENTICATION_ENABLED
  secret_key: "change_this_to_a_random_secret_key_in_production"
                          # 可被环境变量覆盖：AUTHENTICATION_SECRET_KEY
  password: "admin"        # 可被环境变量覆盖：AUTHENTICATION_PASSWORD（明文）
  password_hash: ""        # 可被环境变量覆盖：AUTHENTICATION_PASSWORD_HASH（预哈希）
  session_max_age: 604800  # 可被环境变量覆盖：AUTHENTICATION_SESSION_MAX_AGE（秒）
  secure_cookie: false     # 可被环境变量覆盖：AUTHENTICATION_SECURE_COOKIE（自动 HTTPS 检测）
```

### 缓存配置 / Cache

```yaml
cache:
  ttl: 60                  # 可被环境变量覆盖：CACHE_TTL（秒）
  capacity: 1000           # 可被环境变量覆盖：CACHE_CAPACITY
  scan_interval: 30        # 可被环境变量覆盖：CACHE_SCAN_INTERVAL（秒）
```

### 限流配置 / Rate Limiting

```yaml
rate_limit:
  enabled: false           # 可被环境变量覆盖：RATE_LIMIT_ENABLED
  max_requests: 30         # 可被环境变量覆盖：RATE_LIMIT_MAX
  window_seconds: 1        # 可被环境变量覆盖：RATE_LIMIT_WINDOW
```

### 文件上传配置 / Upload

```yaml
upload:
  max_file_size_mb: 50     # 可被环境变量覆盖：UPLOAD_MAX_FILE_SIZE_MB
  max_body_size_mb: 100    # 可被环境变量覆盖：UPLOAD_MAX_BODY_SIZE_MB
  allowed_types: []        # 可被环境变量覆盖：UPLOAD_ALLOWED_TYPES（空 = 允许所有）
  # 支持的 MIME 类型示例：
  # - image/jpeg, image/png, image/gif, image/webp
  # - audio/mpeg, audio/wav, audio/ogg, audio/mp4
  # - video/mp4, video/webm, video/quicktime, video/x-msvideo
  # - application/pdf
```

---

## 📚 相关文档 / Related Documentation

- **认证配置**：[../security/AUTHENTICATION_CN.md](../security/AUTHENTICATION_CN.md)
- **API 文档**：[API_CN.md](API_CN.md#rate-limiting-限流)
- **配置文件**：`go/config.yaml`（项目根目录）

---

## 💡 实用建议 / Practical Recommendations

**开发环境**：
- 使用 `config.yaml` 本地配置，避免环境变量过多
- 启用 `DEBUG=true` 提升开发体验

**生产环境**：
- 使用环境变量或 Docker secrets 管理敏感信息（如 `AUTHENTICATION_SECRET_KEY`、`AUTHENTICATION_PASSWORD`）
- 启用 `AUTHENTICATION_ENABLED=true` 并设置强密码
- 设置 `ALLOWED_ORIGINS` 为具体域名，避免 `["*"]`
- 生产环境勿启用 `DEBUG=true`

**Docker 部署**：
- 在 `docker-compose.yml` 的 `environment` 部分设置环境变量
- 敏感信息使用 Docker secrets 或外部环境变量文件

**PaaS 平台**：
- 使用平台的"环境变量"配置面板（如 Render、Railway、Heroku）
- PaaS 通常会自动设置 `HTTPS=true`，安全 Cookie 会自动启用

**安全性**：
- 永远不要将敏感配置提交到版本控制
- 使用 `.env` 文件（添加到 `.gitignore`）管理本地开发配置
- 生产环境使用平台提供的密钥管理功能

---

## 🎯 最佳实践 / Best Practices

**简单记忆原则**：

- 用环境变量管理**部署相关**的设置（端口、密码、密钥等）
- 用 `config.yaml` 管理**应用默认**设置和复杂结构
- 这样可以让配置保持灵活且易于维护

**配置检查清单**（部署前）：

- [ ] 已更改默认密码 `admin` 为强密码
- [ ] 已生成随机 `AUTHENTICATION_SECRET_KEY`
- [ ] 如果启用认证，`AUTHENTICATION_SECURE_COOKIE` 已在 HTTPS 环境下正确设置
- [ ] `ALLOWED_ORIGINS` 已限制为具体域名（生产环境）
- [ ] `DEBUG` 已设置为 `false`（生产环境）
- [ ] 敏感配置未提交到版本控制

---

**最后更新**：2025 年 1 月  
**适用版本**：GoNote v1.0+
