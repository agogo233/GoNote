# 🔒 GoNote 认证指南

## ⚠️ 默认密码警告

> **默认密码是 `admin`** — 在暴露到任何网络之前，务必修改！

---

## 概述

GoNote 为单用户部署提供了简单的密码保护功能。启用后，用户必须先登录才能访问笔记。

- ✅ 适用于单用户 / 自托管
- ✅ 使用 bcrypt 哈希密码
- ✅ 基于会话（默认 7 天，可配置）

---

## 快速测试（仅限本地）

本地测试时，认证**默认关闭**。要启用认证测试：

1. 在 `config.yaml` 中设置 `authentication.enabled: true`
2. 重启应用
3. 使用密码 `admin` 登录

⚠️ **不要在任何网络环境中使用默认密码！**

---

## 生产环境设置

对于任何暴露到网络的部署，请按照以下步骤操作：

### 第一步：生成密钥

密钥用于加密会话 Cookie。生成一个随机密钥：

```bash
# 使用 openssl（推荐）
openssl rand -hex 32

# 使用 Docker
docker exec -it gonote sh -c 'openssl rand -hex 32'
```

**保存这个密钥** — 第二步需要用到。

---

### 第二步：配置认证

选择**以下一种**方式：

#### 方案 A：明文密码（推荐）

最简单的方式。你的密码会在启动时自动哈希。

**通过环境变量（Docker）：**
```bash
docker run -d \
  -e AUTHENTICATION_ENABLED=true \
  -e AUTHENTICATION_PASSWORD=你的安全密码 \
  -e AUTHENTICATION_SECRET_KEY=你生成的密钥 \
  ...
```

**通过 config.yaml：**
```yaml
authentication:
  enabled: true
  password: "你的安全密码"
  secret_key: "你生成的密钥"
```

---

#### 方案 B：预哈希密码（高级）

适合自己手动哈希密码的用户。

**生成哈希值：**
```bash
# 使用 Go bcrypt
cd go && go run tools/hash_password.go your_password

# 使用 htpasswd（如可用）
htpasswd -bnBC 12 "" your_password | tr -d ':\n'
```

**然后配置：**
```yaml
authentication:
  enabled: true
  password_hash: "$2b$12$..."  # 粘贴你的哈希值
  secret_key: "你生成的密钥"
```

---

### 第三步：重启并测试

```bash
# Docker Compose
docker-compose restart

# Docker run
docker restart gonote

# 本地（Go 后端）
cd go && go run cmd/server/main.go
```

访问 `http://localhost:9000` — 你将被重定向到登录页面。

---

## 配置优先级

如果配置了多个来源，按以下优先级应用（第一个匹配生效）：

| 优先级 | 来源 | 类型 |
|----------|--------|------|
| 第一 | `AUTHENTICATION_PASSWORD` 环境变量 | 明文 |
| 第二 | `AUTHENTICATION_PASSWORD_HASH` 环境变量 | 预哈希 |
| 第三 | config.yaml 中的 `password` | 明文 |
| 第四 | config.yaml 中的 `password_hash` | 预哈希 |

**示例：** 如果设置了环境变量 `AUTHENTICATION_PASSWORD`，它会覆盖 config.yaml 中的任何配置。

---

## 安全注意事项

### ✅ 这能保护什么

- 未授权访问你的笔记
- 所有 API 端点
- 查看、创建、编辑、删除笔记

### ⚠️ 这不能保护什么

这是一个**简单的单用户**系统。**不适合**以下场景：

- ❌ 多用户环境
- ❌ 没有 HTTPS 的公开互联网
- ❌ 合规要求（HIPAA、GDPR 等）

### 🛡️ 最佳实践

1. **使用 HTTPS** — 通过反向代理运行（Traefik、Nginx、Caddy）
2. **强密码** — 至少 12 个字符，大小写混合，含数字和符号
3. **唯一密钥** — 永远不要在不同应用间复用
4. **保护配置文件** — 不要将凭据提交到版本控制

---

## 关闭认证

```yaml
authentication:
  enabled: false
```

重启应用以生效。
