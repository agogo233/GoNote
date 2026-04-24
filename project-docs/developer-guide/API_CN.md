# 📡 API 文档

基础 URL：`http://localhost:9000`

## 🗂️ 笔记

### 列出所有笔记
```http
GET /api/notes
```
返回所有笔记及其元数据和文件夹结构。

**示例：**
```bash
curl http://localhost:9000/api/notes
```

### 获取笔记内容
```http
GET /api/notes/{note_path}
```
检索特定笔记的内容。

**示例：**
```bash
curl http://localhost:9000/api/notes/folder/mynote.md
```

### 创建/更新笔记
```http
POST /api/notes/{note_path}
Content-Type: application/json

{
  "content": "# 我的笔记\n内容在这里..."
}
```

**响应：**
```json
{
  "success": true,
  "path": "test.md",
  "message": "笔记创建成功",
  "content": "# 我的笔记\n内容在这里..."
}
```

**注意：** 新创建笔记的内容直接返回，不会被修改。


**Linux/Mac：**
```bash
curl -X POST http://localhost:9000/api/notes/test.md \
  -H "Content-Type: application/json" \
  -d '{"content": "# Hello World"}'
```

**Windows PowerShell：**
```powershell
curl.exe -X POST http://localhost:9000/api/notes/test.md -H "Content-Type: application/json" -d "{\"content\": \"# Hello World\"}"
```

### 删除笔记
```http
DELETE /api/notes/{note_path}
```

**示例：**
```bash
curl -X DELETE http://localhost:9000/api/notes/test.md
```

### 移动笔记
```http
POST /api/notes/move
Content-Type: application/json

{
  "oldPath": "note.md",
  "newPath": "folder/note.md"
}
```

## 🎬 媒体

### 获取媒体
```http
GET /api/media/{media_path}
```
检索媒体文件（图片、音频、视频、PDF），带有认证保护。

**示例：**
```bash
curl http://localhost:9000/api/media/folder/_attachments/image-20240417093343.png
```

**安全提示：** 此端点需要认证，并验证：
- 媒体路径在笔记目录内（防止目录遍历）
- 文件存在且是有效的媒体格式
- 请求用户已认证（如果启用了认证）

### 上传媒体
```http
POST /api/upload-media
Content-Type: multipart/form-data

file: <媒体文件>
note_path: <要附加到的笔记路径>
```

上传媒体文件到 `_attachments` 目录。文件会自动按文件夹组织，并以时间戳命名防止冲突。

**支持的格式和大小限制：**
| 类型 | 格式 | 最大大小 |
|------|---------|----------|
| 图片 | JPG、PNG、GIF、WebP | 10 MB |
| 音频 | MP3、WAV、OGG、M4A | 50 MB |
| 视频 | MP4、WebM、MOV、AVI | 100 MB |
| 文档 | PDF | 20 MB |

**响应：**
```json
{
  "success": true,
  "path": "folder/_attachments/media-20240417093343.png",
  "filename": "media-20240417093343.png",
  "message": "媒体上传成功"
}
```

**示例（使用 curl）：**
```bash
curl -X POST http://localhost:9000/api/upload-media \
  -F "file=@/path/to/file.mp3" \
  -F "note_path=folder/mynote.md"
```

**Windows PowerShell：**
```powershell
curl.exe -X POST http://localhost:9000/api/upload-media -F "file=@C:\path\to\video.mp4" -F "note_path=folder/mynote.md"
```

### 移动媒体
```http
POST /api/media/move
Content-Type: application/json

{
  "oldPath": "_attachments/image.png",
  "newPath": "folder/_attachments/image.png"
}
```

移动媒体文件到不同位置。支持 UI 中的拖拽操作。

**响应：**
```json
{
  "success": true,
  "message": "媒体移动成功",
  "newPath": "folder/_attachments/image.png"
}
```

**注意：**
- 媒体存储在相对于笔记位置的 `_attachments` 文件夹中
- 文件名自动添加时间戳（如 `media-20240417093343.mp3`）
- 媒体出现在侧边栏导航中，可直接查看/删除
- 拖拽文件到编辑器会自动上传并插入 Markdown
- 启用安全时，所有媒体访问需要认证

## 📁 文件夹

### 创建文件夹
```http
POST /api/folders
Content-Type: application/json

{
  "path": "Projects/2025"
}
```

### 删除文件夹
```http
DELETE /api/folders/{folder_path}
```
删除文件夹及其所有内容。

**示例：**
```bash
curl -X DELETE http://localhost:9000/api/folders/Projects/Archive
```

### 移动文件夹
```http
POST /api/folders/move
Content-Type: application/json

{
  "oldPath": "OldFolder",
  "newPath": "NewFolder"
}
```

### 重命名文件夹
```http
POST /api/folders/rename
Content-Type: application/json

{
  "oldPath": "Projects",
  "newName": "Work"
}
```

## 🔍 搜索

### 搜索笔记
```http
GET /api/search?q={query}
```

**示例：**
```bash
curl "http://localhost:9000/api/search?q=hello"
```

## 🎨 主题

### 列出主题
```http
GET /api/themes
```

### 获取主题 CSS
```http
GET /api/themes/{theme_id}
```

**示例：**
```bash
curl http://localhost:9000/api/themes/dark
```



## 🔗 知识图谱

### 获取笔记图谱
```http
GET /api/graph
```
返回笔记之间的关系图谱，带有链接检测。

**响应：**
```json
{
  "nodes": [
    { "id": "folder/note.md", "label": "note" },
    { "id": "another.md", "label": "another" }
  ],
  "edges": [
    { "source": "folder/note.md", "target": "another.md", "type": "wikilink" }
  ]
}
```

**链接检测：**
- **Wiki 链接** - `[[note]]` 或 `[[note|显示文本]]` 语法（Obsidian 风格）
- **Markdown 链接** - `[text](note.md)` 标准内部链接
- **边类型** - `"wikilink"` 或 `"markdown"` 区分链接来源

## ⚙️ 系统

### 获取配置
```http
GET /api/config
```
返回应用程序配置。

### 健康检查
```http
GET /health
```
返回系统健康状态。

### Swagger UI（交互式文档）
```http
GET /api
```
交互式 API 文档，支持在线测试（Swagger UI）。

---

## 🏷️ 标签

### 列出所有标签
`GET /api/tags`

返回笔记中所有标签及其使用次数。

**响应：**
```json
{
  "tags": {
    "python": 5,
    "tutorial": 3,
    "backend": 2
  }
}
```

### 按标签获取笔记
`GET /api/tags/{tag_name}`

返回具有特定标签的所有笔记。

**响应：**
```json
{
  "tag": "python",
  "notes": [
    {
      "path": "tutorials/python-basics.md",
      "name": "python-basics",
      "folder": "tutorials",
      "tags": ["python", "tutorial"]
    }
  ]
}
```

---

## 📄 模板

### 列出模板
`GET /api/templates`

返回 `_templates` 文件夹中所有可用模板。

**响应：**
```json
{
  "templates": [
    {
      "name": "meeting-notes",
      "path": "_templates/meeting-notes.md",
      "modified": "2025-11-26T10:30:00"
    },
    {
      "name": "daily-journal",
      "path": "_templates/daily-journal.md",
      "modified": "2025-11-26T10:25:00"
    }
  ]
}
```

### 获取模板内容
`GET /api/templates/{template_name}`

返回特定模板的内容。

**参数：**
- `template_name` - 模板名称（不含 .md 扩展名）

**响应：**
```json
{
  "name": "meeting-notes",
  "content": "# 会议笔记\n\n日期：{{date}}\n..."
}
```

### 从模板创建笔记
`POST /api/templates/create-note`

从模板创建新笔记，并替换占位符。

**请求体：**
```json
{
  "templateName": "meeting-notes",
  "notePath": "meetings/weekly-sync.md"
}
```

**占位符：**
- `{{date}}` - 当前日期（YYYY-MM-DD）
- `{{time}}` - 当前时间（HH:MM:SS）
- `{{datetime}}` - 当前日期时间
- `{{timestamp}}` - Unix 时间戳
- `{{year}}` - 当前年份（YYYY）
- `{{month}}` - 当前月份（MM）
- `{{day}}` - 当前日期（DD）
- `{{title}}` - 笔记名称（不含扩展名）
- `{{folder}}` - 父文件夹名称

**响应：**
```json
{
  "success": true,
  "path": "meetings/weekly-sync.md",
  "message": "从模板创建笔记成功",
  "content": "# 会议笔记\n\n日期：2025-11-26\n..."
}
```

---

## 🔗 分享

公开分享笔记，无需认证。

### 创建分享链接
```http
POST /api/share/{note_path}
Content-Type: application/json

{
  "theme": "dracula"
}
```
为笔记创建分享令牌。`theme` 可选（默认 "light"）。

**响应：**
```json
{
  "success": true,
  "token": "LRFEo86oSVeJ3Gju",
  "url": "http://localhost:9000/share/LRFEo86oSVeJ3Gju",
  "note_path": "folder/note.md"
}
```

### 获取分享状态
```http
GET /api/share/{note_path}
```
检查笔记当前是否已分享。

**响应：**
```json
{
  "shared": true,
  "token": "LRFEo86oSVeJ3Gju",
  "url": "http://localhost:9000/share/LRFEo86oSVeJ3Gju",
  "theme": "dracula",
  "created": "2026-01-15T10:30:00+00:00"
}
```

### 撤销分享
```http
DELETE /api/share/{note_path}
```
移除笔记的公开访问权限。

### 列出已分享的笔记
```http
GET /api/shared-notes
```
返回当前所有已分享笔记的路径。

**响应：**
```json
{
  "paths": ["folder/note.md", "another.md"]
}
```

### 查看分享笔记（公开）
```http
GET /share/{token}
```
公开端点 — 无需认证。返回独立 HTML 页面，带有创建分享时设置的主题。

---

## 📝 响应格式

所有端点返回 JSON 响应：

**成功：**
```json
{
  "success": true,
  "data": { ... }
}
```

**错误：**
```json
{
  "detail": "错误信息"
}
```
---

💡 **提示：** 访问 `/api` 查看交互式 Swagger UI 文档，可以直接在浏览器中测试端点！
