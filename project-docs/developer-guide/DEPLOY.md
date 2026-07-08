# 🚀 部署配置 / Deployment Guide

本指南提供 GoNote 在不同平台上的部署配置说明，包括 Docker、Render.com、本地直接运行等多种方式。

---

## 📁 部署文件总览 / Deployment Files

| 文件 (File) | 平台 (Platform) | 用途 (Purpose) |
|-------------|-----------------|----------------|
| `render.yaml` | Render.com | Render 部署蓝图（Blueprint） |
| `docker-compose.ghcr.yml` | Docker Compose | 生产环境配置（使用 GHCR 预构建镜像） |
| `docker/compose/production.yml` | Docker Compose | 生产环境推荐配置 |
| `docker/compose/development.yml` | Docker Compose | 开发环境配置（从源码构建） |
| `docker/go/Dockerfile` | Docker | Go 后端 Dockerfile |

---

## 🏗️ Render.com 部署 / Deploy to Render.com

### 快速部署步骤

1. **Fork 仓库或连接到 Render.com**
2. **使用 `deploy/render.yaml` 作为部署蓝图**
3. **配置环境变量**（参考下方建议）

Render.yaml 预配置项：

- **服务类型**：Web Service (Docker)
- **镜像源**：GHCR（GitHub Container Registry）预构建镜像
- **健康检查**：`/health` 路径，每 10 秒检查一次
- **端口**：8000（Render 自动映射）
- **计划类型**：Free（可随时升级）

---

### 环境变量配置（生产环境）

**Render.yaml 示例配置**：

```yaml
envVars:
  - key: PORT
    value: 8000
  - key: DEMO_MODE
    value: "true"
  - key: AUTHENTICATION_ENABLED
    value: "true"
  - key: AUTHENTICATION_PASSWORD_HASH
    value: "$2b$12$..."  # 请替换为实际密码的 bcrypt 哈希
  - key: AUTHENTICATION_SECRET_KEY
    value: "4f36da5af76627301dcdc0347c4b111bdc6c86ae830444af852de935198c3210"
```

> ⚠️ **安全警告**：Render.yaml 中的 `AUTHENTICATION_SECRET_KEY` 和密码哈希是**公开的演示凭据**！生产部署必须使用 Render Dashboard 生成新的随机密钥，并标记为 **Secret**。

**推荐生产配置**（在 Render Dashboard 中设置）：

| 环境变量 (Variable) | 建议值 (Recommended) | 说明 (Description) |
|---------------------|----------------------|-------------------|
| `PORT` | `8000` | Render 默认端口 |
| `AUTHENTICATION_ENABLED` | `true` | 启用认证 |
| `AUTHENTICATION_PASSWORD` 或 `AUTHENTICATION_PASSWORD_HASH` | 强密码或哈希 | 在 Dashboard 中作为 Secret 设置 |
| `AUTHENTICATION_SECRET_KEY` | 随机字符串 | 使用 `openssl rand -hex 32` 生成，标记为 Secret |
| `AUTHENTICATION_SECURE_COOKIE` | `true` | Render 自动提供 HTTPS |
| `LOG_ENABLED` | `true` | 启用日志 |

---

### 部署流程

1. 登录 [Render Dashboard](https://dashboard.render.com)
2. 点击 **"New +"** → **"Blueprint"**
3. 连接您的 GitHub 仓库（授权 Render 访问）
4. Render 会自动检测并解析 `render.yaml`
5. 点击 **"Create Resources"** 创建所有服务
6. 在 Render Dashboard 中设置敏感环境变量（建议标记为 "Secret"）
7. 等待部署完成（首次构建约 5-10 分钟）
8. 访问分配的 URL（如 `https://gonote.onrender.com`）

---

### 手动触发部署

如果需要手动触发新版本部署：

1. 在 Render Dashboard 中找到您的服务
2. 点击服务详情页的 **"Manual Deploy"** 按钮
3. 选择 **"Deploy latest commit from main branch"**
4. 等待部署完成

---

## 🐳 Docker Compose 部署 / Docker Compose Deployment

提供了多个 Docker Compose 配置文件适配不同场景。

### 推荐方式：生产环境（预构建镜像）

使用 `docker/compose/production.yml`：

```bash
# 从项目根目录运行
mkdir -p data # 确保创建数据目录

docker-compose -f docker/compose/production.yml up -d
```

**配置特点**：

- ✅ 使用 GHCR 预构建镜像（`ghcr.io/gamosoft/gonote:latest`）
- ✅ 启动速度快（无需本地构建）
- ✅ 自动挂载 `./data` 到容器 `/app/data`
- ✅ 适合生产环境

**停止服务**：

```bash
docker-compose -f docker/compose/production.yml down
```

**查看日志**：

```bash
docker-compose -f docker/compose/production.yml logs -f
```

---

### 开发环境（从源码构建）

使用 `docker/compose/development.yml`：

```bash
docker-compose -f docker/compose/development.yml up -d
```

**配置特点**：

- ✅ 从本地源代码构建镜像
- ✅ 适合开发调试
- ✅ 文件更改可能实时反映（需配置卷挂载）

**注意事项**：

- 首次构建会下载依赖并编译，时间较长（约 2-5 分钟）
- 修改代码后需重建镜像：`docker-compose build`

---

### 📦 Docker 路径映射说明 / Docker Volume Mapping

**重要**：Docker 容器使用卷挂载来持久化数据。正确理解路径映射对数据管理至关重要。

#### 路径对应关系

```yaml
# docker-compose.yml 示例
volumes:
  - ./data:/app/data
```

| 容器内路径 | 宿主机路径（项目根目录） | 说明 |
|-----------|------------------------|------|
| `/app/data/notes/` | `./data/notes/` | **笔记数据**（核心） |
| `/app/data/cache/` | `./data/cache/` | 搜索索引和缓存 |
| `/app/data/temp/` | `./data/temp/` | 临时文件 |
| `/app/config.yaml` | 需手动挂载 | 配置文件（可选） |

#### 数据持久化验证

**方法 1：在宿主机查看（推荐）**
```bash
# 直接查看宿主机上的数据
ls -la ./data/notes/
cat ./data/notes/test.md
```

**方法 2：进入容器查看**
```bash
# 1. 找到容器
docker ps | grep gonote

# 2. 进入容器
docker exec -it gonote sh
# 或
docker exec -it gonote-go sh

# 3. 在容器内查看
ls -la /app/data/notes/
cat /app/data/notes/test.md
exit
```

**方法 3：使用 docker exec 直接执行**
```bash
# 不进入容器，直接执行命令
docker exec gonote-go ls -la /app/data/notes/
docker exec gonote-go find /app/data -name "*.md"
```

#### 常见问题：为什么我的笔记数据不见了？

**问题现象**：启动容器后，发现 `./data/notes/` 目录为空。

**可能原因**：

1. **容器未运行** - 数据是在容器运行时写入的
   ```bash
   # 检查容器状态
   docker ps | grep gonote
   
   # 启动容器
   docker-compose -f docker/compose/development.yml up -d
   ```

2. **挂载路径不正确** - 确认 `docker-compose.yml` 中的 `volumes` 配置
   ```yaml
   volumes:
     - ./data:/app/data  # 确保 ./data 在项目根目录
   ```

3. **工作目录错误** - 必须从**项目根目录**启动 Docker Compose
   ```bash
   # ❌ 错误：在 docker/compose/ 目录下运行
   cd docker/compose/
   docker-compose -f development.yml up -d  # 会创建新的 ./docker/compose/data/
   
   # ✅ 正确：在项目根目录运行
   cd /path/to/gonote
   docker-compose -f docker/compose/development.yml up -d
   ```

4. **使用了匿名卷** - 如果之前运行过但未指定挂载，数据在 Docker 管理的卷中
   ```bash
   # 查看 Docker 卷
   docker volume ls | grep gonote
   
   # 查看卷的详细信息
   docker inspect gonote-go | grep -A 20 Mounts
   ```

#### 创建测试笔记验证挂载

```bash
# 1. 启动容器
docker-compose -f docker/compose/development.yml up -d

# 2. 在宿主机创建测试笔记
cat > ./data/notes/test.md << 'EOF'
# Test Note
This is a test note created at $(date)
EOF

# 3. 在浏览器访问 http://localhost:9000 查看是否显示

# 4. 或在容器内验证
docker exec gonote-go cat /app/data/notes/test.md
```

#### 进入容器调试

```bash
# 进入容器 shell
docker exec -it gonote-go sh

# 查看当前工作目录
pwd

# 查看应用目录结构
ls -la /app/

# 查看数据目录
ls -la /app/data/
ls -la /app/data/notes/

# 查看配置文件（如果挂载）
cat /app/config.yaml

# 退出
exit
```

---

### Makefile 快捷命令

项目提供了 `Makefile` 快捷命令：

```bash
# 启动生产环境 Docker Compose
make docker-up

# 停止并清理容器
make docker-down

# 查看日志
make docker-logs

# 完整清理（包括数据）
make clean
```

---

## 💻 本地直接运行（无 Docker） / Local Native Run

适用于开发环境或希望直接运行二进制文件的场景。

### 前提条件

- **Go 1.24+**（建议使用最新稳定版）
- 可选：Make 工具（用于使用 `make` 命令）

---

### 运行步骤

```bash
# 1. 克隆仓库（如果还没有）
git clone https://github.com/gamosoft/gonote.git
cd gonote

# 2. 下载 Go 依赖
cd go
go mod download
cd ..

# 3. 运行（关键：从项目根目录运行！）
go run go/cmd/server/main.go --config go/config.yaml
```

⚠️ **重要提示**：

- 必须从**项目根目录**运行命令
- 不要进入 `go/` 目录后再运行
- 数据目录 `./data/` 是相对于当前工作目录解析的
- 如果从 `go/` 目录运行，数据会写入 `go/data/`（错误位置）

---

### 构建二进制文件

**使用 Make**（推荐）：

```bash
make build

# 二进制文件位于 ./bin/gonote
./bin/gonote --config go/config.yaml
```

**手动构建**：

```bash
cd go
go build -o ../gonote ./cmd/server
cd ..
./gonote --config go/config.yaml
```

---

## 🔐 安全配置要点 / Security Checklist

### 暴露到互联网前必须完成的配置

默认配置仅适用于本地测试。**任何公网暴露前**，请确保完成以下检查：

| # | 检查项 (Check Item) | 操作说明 (Action) |
|---|---------------------|-------------------|
| 1 | **修改默认密码** | 将 `config.yaml` 的 `authentication.password` 从 `admin` 改为强密码，或设置 `AUTHENTICATION_PASSWORD` 环境变量 |
| 2 | **生成随机密钥** | 修改 `authentication.secret_key`：<br>`openssl rand -hex 32` |
| 3 | **启用认证** | 设置 `authentication.enabled: true` |
| 4 | **启用限流** | 设置 `rate_limit.enabled: true` 防止暴力攻击 |
| 5 | **配置 CORS** | 修改 `server.allowed_origins` 为具体域名（不要用 `["*"]`） |
| 6 | **启用 HTTPS** | 生产环境必须使用 HTTPS，并设置 `secure_cookie: true` 或启用自动检测 |
| 7 | **使用反向代理** | 建议使用 nginx/Caddy 作为反向代理处理 SSL 终止 |

详细安全指南请参阅 [../security/SECURITY.md](../security/SECURITY.md)。

---

## ⚙️ 环境变量参考 / Environment Variables Reference

所有配置都可以通过环境变量覆盖，详见 [ENVIRONMENT_VARIABLES.md](ENVIRONMENT_VARIABLES.md)。

**常用生产环境变量示例**（Docker Compose `.env` 文件）：

```bash
# 服务器配置
PORT=8000

# 认证配置
AUTHENTICATION_ENABLED=true
AUTHENTICATION_PASSWORD=your_secure_password_here
AUTHENTICATION_SECRET_KEY=openssl_rand_hex_32_output_here
AUTHENTICATION_SECURE_COOKIE=true

# 限流配置
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=60
RATE_LIMIT_WINDOW=60

# 日志配置
LOG_ENABLED=true
```

---

## 📊 数据持久化 / Data Persistence

### 数据目录结构

```bash
data/
├── notes/          # 笔记 Markdown 文件（核心数据）
├── cache/          # 应用缓存
│   └── search/     # 搜索索引
├── temp/           # 临时文件
└── backups/        # 自动备份文件（如启用）
```

### Docker 数据挂载

Docker 部署时，必须将宿主机的 `./data` 目录挂载到容器的 `/app/data`：

```yaml
# docker-compose.yml 示例
volumes:
  - ./data:/app/data  # 关键：数据持久化
```

**确保数据目录存在**：

```bash
# 宿主机上创建数据目录
mkdir -p /path/to/your/gonote/data
```

---

### 备份策略

定期备份 `data/` 目录到安全位置：

- **云存储**：AWS S3、Backblaze B2、Google Cloud Storage
- **外部硬盘**：rsync + cron 定时备份
- **版本控制**：Git 仅适合备份配置和模板，媒体文件建议用专用备份方案

**自动备份脚本示例**：

```bash
#!/bin/bash
# backup.sh
tar -czf "/backup/gonote-$(date +%Y%m%d-%H%M%S).tar.gz" data/
```

使用 cron 定时执行：

```bash
# 每天凌晨 2 点备份
0 2 * * * /path/to/backup.sh
```

---

## 🌐 反向代理配置 / Reverse Proxy Configuration

### Nginx 完整示例

```nginx
# HTTP → HTTPS 重定向
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS 反向代理
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL 证书（Let's Encrypt 或其他 CA）
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    # 安全头部（可选但推荐）
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # 代理到 GoNote
    location / {
        proxy_pass http://localhost:9000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 3600s;  # WebSocket 需要长超时
    }
}
```

**使用 Let's Encrypt 免费证书**（Certbot）：

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书（自动配置 Nginx）
sudo certbot --nginx -d yourdomain.com
```

---

### Caddy 示例（更简单）

Caddy 自动处理 HTTPS，配置更简洁：

```caddyfile
yourdomain.com {
    reverse_proxy localhost:9000
    
    # 自动 HTTPS，无需证书配置
    # Caddy 会自动从 Let's Encrypt 获取证书
}
```

启动 Caddy：

```bash
caddy run
```

---

## 📱 平台特定指南 / Platform-Specific Guides

### Railway

Railway 自动检测 Dockerfile 或 Docker Compose 部署。

```bash
# 初始化项目
railway init

# 添加服务
railway add

# 部署
railway up
```

**环境变量设置**：

- 在 Railway Dashboard 的"Variables"中配置
- 注意设置 `PORT=8000`（Railway 默认端口）

---

### Fly.io

Fly.io 适合全球多区域部署。

```bash
# 生成 fly.toml 配置
fly launch
```

**配置要点**：

- `fly.toml` 中设置 `internal_port`（默认 `8080`）
- 设置 `PORT` 环境变量匹配 `internal_port`
- Fly 自动提供 HTTPS

---

## 🚨 常见问题与注意事项 / Common Issues & Notes

### 数据持久化

⚠️ **必须**确保 `data/` 目录持久化。Docker 容器重启后数据必须保留，务必正确挂载卷。

### 安全性

🔒 **不要暴露默认密码**：修改 `admin` 默认密码是首要安全步骤。

### HTTPS

🌐 **必须使用 HTTPS**：公网暴露时，务必使用反向代理 + SSL 证书（Let's Encrypt 免费）。

### 存储监控

📊 **监控磁盘空间**：媒体文件可能快速增长，定期检查存储使用量。

### 备份

💾 **定期备份**：虽然笔记是 Markdown 文件，但丢失仍会影响工作流。建议设置自动化备份。

---

## 📚 相关文档 / Related Documentation

- [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md) - **Docker 部署完整指南（路径映射、故障排查）**
- [ENVIRONMENT_VARIABLES.md](ENVIRONMENT_VARIABLES.md) - 环境变量完整参考
- [../docker/README.md](../docker/README.md) - Docker 详细使用说明
- [../security/SECURITY.md](../security/SECURITY.md) - 安全最佳实践
- [../security/AUTHENTICATION.md](../security/AUTHENTICATION.md) - 认证配置详解
- [../README.md](../README.md) - 项目快速入门
- [API.md](API.md) - REST API 参考

---

## 💡 部署检查清单 / Deployment Checklist

部署前逐项检查：

- [ ] 已从示例配置复制 `config.yaml` 并调整路径
- [ ] 已修改默认密码（如果启用认证）
- [ ] 已生成随机 `AUTHENTICATION_SECRET_KEY`
- [ ] 数据目录 `./data` 有正确读写权限
- [ ] 防火墙允许指定端口（如 8000、9000）
- [ ] 反向代理配置正确（生产环境）
- [ ] SSL 证书已安装并自动续期
- [ ] 已配置日志轮转或日志服务
- [ ] 已设置定期备份脚本

---

**祝部署顺利！** 🎉  
如有问题，请查阅各平台文档或在 GitHub Issues 提问。
