# 🔐 安全指南 / Security Guide

本文档提供在生产环境中部署和运行 GoNote 的安全建议与最佳实践。

---

## 🔐 快速安全检查清单 / Quick Security Checklist

在将 GoNote 暴露到互联网之前，请完成以下检查：

- [ ] **修改默认密码**：将 `admin` 改为强密码（至少 12 位，含大小写字母、数字、符号）
- [ ] **生成随机密钥**：使用 `openssl rand -hex 32` 生成会话加密密钥
- [ ] **启用认证**（`authentication.enabled: true`）
- [ ] **启用限流**（`rate_limit.enabled: true`）
- [ ] **配置 CORS**：指定允许的源（不要用 `["*"]`）
- [ ] **启用安全 Cookie**：HTTPS 环境下设置 `secure_cookie: true`
- [ ] **更新 Go 版本**到最新稳定版（最低 1.24.13+）

---

## 🚨 关键安全设置 / Critical Security Settings

### 1. 认证配置 / Authentication

**默认配置（不安全）：**

```yaml
authentication:
  enabled: false
  password: "admin"
  secret_key: "change_this_to_a_random_secret_key_in_production"
```

**生产环境配置（安全）：**

```yaml
authentication:
  enabled: true
  password: "YourStrongPassword123!"  # 修改为强密码
  secret_key: "a3f8b2c1d4e5f6789012345678901234567890abcdef"  # 生成新的随机密钥
  session_max_age: 604800  # 7 天（单位：秒）
  secure_cookie: true
```

**生成安全密钥：**

```bash
# 使用 OpenSSL（推荐）
openssl rand -hex 32

# 或使用 Go
go run -e 'package main; import "crypto/rand"; import "fmt"; import "encoding/hex"; func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'
```

---

### 2. 限流配置 / Rate Limiting

**默认配置（本地开发关闭）：**

```yaml
rate_limit:
  enabled: false
  max_requests: 30
  window_seconds: 1
```

**生产环境配置（启用）：**

```yaml
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
```

**限流保护范围：**

- ✅ 防止登录暴力破解
- ✅ 防止 API 滥用
- ✅ 减轻拒绝服务攻击（DoS）影响

---

### 3. CORS 配置 / CORS Configuration

**默认配置（宽松）：**

```yaml
server:
  allowed_origins: ["*"]
```

**生产环境配置（严格）：**

```yaml
server:
  allowed_origins:
    - "https://yourdomain.com"
    - "https://www.yourdomain.com"
```

**CORS 安全提示：**

- `["*"]` 仅适用于私有网络自托管
- 公网部署必须指定具体域名
- 配合认证使用时，可防止 CSRF 攻击

---

### 4. 安全 Cookie / Secure Cookies

使用 HTTPS 时（生产环境推荐）启用：

```yaml
authentication:
  secure_cookie: true
```

**自动检测机制：**

应用会在以下情况自动启用安全 Cookie：

| 检测条件 (Condition) | 环境变量 / 配置 | 触发值 (Trigger Value) |
|---------------------|-----------------|------------------------|
| HTTPS 标志 | `HTTPS` 环境变量 | `true`、`1` 或 `on` |
| 反向代理 | `X_FORWARDED_PROTO` | `https` |
| 允许的源 | `allowed_origins` 配置 | 包含 `https://` URL |

---

## 🛡️ 内置安全特性 / Built-in Security Features

### 已启用的保护机制

| 特性 (Feature) | 说明 (Description) | 状态 (Status) |
|----------------|-------------------|---------------|
| **CSRF 防护** | Double Submit Cookie 模式，使用 `X-CSRF-Token` 请求头 | ✅ 已启用 |
| **路径遍历防护** | `ValidatePathSecurity()` 函数验证所有文件路径 | ✅ 已启用 |
| **会话安全** | HTTPOnly Cookie、SameSite=Lax、可配置 Secure | ✅ 已启用 |
| **文件上传验证** | 大小限制、MIME 类型检查、原子写入 | ✅ 已启用 |
| **密码哈希** | bcrypt 算法，成本因子 12 | ✅ 已启用 |
| **优雅关闭** | 正确清理 goroutine 和网络连接 | ✅ 已启用 |

---

### 文件上传安全配置

```yaml
upload:
  max_file_size_mb: 50       # 根据需求调整
  max_body_size_mb: 100
  allowed_types: []          # 空列表 = 允许所有类型
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

**文件上传安全要点：**

- 限制文件大小防止磁盘空间耗尽
- 限制 MIME 类型避免恶意文件上传
- 文件名自动重命名为时间戳，避免路径注入
- 媒体文件存储在 `_attachments` 文件夹，隔离于笔记内容

---

## ⚠️ 已知漏洞与应对 / Known Vulnerabilities

### Go 标准库漏洞（Go 1.24）

GoNote 继承了 Go 1.24 标准库中的若干漏洞。这些**不是代码缺陷**，而是 Go 自身已知问题。

**Go 1.25.8 修复的关键漏洞：**

- GO-2026-4602：`os` 包中的 FileInfo 转义问题
- GO-2026-4601：`net/url` 中的 IPv6 解析问题
- GO-2026-4340：TLS 握手加密问题
- GO-2026-4337：TLS 会话恢复问题

**版本建议：**

- **生产环境**：使用 Go 1.24.13+（修复 11 个漏洞）
- **最高安全性**：升级到 Go 1.25.8+（修复全部 15 个漏洞）

**缓解措施：**  
应用的 `ValidatePathSecurity()` 函数提供了额外的路径遍历攻击防护，降低了某些漏洞的影响范围。

---

### 代码层修复记录

以下问题已在代码库中修复：

| 问题描述 | 文件位置 | 状态 |
|---------|---------|------|
| 无效正则表达式 `(?!` | `internal/services/statistics.go` | ✅ 已修复 |
| 已弃用的 `strings.Title()` | `internal/services/theme.go` | ✅ 已修复 |
| 未使用的函数 | `internal/services/backlink.go` | ✅ 已移除 |
| 冗余的布尔比较 | `internal/services/backlink.go` | ✅ 已修复 |

---

## 🔒 部署场景配置 / Deployment Scenarios

### 场景 1：本地开发（默认配置）

```yaml
authentication:
  enabled: false
rate_limit:
  enabled: false
server:
  allowed_origins: ["*"]
```

**风险等级：** ✅ 仅 localhost 安全  
**适用场景：** 个人电脑开发测试

---

### 场景 2：自托管（家庭网络）

```yaml
authentication:
  enabled: true
  password: "ChangeMe123!"  # 建议使用更复杂的密码
  secret_key: "generate_random_key_here"
rate_limit:
  enabled: false  # 可信网络可选禁用
server:
  allowed_origins: ["http://192.168.1.100:9000"]
```

**风险等级：** ✅ 低风险（可信内网）  
**适用场景：** 家庭 NAS、本地服务器

---

### 场景 3：生产环境（公网暴露）

```yaml
authentication:
  enabled: true
  password: "StrongPassword123!"  # 强密码（12+位，混合字符）
  secret_key: "random_64_char_hex_string"
  secure_cookie: true
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
server:
  allowed_origins:
    - "https://yourdomain.com"
  debug: false  # 生产环境禁用调试
upload:
  max_file_size_mb: 10  # 根据实际需求调整
  allowed_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
```

**风险等级：** ⚠️ 需启用所有安全措施  
**适用场景：** 公网 VPS、云服务器、PaaS 平台

---

## 🐳 Docker 部署安全 / Docker Security

### 环境变量配置（推荐方式）

```bash
# 认证配置
AUTHENTICATION_ENABLED=true
AUTHENTICATION_PASSWORD=YourStrongPassword123!
AUTHENTICATION_SECRET_KEY=$(openssl rand -hex 32)
AUTHENTICATION_SECURE_COOKIE=true

# 限流配置
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=30

# 服务器配置
DEBUG=false
ALLOWED_ORIGINS=https://yourdomain.com

# HTTPS 检测（如果使用反向代理）
X_FORWARDED_PROTO=https
```

---

### Docker Compose 安全示例

```yaml
version: '3.8'
services:
  gonote:
    image: ghcr.io/gamosoft/gonote:latest
    container_name: gonote
    restart: unless-stopped
    environment:
      - AUTHENTICATION_ENABLED=true
      - AUTHENTICATION_PASSWORD=${AUTH_PASSWORD}
      - AUTHENTICATION_SECRET_KEY=${AUTH_SECRET_KEY}
      - AUTHENTICATION_SECURE_COOKIE=true
      - RATE_LIMIT_ENABLED=true
      - DEBUG=false
      - ALLOWED_ORIGINS=https://yourdomain.com
    volumes:
      - ./data:/app/data:rw  # 数据持久化
      - ./config.yaml:/app/config.yaml:ro  # 配置文件（只读）
    ports:
      - "9000:9000"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

**安全强化建议：**

- 使用非 root 用户运行容器（`user: "1000:1000"`）
- 配置文件挂载为只读（`:ro`）
- 启用健康检查
- 设置资源限制（CPU、内存）

---

## 📋 安全审计命令 / Security Audit Commands

运行以下命令验证安装的安全性：

```bash
# 1. 检查已知漏洞
cd go && go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 2. 运行静态代码分析
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# 3. 运行竞态检测测试
go test -race ./...

# 4. 验证构建
go build ./...
```

**预期结果：**

- `govulncheck`：无发现漏洞（或仅有标准库已知问题）
- `staticcheck`：无严重警告
- `go test -race`：所有测试通过，无竞态条件

---

## 🆘 安全事件响应 / Security Incident Response

如果怀疑发生安全泄露，按以下步骤处理：

### 紧急响应步骤

1. **立即修改密码**

```bash
# 更新 config.yaml 或设置环境变量
AUTHENTICATION_PASSWORD=NewStrongPassword123!
```

2. **重新生成密钥**

```bash
openssl rand -hex 32
# 更新 AUTHENTICATION_SECRET_KEY
```

3. **查看日志排查**

```bash
# Docker 部署
docker logs gonote --tail 100

# 直接运行
tail -f logs/app.log
```

4. **更新到最新版本**

```bash
git pull origin main
# 或更新 Docker 镜像
docker-compose pull
docker-compose up -d
```

5. **检查异常活动**

- 查看 `/data/cache/` 中的会话文件
- 检查是否有未授权的媒体文件
- 核对笔记修改历史

---

## 📚 额外资源 / Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/) - Web 应用安全风险 Top 10
- [Go 安全最佳实践](https://go.dev/doc/security/) - 官方 Go 安全指南
- [Fiber 框架安全](https://docs.gofiber.io/security/) - 框架层安全特性
- [Docker 安全](https://docs.docker.com/engine/security/) - 容器安全最佳实践

---

## 🤝 报告安全问题 / Reporting Security Issues

如果您发现安全漏洞，请通过以下方式**私密报告**：

1. 在 GitHub 上创建 Issue 并标记为 **Confidential**
2. 或直接联系项目维护者（联系方式见 README）

**请勿**在安全问题解决前公开披露，以免造成用户损失。

---

**文档版本**：v1.0  
**最后更新**：2025 年 1 月  
**适用版本**：GoNote v1.0+
