# 🐳 Docker 部署指南 / Docker Deployment Guide

本文档详细介绍 GoNote 的 Docker 部署配置、路径映射和常见问题解决方案。

---

## 📋 目录 / Table of Contents

1. [快速开始](#快速开始)
2. [Docker Compose 配置](#docker-compose-配置)
3. [路径映射说明](#路径映射说明)
4. [数据持久化](#数据持久化)
5. [常见问题](#常见问题)
6. [调试技巧](#调试技巧)

---

## 快速开始 / Quick Start

### 使用 Docker Run（最简单）

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

访问 http://localhost:9000

---

## Docker Compose 配置 / Docker Compose Configuration

### 可用配置文件

| 文件 | 用途 | 镜像来源 |
|------|------|----------|
| `docker/compose/production.yml` | **推荐** - 生产环境 | GHCR 预构建镜像 |
| `docker/compose/development.yml` | 开发环境 | 本地源码构建 |

### 生产环境部署（推荐）

使用预构建镜像，无需本地编译：

```bash
cd /path/to/gonote
docker-compose -f docker/compose/production.yml up -d
```

**配置说明：**
```yaml
services:
  gonote:
    image: ghcr.io/gamosoft/gonote:latest  # 预构建镜像
    ports:
      - "9000:9000"
    volumes:
      - ./data:/app/data  # 数据持久化
    restart: unless-stopped
```

### 开发环境部署

从本地源码构建，适合开发调试：

```bash
cd /path/to/gonote
docker-compose -f docker/compose/development.yml up -d
```

**配置说明：**
```yaml
services:
  gonote-go:
    build:
      context: ../..
      dockerfile: docker/go/Dockerfile
    ports:
      - "9000:9000"
    volumes:
      - ./data:/app/data
    environment:
      - TZ=Asia/Shanghai
      # 可添加其他环境变量
      # - AUTHENTICATION_ENABLED=true
      # - AUTHENTICATION_PASSWORD=your_password
```

### 使用 Make 命令

```bash
# 启动开发环境
make docker-up

# 启动生产环境
make docker-prod-up

# 停止服务
make docker-down

# 查看日志
docker-compose -f docker/compose/development.yml logs -f
```

---

## 路径映射说明 / Volume Mapping

### 核心概念

Docker 使用**卷挂载（Volume Mounting）**将宿主机目录映射到容器内部：

```yaml
volumes:
  - ./data:/app/data
    ^^^^^^^     ^^^^^^^^
    宿主机路径   容器内路径
```

### 路径对应关系表

| 容器内路径 | 宿主机路径 | 说明 | 重要性 |
|-----------|-----------|------|--------|
| `/app/data/notes/` | `./data/notes/` | **笔记数据** | ⭐⭐⭐ |
| `/app/data/cache/` | `./data/cache/` | 搜索索引缓存 | ⭐⭐ |
| `/app/data/temp/` | `./data/temp/` | 临时文件 | ⭐ |
| `/app/config.yaml` | 需手动挂载 | 配置文件（可选） | ⭐ |
| `/app/shared/themes/` | 需手动挂载 | 主题文件（可选） | ⭐ |
| `/app/locales/` | 需手动挂载 | 语言包（可选） | ⭐ |

### 完整挂载示例

```yaml
version: '3'
services:
  gonote:
    image: ghcr.io/gamosoft/gonote:latest
    ports:
      - "9000:9000"
    volumes:
      # 必需：数据持久化
      - ./data:/app/data
      
      # 可选：自定义配置
      - ./config.yaml:/app/config.yaml
      
      # 可选：自定义主题
      - ./shared/themes:/app/themes
      
      # 可选：自定义语言包
      - ./locales:/app/locales
    environment:
      - TZ=Asia/Shanghai
      - AUTHENTICATION_ENABLED=true
      - AUTHENTICATION_PASSWORD=your_password
```

---

## 数据持久化 / Data Persistence

### 为什么需要数据持久化？

Docker 容器是**临时的**，删除容器时，容器内未挂载的数据会**全部丢失**。

```bash
# ❌ 错误：没有挂载数据卷
docker run -d --name gonote -p 9000:9000 ghcr.io/gamosoft/gonote:latest

# 删除容器后，所有笔记数据都会丢失！
docker rm -f gonote
```

### 正确的数据持久化

```bash
# ✅ 正确：挂载数据卷
docker run -d --name gonote -p 9000:9000 \
  -v $(pwd)/data:/app/data \
  ghcr.io/gamosoft/gonote:latest

# 删除容器后，数据保留在宿主机的 ./data/ 目录
docker rm -f gonote

# 重新启动容器，数据依然存在
docker run -d --name gonote -p 9000:9000 \
  -v $(pwd)/data:/app/data \
  ghcr.io/gamosoft/gonote:latest
```

### 数据结构

```
data/
├── notes/          # 笔记文件（核心数据）
│   ├── note1.md
│   ├── note2.md
│   └── ...
├── cache/          # 应用缓存
│   └── search/     # 搜索索引
├── temp/           # 临时文件
└── backups/        # 备份文件（如启用）
```

---

## 常见问题 / Common Issues

### 问题 1：为什么我的笔记数据不见了？

**现象**：启动容器后，发现 `./data/notes/` 目录为空。

**可能原因和解决方案**：

#### 原因 1：容器未运行

数据是在容器运行时写入的，容器未运行自然没有数据。

```bash
# 检查容器状态
docker ps | grep gonote

# 如果容器未运行，启动它
docker-compose -f docker/compose/development.yml up -d

# 查看容器日志
docker logs gonote
```

#### 原因 2：工作目录错误

**必须从项目根目录启动 Docker Compose**，否则 `./data` 路径会解析错误。

```bash
# ❌ 错误：在 docker/compose/ 目录下运行
cd /path/to/gonote/docker/compose/
docker-compose -f development.yml up -d
# 结果：会在 docker/compose/data/ 创建数据，而不是项目根目录的 data/

# ✅ 正确：在项目根目录运行
cd /path/to/gonote
docker-compose -f docker/compose/development.yml up -d
```

#### 原因 3：挂载路径配置错误

检查 `docker-compose.yml` 中的 `volumes` 配置：

```yaml
# ❌ 错误：路径拼写错误
volumes:
  - ./dat:/app/data  # 应该是 ./data

# ✅ 正确
volumes:
  - ./data:/app/data
```

#### 原因 4：使用了匿名卷

如果之前运行过但未指定挂载，数据可能在 Docker 管理的匿名卷中。

```bash
# 查看 Docker 卷
docker volume ls | grep gonote

# 查看卷的详细信息
docker inspect gonote | grep -A 20 Mounts

# 如果使用了匿名卷，需要迁移数据
docker run --rm -v gonote_data:/source -v $(pwd)/data:/target alpine \
  cp -r /source /target
```

### 问题 2：如何验证挂载是否生效？

**方法 1：在宿主机查看（推荐）**
```bash
# 查看宿主机上的数据目录
ls -la ./data/notes/

# 创建测试文件
echo "# Test" > ./data/notes/test.md

# 在浏览器访问 http://localhost:9000 查看是否显示
```

**方法 2：进入容器查看**
```bash
# 进入容器
docker exec -it gonote sh

# 在容器内查看
ls -la /app/data/notes/
cat /app/data/notes/test.md

# 退出
exit
```

**方法 3：使用 docker exec 直接执行**
```bash
# 不进入容器，直接执行命令
docker exec gonote ls -la /app/data/notes/
docker exec gonote find /app/data -name "*.md"
```

### 问题 3：如何备份数据？

**备份到本地：**
```bash
# 简单备份
tar -czf gonote-backup-$(date +%Y%m%d).tar.gz ./data/

# 或使用 rsync
rsync -av ./data/ /backup/gonote-data/
```

**备份到云存储（以 AWS S3 为例）：**
```bash
# 安装 AWS CLI
aws s3 cp ./data/ s3://your-bucket/gonote-backup/ --recursive
```

**自动备份脚本示例：**
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/gonote"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR
tar -czf "$BACKUP_DIR/gonote-$DATE.tar.gz" ./data/

# 保留最近 7 天的备份
find $BACKUP_DIR -name "gonote-*.tar.gz" -mtime +7 -delete
```

**使用 cron 定时备份：**
```bash
# 每天凌晨 2 点备份
0 2 * * * /path/to/backup.sh
```

---

## 调试技巧 / Debugging Tips

### 1. 查看容器日志

```bash
# 查看最近日志
docker logs gonote

# 实时跟踪日志
docker logs -f gonote

# 查看最近 100 行
docker logs --tail 100 gonote
```

### 2. 进入容器调试

```bash
# 进入容器 shell
docker exec -it gonote sh

# 查看当前工作目录
pwd

# 查看应用目录结构
ls -la /app/

# 查看数据目录
ls -la /app/data/
ls -la /app/data/notes/

# 查看配置文件（如果挂载）
cat /app/config.yaml

# 查看环境变量
env | grep -i auth

# 退出
exit
```

### 3. 检查挂载信息

```bash
# 查看容器详细信息
docker inspect gonote

# 只看挂载部分
docker inspect gonote | grep -A 20 Mounts

# 或使用 Docker Compose
docker-compose ps
docker-compose top
```

### 4. 测试数据持久化

```bash
# 1. 启动容器
docker-compose -f docker/compose/development.yml up -d

# 2. 在宿主机创建测试笔记
cat > ./data/notes/test-persistence.md << 'EOF'
# Test Note
Created at: $(date)
This is a test note to verify data persistence.
EOF

# 3. 在容器内验证
docker exec gonote cat /app/data/notes/test-persistence.md

# 4. 删除并重新启动容器
docker-compose down
docker-compose up -d

# 5. 验证数据是否还在
docker exec gonote cat /app/data/notes/test-persistence.md
```

### 5. 健康检查

```bash
# 检查容器健康状态
docker inspect --format='{{.State.Health.Status}}' gonote

# 或查看健康检查日志
docker inspect --format='{{json .State.Health}}' gonote | jq
```

---

## 环境变量配置 / Environment Variables

在 Docker Compose 中配置环境变量：

```yaml
services:
  gonote:
    image: ghcr.io/gamosoft/gonote:latest
    environment:
      # 认证配置
      - AUTHENTICATION_ENABLED=true
      - AUTHENTICATION_PASSWORD=your_password
      - AUTHENTICATION_SECRET_KEY=openssl_rand_hex_32_output
      
      # 服务器配置
      - PORT=9000
      - DEBUG=false
      
      # 其他配置
      - LOG_ENABLED=true
      - RATE_LIMIT_ENABLED=true
```

完整的环境变量参考：[ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md)

---

## 安全建议 / Security Recommendations

1. **修改默认密码**：不要使用默认的 `admin` 密码
2. **生成随机密钥**：使用 `openssl rand -hex 32` 生成 `AUTHENTICATION_SECRET_KEY`
3. **启用认证**：设置 `AUTHENTICATION_ENABLED=true`
4. **使用 HTTPS**：在生产环境使用反向代理（nginx/Caddy）配置 SSL/TLS
5. **限制 CORS**：不要使用 `allowed_origins: ["*"]`
6. **定期备份**：定期备份 `./data/` 目录

详细安全指南：[SECURITY.md](../security/SECURITY.md)

---

## 相关文档 / Related Documentation

- [ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md) - 环境变量完整参考
- [DEPLOY_CN.md](./DEPLOY_CN.md) - 部署配置总览
- [SECURITY.md](../security/SECURITY.md) - 安全最佳实践
- [README.md](../../README.md) - 项目快速入门

---

**祝部署顺利！** 🎉

如有问题，请查阅 GitHub Issues 或社区论坛。
