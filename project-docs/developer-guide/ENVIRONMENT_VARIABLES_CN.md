# 🔧 环境变量

GoNote 支持通过环境变量覆盖配置设置，使应用在不同部署环境（本地、测试、生产）下可以有不同的行为。

## 📋 可用环境变量

### 核心设置

| 变量 | 类型 | 默认值 | 说明 |
|----------|------|---------|-------------|
| `PORT` | 整数 | `9000` | HTTP 端口（Docker、Go 后端） |

> **注意**：高级服务器设置（CORS 源、调试模式）仅通过 `config.yaml` 配置，不支持环境变量。详见 [config.yaml](#高级服务器配置)。

### 认证

| 变量 | 类型 | 默认值 | 说明 |
|----------|------|---------|-------------|
| `AUTHENTICATION_ENABLED` | 布尔 | `config.yaml` | 启用/禁用认证 |
| `AUTHENTICATION_PASSWORD` | 字符串 | `admin` | 明文密码（启动时自动哈希） |
| `AUTHENTICATION_PASSWORD_HASH` | 字符串 | - | 预哈希的 bcrypt 密码（高级用户） |
| `AUTHENTICATION_SECRET_KEY` | 字符串 | `config.yaml` | 会话密钥（用于会话安全） |
| `AUTHENTICATION_SECURE_COOKIE` | 布尔 | `false` | 强制启用安全 Cookie（未设置时自动检测） |

> **密码优先级：** `AUTHENTICATION_PASSWORD` 优先于 `AUTHENTICATION_PASSWORD_HASH`。如果同时设置，将使用明文密码。

#### 🔒 安全 Cookie 的 HTTPS 自动检测

当 `secure_cookie` 未显式设置为 `true` 时，系统会自动检测 HTTPS 并启用安全 Cookie：

| 检测方式 | 环境变量 | 触发值 |
|-----------------|---------------------|------------------|
| HTTPS 标志 | `HTTPS` | `true`、`1` 或 `on` |
| 反向代理 | `X_FORWARDED_PROTO` | `https` |
| 允许的源 | 配置文件中的 `allowed_origins` | 包含 `https://` URL |

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

> **安全提示：** 安全 Cookie 仅通过 HTTPS 连接发送。如果你在使用反向代理终止 SSL，请确保代理转发了正确的协议头。

#### 示例：通过环境变量设置密码

```bash
# Docker
docker run -e AUTHENTICATION_ENABLED=true -e AUTHENTICATION_PASSWORD=mysecretpassword ...

# Docker Compose（在 .env 文件或 docker-compose.yml 中）
AUTHENTICATION_PASSWORD=mysecretpassword
```

### 演示模式

| 变量 | 类型 | 默认值 | 说明 |
|----------|------|---------|-------------|
| `DEMO_MODE` | 布尔 | `false` | 启用演示模式（启用限流和其他演示限制） |

### 支持

| 变量 | 类型 | 默认值 | 说明 |
|----------|------|---------|-------------|
| `ALREADY_DONATED` | 布尔 | `false` | 隐藏设置面板中的赞助按钮 |

> ⚠️ **免责声明：** 没有验证机制。但传说如果不打赏就设置这个为 `true`，你的下一次 `git push` 会悄无声息地失败。就一次，在最关键的时候。
>
> 还没打赏？[☕ 请我喝杯咖啡](https://ko-fi.com/gamosoft) - 只需 30 秒，就能让我开心一整天！

## 🎯 配置优先级

配置按以下顺序加载（后面的覆盖前面的）：

1. **`config.yaml`** - 默认配置文件
2. **环境变量** - 运行时覆盖
3. **命令行参数** - 最高优先级（如适用）

## 🔧 高级服务器配置

以下设置仅在 `config.yaml` 中可用（不支持环境变量）：

### CORS（跨域资源共享）

```yaml
server:
  # CORS 允许的源列表
  # 默认：["*"] 允许所有源（自托管使用没问题）
  # 生产环境：请指定具体域名
  allowed_origins: ["*"]

  # 生产环境示例：
  # allowed_origins: ["http://localhost:9000", "https://yourdomain.com"]
  # allowed_origins: ["https://*.yourdomain.com"]  # 通配符子域名
```

**安全提示：**
- `["*"]` 对于**私有网络自托管**部署是**安全**的
- 对于**公开部署**，请指定具体源，防止未授权 API 访问
- 当启用认证时，这可以防止 CSRF 攻击

### 调试模式

```yaml
server:
  # 启用详细的错误信息（API 响应中包含完整堆栈）
  # 默认：false（生产安全）
  # 开发/排查问题时设为 true
  debug: false
```

**⚠️ 重要**：生产环境切勿启用 `debug: true`！

当 `debug: true` 时：
- 返回完整的错误堆栈信息
- 暴露内部路径和系统细节
- 可能泄露安全漏洞

当 `debug: false`（推荐）：
- 返回通用错误信息
- 完整错误仅记录在服务器日志中
- 生产安全的错误处理

---

## 📚 相关文档

- **认证**：[AUTHENTICATION.md](AUTHENTICATION.md)
- **API 限流**：[API.md](API.md#限流)

---

**专业提示：** 使用环境变量管理**部署相关**的设置，使用 `config.yaml` 管理**应用默认**设置。这样可以让配置保持灵活且易于维护！🎯
