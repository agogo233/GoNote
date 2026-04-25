# REST API 文档 / REST API Documentation

基础 URL (Base URL): `http://localhost:9000`

---

## 📝 笔记 API / Notes API

### 1. 列出所有笔记 / List All Notes

```http
GET /api/notes
```

返回所有笔记及其元数据和文件夹结构。

**响应示例 (Response Example):**

```json
{
  "success": true,
  "notes": [
    {
      "path": "welcome.md",
      "name": "welcome",
      "folder": "",
      "title": "欢迎",
      "tags": ["getting-started"],
      "modified": "2025-01-15T10:30:00+08:00"
    }
  ],
  "folders": ["projects", "archive"]
}
```

**使用示例 (Usage Example):**

```bash
# Linux / macOS
curl http://localhost:9000/api/notes

# Windows PowerShell
curl.exe http://localhost:9000/api/notes
```

---

### 2. 获取笔记内容 / Get Note Content

```http
GET /api/notes/{note_path}
```

检索特定笔记的内容。`note_path` 是相对于笔记根目录的文件路径。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | 笔记文件路径（如 `folder/mynote.md`） |

**响应示例 (Response Example):**

```json
{
  "success": true,
  "path": "tutorials/getting-started.md",
  "content": "# 快速开始\n\n本文档将帮助您...",
  "metadata": {
    "title": "快速开始",
    "tags": ["tutorial", "basics"],
    "modified": "2025-01-15T14:20:00+08:00"
  }
}
```

**错误响应 (Error Response):**

```json
{
  "detail": "笔记不存在：folder/missing.md"
}
```

**使用示例 (Usage Example):**

```bash
curl http://localhost:9000/api/notes/tutorials/getting-started.md
```

---

### 3. 创建或更新笔记 / Create or Update Note

```http
POST /api/notes/{note_path}
Content-Type: application/json
```

创建新笔记或更新现有笔记内容。路径会自动创建不存在的父目录。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | URL 路径中的笔记文件路径 |
| content | string | 是 | 请求体中的完整 Markdown 内容 |

**请求体 (Request Body):**

```json
{
  "content": "# 我的笔记\n\n这是笔记内容..."
}
```

**响应 (Response):**

```json
{
  "success": true,
  "path": "my-note.md",
  "message": "笔记已保存",
  "content": "# 我的笔记\n\n这是笔记内容..."
}
```

**注意 (Note):**

- 返回的 `content` 是处理后的内容（如添加了元数据前缀），与请求内容可能不同
- 父目录不存在时会自动创建
- 文件扩展名应为 `.md`

**使用示例 (Usage Example):**

```bash
# Linux / macOS
curl -X POST http://localhost:9000/api/notes/my-note.md \
  -H "Content-Type: application/json" \
  -d '{"content": "# Hello World\n\n这是测试内容。"}'

# Windows PowerShell
curl.exe -X POST http://localhost:9000/api/notes/my-note.md `
  -H "Content-Type: application/json" `
  -d "{\"content\": \"# Hello World\"}"
```

---

### 4. 删除笔记 / Delete Note

```http
DELETE /api/notes/{note_path}
```

永久删除笔记文件。注意：此操作不可撤销。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | 要删除的笔记文件路径 |

**响应 (Response):**

```json
{
  "success": true,
  "message": "笔记已删除：my-note.md"
}
```

**使用示例 (Usage Example):**

```bash
curl -X DELETE http://localhost:9000/api/notes/old-note.md
```

---

### 5. 移动笔记 / Move Note

```http
POST /api/notes/move
Content-Type: application/json
```

将笔记移动到新位置（支持重命名和跨文件夹移动）。

**请求体 (Request Body):**

```json
{
  "oldPath": "old-location/note.md",
  "newPath": "new-location/renamed-note.md"
}
```

**响应 (Response):**

```json
{
  "success": true,
  "message": "笔记已移动",
  "oldPath": "old-location/note.md",
  "newPath": "new-location/renamed-note.md"
}
```

**错误场景 (Error Scenarios):**

- 源文件不存在
- 目标路径已存在
- 权限不足

**使用示例 (Usage Example):**

```bash
curl -X POST http://localhost:9000/api/notes/move \
  -H "Content-Type: application/json" \
  -d '{"oldPath": "drafts/note.md", "newPath": "published/note.md"}'
```

---

## 🎬 媒体 API / Media API

### 6. 获取媒体文件 / Get Media File

```http
GET /api/media/{media_path}
```

检索媒体文件（图片、音频、视频、PDF）。此端点受认证保护。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| media_path | string | 是 | 媒体文件路径（相对于笔记根目录） |

**安全验证 (Security Validation):**

- ✅ 路径必须在笔记目录内（防止目录遍历攻击）
- ✅ 文件必须是支持的媒体格式
- ✅ 如果启用了认证，请求必须包含有效的会话 Cookie

**支持格式 (Supported Formats):**

| 类型 | 扩展名 | 最大文件大小 |
|------|--------|--------------|
| 图片 | jpg, jpeg, png, gif, webp | 10 MB |
| 音频 | mp3, wav, ogg, m4a | 50 MB |
| 视频 | mp4, webm, mov, avi | 100 MB |
| 文档 | pdf | 20 MB |

**使用示例 (Usage Example):**

```bash
curl http://localhost:9000/api/media/attachments/image-20240417.png
```

---

### 7. 上传媒体文件 / Upload Media File

```http
POST /api/upload-media
Content-Type: multipart/form-data
```

上传媒体文件到笔记的 `_attachments` 目录。文件会自动按文件夹组织，并以时间戳重命名以避免冲突。

**表单字段 (Form Fields):**

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | file | 是 | 要上传的媒体文件 |
| note_path | string | 否 | 关联的笔记路径（用于在侧边栏显示） |

**响应 (Response):**

```json
{
  "success": true,
  "path": "my-folder/_attachments/image-20240417093343.png",
  "filename": "image-20240417093343.png",
  "size": 245760,
  "message": "媒体上传成功"
}
```

**文件命名规则 (Naming Rule):**

原始文件名会转换为：`{original_name}-{timestamp}.{ext}`

示例：
- 上传 `photo.jpg` → `photo-20240417093343.jpg`
- 上传 `screen shot.png` → `screen shot-20240417093343.png`

**使用示例 (Usage Example):**

```bash
# 上传图片并关联到笔记
curl -X POST http://localhost:9000/api/upload-media \
  -F "file=@/path/to/photo.jpg" \
  -F "note_path=notes/my-note.md"

# 仅上传图片（不关联笔记）
curl -X POST http://localhost:9000/api/upload-media \
  -F "file=@/path/to/diagram.png"
```

**Windows PowerShell 示例:**

```powershell
curl.exe -X POST http://localhost:9000/api/upload-media `
  -F "file=@C:\Users\name\Pictures\image.jpg" `
  -F "note_path=project/notes.md"
```

---

### 8. 移动媒体文件 / Move Media File

```http
POST /api/media/move
Content-Type: application/json
```

移动媒体文件到新位置。支持在 UI 中拖拽操作。

**请求体 (Request Body):**

```json
{
  "oldPath": "_attachments/image.png",
  "newPath": "project/_attachments/image.png"
}
```

**响应 (Response):**

```json
{
  "success": true,
  "message": "媒体移动成功",
  "oldPath": "_attachments/old.png",
  "newPath": "projects/2025/_attachments/new.png"
}
```

**注意事项 (Notes):**

- 目标文件夹必须已存在
- 如果目标路径已存在文件，操作会失败
- 移动后，笔记中已存在的 Markdown 图片引用不会自动更新

---

## 📁 文件夹 API / Folders API

### 9. 创建文件夹 / Create Folder

```http
POST /api/folders
Content-Type: application/json
```

创建新文件夹（支持嵌套路径）。

**请求体 (Request Body):**

```json
{
  "path": "projects/2025/q1"
}
```

**响应 (Response):**

```json
{
  "success": true,
  "message": "文件夹已创建：projects/2025/q1"
}
```

**使用示例 (Usage Example):**

```bash
curl -X POST http://localhost:9000/api/folders \
  -H "Content-Type: application/json" \
  -d '{"path": "projects/report-2025"}'
```

---

### 10. 删除文件夹 / Delete Folder

```http
DELETE /api/folders/{folder_path}
```

删除文件夹及其所有内容（包括子文件夹和文件）。**此操作不可撤销。**

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| folder_path | string | 是 | 要删除的文件夹路径 |

**响应 (Response):**

```json
{
  "success": true,
  "message": "文件夹已删除：archive/2024",
  "deleted_files": 15
}
```

**使用示例 (Usage Example):**

```bash
curl -X DELETE http://localhost:9000/api/folders/archive/old
```

---

### 11. 移动文件夹 / Move Folder

```http
POST /api/folders/move
Content-Type: application/json
```

移动文件夹到新位置（支持重命名）。

**请求体 (Request Body):**

```json
{
  "oldPath": "drafts",
  "newPath": "archived/drafts-2024"
}
```

**响应 (Response):**

```json
{
  "success": true,
  "message": "文件夹已移动",
  "oldPath": "drafts",
  "newPath": "archived/drafts-2024"
}
```

---

### 12. 重命名文件夹 / Rename Folder

```http
POST /api/folders/rename
Content-Type: application/json
```

重命名文件夹（移动的快捷方式，保持了向后兼容）。

**请求体 (Request Body):**

```json
{
  "oldPath": "projects",
  "newName": "work"
}
```

**响应 (Response):**

```json
{
  "success": true,
  "message": "文件夹已重命名：projects → work",
  "oldPath": "projects",
  "newPath": "work"
}
```

---

## 🔍 搜索 API / Search API

### 13. 搜索笔记 / Search Notes

```http
GET /api/search?q={query}
```

在笔记内容中搜索文本。返回匹配的笔记列表及其相关片段。

**查询参数 (Query Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| q | string | 是 | 搜索关键词 |
| limit | integer | 否 | 返回结果数量，默认 50 |
| offset | integer | 否 | 分页偏移量，默认 0 |

**响应 (Response):**

```json
{
  "success": true,
  "query": "api",
  "total": 3,
  "results": [
    {
      "path": "api-overview.md",
      "title": "API 概览",
      "snippet": "本文档列出了所有可用的 REST API 端点...",
      "score": 0.95
    }
  ]
}
```

**使用示例 (Usage Example):**

```bash
# 基本搜索
curl "http://localhost:9000/api/search?q=authentication"

# 限制结果数量
curl "http://localhost:9000/api/search?q=config&limit=10"
```

---

## 🎨 主题 API / Themes API

### 14. 列出所有主题 / List Themes

```http
GET /api/themes
```

返回系统中所有可用的主题列表。

**响应 (Response):**

```json
{
  "success": true,
  "themes": [
    {
      "id": "light",
      "name": "Light",
      "description": "浅色主题，适合日间使用"
    },
    {
      "id": "dark",
      "name": "Dark",
      "description": "深色主题，适合夜间使用"
    }
  ]
}
```

---

### 15. 获取主题 CSS / Get Theme CSS

```http
GET /api/themes/{theme_id}
```

获取指定主题的 CSS 文件内容。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| theme_id | string | 是 | 主题 ID（如 "light"、"dark"） |

**响应 (Response):**

```css
/* 主题 CSS 内容 */
:root {
  --bg-primary: #ffffff;
  --text-primary: #1a1a1a;
}
```

**使用示例 (Usage Example):**

```bash
curl http://localhost:9000/api/themes/dark
```

---

## 🔗 知识图谱 API / Graph API

### 16. 获取笔记关系图谱 / Get Note Graph

```http
GET /api/graph
```

返回笔记之间的链接关系图谱数据。支持以下链接类型：

- **Wiki 链接**：`[[note]]` 或 `[[note|显示文本]]`
- **Markdown 链接**：`[文本](note.md)`

**响应 (Response):**

```json
{
  "success": true,
  "nodes": [
    {
      "id": "getting-started.md",
      "label": "快速开始",
      "group": 1
    },
    {
      "id": "api-docs.md",
      "label": "API 文档",
      "group": 2
    }
  ],
  "edges": [
    {
      "source": "getting-started.md",
      "target": "api-docs.md",
      "type": "wikilink"
    }
  ]
}
```

**图谱说明 (Graph Explanation):**

- `nodes`：笔记节点，`id` 为文件路径，`label` 为显示名称
- `edges`：链接边，`type` 为 `wikilink` 或 `markdown`
- `group`：用于前端颜色分组（相同文件夹的笔记分组相同）

---

## ⚙️ 系统 API / System API

### 17. 获取应用配置 / Get Application Config

```http
GET /api/config
```

返回当前运行的应用配置（不包含敏感信息）。

**响应 (Response):**

```json
{
  "success": true,
  "config": {
    "server": {
      "port": 9000,
      "host": "localhost"
    },
    "storage": {
      "notes_dir": "./data/notes"
    },
    "features": {
      "search_enabled": true,
      "authentication_enabled": false
    }
  }
}
```

---

### 18. 健康检查 / Health Check

```http
GET /health
```

返回系统健康状态。适用于负载均衡器和容器编排的健康检查。

**响应 (Response):**

```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T14:30:00+08:00",
  "version": "1.5.0",
  "uptime_seconds": 86400
}
```

**状态码 (Status Codes):**

- `200 OK`：系统正常运行
- `503 Service Unavailable`：系统不健康

---

### 19. Swagger UI 交互式文档 / Swagger UI Interactive Docs

```http
GET /api
```

访问交互式 API 文档（Swagger UI），可以直接在浏览器中测试所有端点。

**访问地址 (URL):**

```
http://localhost:9000/api
```

无需认证即可访问（如果认证已启用，需要先登录）。

---

## 🏷️ 标签 API / Tags API

### 20. 列出所有标签 / List All Tags

```http
GET /api/tags
```

返回笔记中使用的所有标签及其出现次数。

**响应 (Response):**

```json
{
  "success": true,
  "tags": {
    "python": 12,
    "tutorial": 8,
    "backend": 5,
    "frontend": 3
  }
}
```

---

### 21. 按标签获取笔记 / Get Notes by Tag

```http
GET /api/tags/{tag_name}
```

返回具有指定标签的所有笔记。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| tag_name | string | 是 | 标签名称（不包含 `#` 前缀） |

**响应 (Response):**

```json
{
  "success": true,
  "tag": "python",
  "count": 12,
  "notes": [
    {
      "path": "tutorials/python-basics.md",
      "name": "python-basics",
      "folder": "tutorials",
      "title": "Python 基础",
      "tags": ["python", "tutorial", "basics"],
      "modified": "2025-01-14T10:00:00+08:00"
    }
  ]
}
```

**使用示例 (Usage Example):**

```bash
curl http://localhost:9000/api/tags/python
```

---

## 📄 模板 API / Templates API

### 22. 列出所有模板 / List Templates

```http
GET /api/templates
```

返回 `_templates` 文件夹中所有可用模板。

**响应 (Response):**

```json
{
  "success": true,
  "templates": [
    {
      "name": "meeting-notes",
      "path": "_templates/meeting-notes.md",
      "modified": "2025-01-15T10:30:00+08:00"
    },
    {
      "name": "daily-journal",
      "path": "_templates/daily-journal.md",
      "modified": "2025-01-14T16:45:00+08:00"
    }
  ]
}
```

---

### 23. 获取模板内容 / Get Template Content

```http
GET /api/templates/{template_name}
```

返回指定模板的完整内容。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| template_name | string | 是 | 模板名称（不含 `.md` 扩展名） |

**响应 (Response):**

```json
{
  "success": true,
  "name": "meeting-notes",
  "content": "# 会议笔记\n\n**日期：** {{date}}\n\n**参会人：**\n\n## 讨论内容\n\n## 行动项\n\n## 下一步计划\n"
}
```

**使用示例 (Usage Example):**

```bash
curl http://localhost:9000/api/templates/meeting-notes
```

---

### 24. 从模板创建笔记 / Create Note from Template

```http
POST /api/templates/create-note
Content-Type: application/json
```

使用模板创建新笔记，自动替换所有占位符。

**请求体 (Request Body):**

```json
{
  "templateName": "meeting-notes",
  "notePath": "meetings/2025-q1.md"
}
```

**支持的占位符 (Supported Placeholders):**

| 占位符 | 替换为 | 示例 |
|--------|--------|------|
| `{{date}}` | 当前日期（YYYY-MM-DD） | 2025-01-15 |
| `{{time}}` | 当前时间（HH:MM:SS） | 14:30:00 |
| `{{datetime}}` | 完整日期时间 | 2025-01-15 14:30:00 |
| `{{timestamp}}` | Unix 时间戳（秒） | 1736958600 |
| `{{year}}` | 当前年份（YYYY） | 2025 |
| `{{month}}` | 当前月份（MM） | 01 |
| `{{day}}` | 当前日期（DD） | 15 |
| `{{title}}` | 笔记名称（不含扩展名） | weekly-report |
| `{{folder}}` | 父文件夹名称 | reports |

**响应 (Response):**

```json
{
  "success": true,
  "path": "meetings/2025-q1.md",
  "message": "笔记已从模板创建",
  "content": "# 会议笔记\n\n**日期：** 2025-01-15\n..."
}
```

**使用示例 (Usage Example):**

```bash
curl -X POST http://localhost:9000/api/templates/create-note \
  -H "Content-Type: application/json" \
  -d '{
    "templateName": "project-plan",
    "notePath": "projects/new-feature.md"
  }'
```

---

## 🔗 分享 API / Sharing API

### 25. 创建分享链接 / Create Share Link

```http
POST /api/share/{note_path}
Content-Type: application/json
```

为笔记生成公开分享链接（令牌 URL）。

**请求体 (Request Body):**

```json
{
  "theme": "dracula"
}
```

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | URL 路径中的笔记路径 |
| theme | string | 否 | 分享时使用的主题（默认：light） |

**响应 (Response):**

```json
{
  "success": true,
  "token": "aBcDeFgHiJk",
  "url": "http://localhost:9000/share/aBcDeFgHiJk",
  "note_path": "public/readme.md",
  "theme": "light",
  "expires_at": null
}
```

**使用示例 (Usage Example):**

```bash
curl -X POST http://localhost:9000/api/share/public/readme.md \
  -H "Content-Type: application/json" \
  -d '{"theme": "light"}'
```

---

### 26. 获取分享状态 / Get Share Status

```http
GET /api/share/{note_path}
```

检查笔记当前是否已分享。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | 要检查的笔记路径 |

**响应示例（已分享）(Response Example - Shared):**

```json
{
  "shared": true,
  "token": "aBcDeFgHiJk",
  "url": "http://localhost:9000/share/aBcDeFgHiJk",
  "theme": "light",
  "created": "2025-01-10T14:20:00+08:00"
}
```

**响应示例（未分享）(Response Example - Not Shared):**

```json
{
  "shared": false
}
```

---

### 27. 撤销分享 / Revoke Share

```http
DELETE /api/share/{note_path}
```

移除笔记的公开访问权限（使分享令牌失效）。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| note_path | string | 是 | 要撤销分享的笔记路径 |

**响应 (Response):**

```json
{
  "success": true,
  "message": "分享已撤销"
}
```

**使用示例 (Usage Example):**

```bash
curl -X DELETE http://localhost:9000/api/share/public/readme.md
```

---

### 28. 列出已分享笔记 / List Shared Notes

```http
GET /api/shared-notes
```

返回当前所有已分享笔记的路径。

**响应 (Response):**

```json
{
  "success": true,
  "paths": [
    "public/readme.md",
    "guides/quick-start.md"
  ],
  "count": 2
}
```

---

### 29. 查看分享笔记（公开端点）/ View Shared Note (Public Endpoint)

```http
GET /share/{token}
```

**公开端点** — 无需认证。返回独立 HTML 页面，渲染指定笔记。

**参数 (Parameters):**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| token | string | 是 | 分享令牌（由创建分享时生成） |

**响应 (Response):**

完整的 HTML 页面，包含：

- 笔记内容的渲染视图
- 创建分享时选择的主题样式
- 可选的 QR 码（方便移动端访问）
- "返回应用" 链接（指向首页）

**访问示例 (Access Example):**

```
http://localhost:9000/share/aBcDeFgHiJk
```

---

## 📝 通用响应格式 / Common Response Formats

### 成功响应 / Success Response

所有 API 端点成功时返回：

```json
{
  "success": true,
  "data": { ... }
}
```

部分端点使用扁平结构：

```json
{
  "success": true,
  "message": "操作成功",
  "path": "notes/new.md"
}
```

---

### 错误响应 / Error Response

```json
{
  "detail": "错误描述信息"
}
```

**HTTP 状态码 (HTTP Status Codes):**

| 状态码 | 说明 | 常见原因 |
|--------|------|----------|
| 200 OK | 请求成功 | - |
| 400 Bad Request | 请求参数错误 | JSON 格式错误、缺少必填字段 |
| 404 Not Found | 资源不存在 | 笔记/文件夹不存在 |
| 409 Conflict | 资源冲突 | 目标路径已存在 |
| 500 Internal Server Error | 服务器错误 | 文件系统错误、数据库错误 |

---

## 💡 使用建议 / Usage Tips

1. **交互式文档**：访问 `/api` 查看 Swagger UI，可直接在浏览器中测试端点
2. **认证**：如果启用了认证，大多数端点需要登录会话 Cookie
3. **路径规范**：所有路径都应使用正斜杠 `/`，不要以 `/` 开头
4. **文件扩展名**：笔记文件应使用 `.md` 扩展名
5. **编码**：请求和响应均为 UTF-8 编码
6. **速率限制**：当前版本无速率限制，生产环境建议配置反向代理限制

---

## 🔄 版本兼容性 / Version Compatibility

本 API 文档适用于 GoNote v1.0 及以上版本。向后不兼容的变更会随主要版本更新而发布。

**当前版本 (Current Version):** v1.0

**最后更新 (Last Updated):** 2025-01-15
