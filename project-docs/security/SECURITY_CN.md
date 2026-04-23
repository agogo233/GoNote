# 安全指南 - GoNote

本文档提供了在生产环境中部署和运行 GoNote 的安全建议。

---

## 🔐 快速安全检查清单

在将 GoNote 暴露到互联网之前，请完成以下检查：

- [ ] **修改默认密码**：将 `admin` 改为强密码
- [ ] **生成随机密钥**：用于会话加密
- [ ] **启用认证**（`authentication.enabled: true`）
- [ ] **启用限流**（`rate_limit.enabled: true`）
- [ ] **配置 CORS**：指定允许的源（不要用 `*`）
- [ ] **启用安全 Cookie**：如果使用 HTTPS（`secure_cookie: true`）
- [ ] **更新 Go 版本**到最新稳定版（最低 1.24.13+）

---

## 🚨 关键安全设置

### 1. 认证

**默认（不安全）：**
```yaml
authentication:
  enabled: false
  password: "admin"
  secret_key: "change_this_to_a_random_secret_key_in_production"
```

**生产环境（安全）：**
```yaml
authentication:
  enabled: true
  password: "YourStrongPassword123!"  # 修改这个！
  secret_key: "a3f8b2c1d4e5f6789012345678901234567890abcdef"  # 生成新的！
  session_max_age: 604800  # 7 天
  secure_cookie: true
```

**生成安全密钥：**
```bash
# 使用 OpenSSL
openssl rand -hex 32

# 使用 Python
python3 -c "import secrets; print(secrets.token_hex(32))"
```

### 2. 限流

**默认（本地开发关闭）：**
```yaml
rate_limit:
  enabled: false
  max_requests: 30
  window_seconds: 1
```

**生产环境（启用）：**
```yaml
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
```

限流保护防止：
- 登录暴力破解
- API 滥用
- 拒绝服务攻击（DoS）

### 3. CORS 配置

**默认（宽松）：**
```yaml
server:
  allowed_origins: ["*"]
```

**生产环境（严格）：**
```yaml
server:
  allowed_origins:
    - "https://yourdomain.com"
    - "https://www.yourdomain.com"
```

### 4. 安全 Cookie

使用 HTTPS 时（生产环境推荐）：

```yaml
authentication:
  secure_cookie: true
```

**自动检测：** 应用在以下情况下会自动启用安全 Cookie：
- 设置了 `HTTPS=true` 环境变量
- `X_FORWARDED_PROTO=https`（反向代理场景）
- `allowed_origins` 包含 HTTPS URL

---

## 🛡️ 安全特性

### 内置保护

| 特性 | 说明 | 状态 |
|---------|-------------|--------|
| **CSRF 保护** | Double Submit Cookie 模式，使用 `X-CSRF-Token` 头 | ✅ 已启用 |
| **路径遍历防护** | `ValidatePathSecurity()` 验证所有文件路径 | ✅ 已启用 |
| **会话安全** | HTTPOnly Cookie，SameSite=Lax，可配置 Secure | ✅ 已启用 |
| **文件上传验证** | 大小限制、MIME 类型检查、原子写入 | ✅ 已启用 |
| **密码哈希** | bcrypt，成本因子 12 | ✅ 已启用 |
| **优雅关闭** | 正确清理 goroutine 和连接 | ✅ 已启用 |

### 文件上传安全

```yaml
upload:
  max_file_size_mb: 50       # 根据需求调整
  max_body_size_mb: 100
  allowed_types: []          # 空 = 允许所有，或指定 MIME 类型
```

**生产环境推荐限制：**
```yaml
upload:
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
    - application/pdf
```

---

## ⚠️ 已知漏洞

### Go 标准库（截至 Go 1.24）

GoNote 继承了 Go 1.24 标准库中的 15 个漏洞。这些**不是代码 bug**，而是 Go 自身的已知问题。

**Go 1.25.8 修复的关键漏洞：**
- GO-2026-4602：`os` 包中的 FileInfo 转义
- GO-2026-4601：`net/url` 中的 IPv6 解析
- GO-2026-4340：TLS 握手加密
- GO-2026-4337：TLS 会话恢复

**建议：**
- 生产环境：使用 **Go 1.24.13+**（修复 11 个漏洞）
- 最大安全性：等待 **Go 1.25.8**（修复全部）

**缓解措施：** 应用的 `ValidatePathSecurity()` 函数提供了额外的路径遍历攻击保护，降低了某些漏洞的影响。

### 应用层修复

以下问题已在代码库中修复：

| 问题 | 文件 | 状态 |
|-------|------|--------|
| 无效正则 `(?!` | `internal/services/statistics.go` | ✅ 已修复 |
| 已弃用的 `strings.Title()` | `internal/services/theme.go` | ✅ 已修复 |
| 未使用的函数 | `internal/services/backlink.go` | ✅ 已移除 |
| 冗余的 bool 比较 | `internal/services/backlink.go` | ✅ 已修复 |

---

## 🔒 部署场景

### 本地开发（默认配置）

```yaml
authentication:
  enabled: false
rate_limit:
  enabled: false
server:
  allowed_origins: ["*"]
```

**风险等级：** ✅ 仅 localhost 安全

### 自托管（家庭网络）

```yaml
authentication:
  enabled: true
  password: "ChangeMe123!"
  secret_key: "generate_random_key"
rate_limit:
  enabled: false  # 可信网络可选
server:
  allowed_origins: ["http://192.168.1.100:9000"]
```

**风险等级：** ✅ 低（可信网络）

### 生产环境（公开互联网）

```yaml
authentication:
  enabled: true
  password: "StrongPassword123!"
  secret_key: "random_64_char_hex_string"
  secure_cookie: true
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
server:
  allowed_origins:
    - "https://yourdomain.com"
  debug: false
upload:
  max_file_size_mb: 10
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
```

**风险等级：** ⚠️ 需要启用所有安全措施

---

## 🚀 Docker 部署安全

### 环境变量（推荐）

```bash
# 认证
AUTHENTICATION_ENABLED=true
AUTHENTICATION_PASSWORD=YourStrongPassword123!
AUTHENTICATION_SECRET_KEY=$(openssl rand -hex 32)
AUTHENTICATION_SECURE_COOKIE=true

# 限流
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=30

# 服务器
DEBUG=false
ALLOWED_ORIGINS=https://yourdomain.com

# HTTPS 检测（反向代理后面）
X_FORWARDED_PROTO=https
```

### Docker Compose 示例

```yaml
version: '3.8'
services:
  gonote:
    image: gonote/gonote:latest
    environment:
      - AUTHENTICATION_ENABLED=true
      - AUTHENTICATION_PASSWORD=${AUTH_PASSWORD:-ChangeMe123!}
      - AUTHENTICATION_SECRET_KEY=${AUTH_SECRET_KEY:-change_me}
      - RATE_LIMIT_ENABLED=true
      - DEBUG=false
    volumes:
      - ./data:/app/data
      - ./config.yaml:/app/config.yaml
    ports:
      - "9000:9000"
    restart: unless-stopped
```

---

## 📋 安全审计命令

运行这些命令验证安装：

```bash
# 检查漏洞
cd go && go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 运行静态分析
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# 运行竞态检测测试
go test -race ./...

# 构建并验证
go build ./...
```

---

## 🆘 安全事件响应

如果怀疑安全泄露：

1. **立即修改密码：**
   ```bash
   # 更新 config.yaml
   password: "NewStrongPassword!"
   ```

2. **重新生成密钥：**
   ```bash
   openssl rand -hex 32
   ```

3. **查看日志：**
   ```bash
   # 检查应用日志
   docker logs gonote
   ```

4. **更新到最新版本：**
   ```bash
   git pull origin main
   docker-compose pull
   docker-compose up -d
   ```

---

## 📚 额外资源

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go 安全最佳实践](https://go.dev/doc/security/)
- [Fiber 安全](https://docs.gofiber.io/security/)

---

## 🤝 报告安全问题

如果你发现安全漏洞，请通过在 GitHub 上私密提交 issue 来报告，并标记为机密，或直接联系维护者。

**请勿**在安全问题得到解决之前公开披露。
