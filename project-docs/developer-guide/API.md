# 📡 API Documentation

Base URL: `http://localhost:9000`

All protected endpoints require authentication when `authentication.enabled: true` in config.yaml.

---

## 🔐 Authentication

### Login Page
```http
GET /login
```
Returns the HTML login page.

### Login
```http
POST /login
Content-Type: application/json

{
  "password": "your_password"
}
```
Authenticates the user session.

**Response:**
```json
{
  "success": true,
  "message": "Login successful"
}
```

### Logout
```http
POST /logout
```
Logs out the current user session.

---

## 🗂️ Notes

### List All Notes
```http
GET /api/notes?limit=50&page=1
```

Returns all notes with their metadata and folder structure. Supports pagination.

**Optional query parameters:**
- `limit` - Notes per page (default: 50, use with `page` for pagination)
- `page` - Page number (default: 1)
- `include_media` - Include media notes (`true`/`false`, default: `false`)

**Response:**
```json
{
  "notes": [
    {
      "name": "note",
      "path": "folder/note.md",
      "folder": "folder",
      "modified": "2025-11-26T11:00:00Z",
      "size": 1234,
      "type": "md",
      "tags": ["tag1", "tag2"]
    }
  ],
  "folders": ["folder"],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 10,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### Get Note Content
```http
GET /api/notes/{note_path}
```
Retrieve the content of a specific note.

**Example:**
```bash
curl http://localhost:9000/api/notes/folder/mynote.md
```

**Response:**
```json
{
  "path": "folder/mynote.md",
  "content": "# My Note\nContent here...",
  "metadata": {
    "created": "",
    "modified": "2025-11-26T11:00:00Z",
    "size": 1234,
    "lines": 10
  }
}
```

### Create/Update Note
```http
POST /api/notes/{note_path}
Content-Type: application/json

{
  "content": "# My Note\nNote content here...",
  "modified": "2026-07-03T14:30:00.123456789Z"
}
```

- `content` (required): Full Markdown content
- `modified` (optional): ISO 8601 timestamp of the last known file modification time (from a previous `GET /api/notes/{note_path}` response). If provided and the file has been modified by another source since, returns **409 Conflict**. Omit or pass empty string to skip the optimistic lock check (backward-compatible).

**Success Response (200):**
```json
{
  "success": true,
  "path": "test.md",
  "message": "Note saved successfully",
  "content": "# My Note\nNote content here...",
  "modified": "2026-07-03T14:30:05.987654321Z"
}
```

- `modified`: Server-authoritative file modification timestamp after save.

**Conflict Response (409):**
```json
{
  "detail": "Note modified by another source",
  "modified": "2026-07-03T14:30:05.987654321Z"
}
```
- The `modified` field contains the current server-side mtime. The client should prompt the user to either "Load Server Version" or "Keep My Version (Overwrite)".

**Linux/Mac:**
```bash
curl -X POST http://localhost:9000/api/notes/test.md \
  -H "Content-Type: application/json" \
  -d '{"content": "# Hello World"}'
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:9000/api/notes/test.md -H "Content-Type: application/json" -d "{\"content\": \"# Hello World\"}"
```

### Delete Note
```http
DELETE /api/notes/{note_path}
```

**Example:**
```bash
curl -X DELETE http://localhost:9000/api/notes/test.md
```

### Move Note
```http
POST /api/notes/move
Content-Type: application/json

{
  "oldPath": "note.md",
  "newPath": "folder/note.md"
}
```

**Response:**
```json
{
  "success": true,
  "oldPath": "note.md",
  "newPath": "folder/note.md",
  "message": "Note moved successfully"
}
```

---

## 🎬 Media

### Get Media
```http
GET /api/media/{media_path}
```
Retrieve a media file (image, audio, video, PDF) with authentication protection.

**Example:**
```bash
curl http://localhost:9000/api/media/folder/_attachments/image-20240417093343.png
```

**Security Note:** This endpoint requires authentication and validates that:
- The media path is within the notes directory (prevents directory traversal)
- The file exists and is a valid media format
- The requesting user is authenticated (if auth is enabled)

### Upload Media
```http
POST /api/upload-media
Content-Type: multipart/form-data

file: <media file>
note_path: <path of note to attach to>
```

Upload a media file to the `_attachments` directory. Files are automatically organized per-folder and named with timestamps to prevent conflicts.

**Supported formats & size limits:**
| Type | Formats | Max Size |
|------|---------|----------|
| Images | JPG, PNG, GIF, WebP | 10 MB |
| Audio | MP3, WAV, OGG, M4A | 50 MB |
| Video | MP4, WebM, MOV, AVI | 100 MB |
| Documents | PDF | 20 MB |

**Response:**
```json
{
  "success": true,
  "path": "folder/_attachments/media-20240417093343.png",
  "filename": "media-20240417093343.png",
  "message": "Media uploaded successfully"
}
```

**Example (using curl):**
```bash
curl -X POST http://localhost:9000/api/upload-media \
  -F "file=@/path/to/file.mp3" \
  -F "note_path=folder/mynote.md"
```

**Windows PowerShell:**
```powershell
curl.exe -X POST http://localhost:9000/api/upload-media -F "file=@C:\path\to\video.mp4" -F "note_path=folder/mynote.md"
```

### Move Media
```http
POST /api/media/move
Content-Type: application/json

{
  "oldPath": "_attachments/image.png",
  "newPath": "folder/_attachments/image.png"
}
```

Move a media file to a different location. Supports drag & drop in the UI.

**Response:**
```json
{
  "success": true,
  "message": "Media moved successfully",
  "newPath": "folder/_attachments/image.png"
}
```

### List Orphaned Media
```http
GET /api/media/orphaned
```

Lists media files that are not referenced by any notes (dangling attachments).

**Response:**
```json
{
  "success": true,
  "count": 1,
  "files": [
    {
      "path": "folder/_attachments/old-image.png",
      "filename": "old-image.png",
      "size": 102400,
      "mediaType": "image",
      "type": "image"
    }
  ],
  "totalSize": 102400
}
```

### Cleanup Orphaned Media
```http
DELETE /api/media/orphaned
```

Deletes all orphaned media files. Returns summary of deleted files.

**Response:**
```json
{
  "success": true,
  "deletedCount": 5,
  "deletedFiles": ["file1.png", "file2.png"],
  "freedSpace": 1048576,
  "message": "Successfully deleted orphaned media files"
}
```

**Rate Limit:** 30 requests per 60 seconds

---

## 📁 Folders

### Create Folder
```http
POST /api/folders
Content-Type: application/json

{
  "path": "Projects/2025"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Folder created successfully"
}
```

**Rate Limit:** 30 requests per 60 seconds

### Delete Folder
```http
DELETE /api/folders/{folder_path}
```
Deletes a folder and all its contents.

**Example:**
```bash
curl -X DELETE http://localhost:9000/api/folders/Projects/Archive
```

**Rate Limit:** 20 requests per 60 seconds

### Move Folder
```http
POST /api/folders/move
Content-Type: application/json

{
  "oldPath": "OldFolder",
  "newPath": "NewFolder"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Folder moved successfully"
}
```

**Rate Limit:** 20 requests per 60 seconds

### Rename Folder
```http
POST /api/folders/rename
Content-Type: application/json

{
  "oldPath": "Projects",
  "newName": "Work"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Folder renamed successfully"
}
```

**Rate Limit:** 30 requests per 60 seconds

---

## 🔍 Search

### Search Notes
```http
GET /api/search?q={query}
```

**Optional query parameters:**
- `q` - Search query (required)
- `mode` - Search mode (`full`, `title`, `smart`; default: `full`)
- `page` - Page number (default: 1)
- `limit` - Results per page (default: 50)

**Example:**
```bash
curl "http://localhost:9000/api/search?q=hello&page=1&limit=20"
```

**Response (pagination requested):**
```json
{
  "results": [
    {
      "name": "note",
      "path": "folder/note.md",
      "folder": "folder",
      "type": "md",
      "matches": [
        {
          "line_number": 10,
          "context": "...<mark>hello</mark> world..."
        }
      ],
      "score": 1.234
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 10,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

**Response (no pagination, backward-compatible):**
```json
{
  "results": [
    {
      "name": "note",
      "path": "folder/note.md",
      "folder": "folder",
      "type": "md",
      "matches": [
        {
          "line_number": 10,
          "context": "...<mark>hello</mark> world..."
        }
      ],
      "score": 1.234
    }
  ]
}
```

**Rate Limit:** 60 requests per 60 seconds

---

## 🏷️ Tags

### List All Tags
```http
GET /api/tags
```

Returns all tags found in notes with their usage counts.

**Response:**
```json
{
  "tags": {
    "python": 5,
    "tutorial": 3,
    "backend": 2
  }
}
```

### Get Notes by Tag
```http
GET /api/tags/{tag_name}
```

Returns all notes that have a specific tag.

**Response:**
```json
{
  "tag": "python",
  "count": 5,
  "notes": [
    {
      "name": "python-basics",
      "path": "tutorials/python-basics.md",
      "folder": "tutorials",
      "modified": "2025-11-26T11:00:00Z",
      "size": 1234,
      "type": "md",
      "tags": ["python", "tutorial"]
    }
  ]
}
```

---

## 📄 Templates

### List Templates
```http
GET /api/templates
```

Returns all available note templates from the `_templates` folder.

**Rate Limit:** 120 requests per 60 seconds

**Response:**
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

### Get Template Content
```http
GET /api/templates/{template_name}
```

Returns the content of a specific template.

**Parameters:**
- `template_name` - Template name (without .md extension)

**Response:**
```json
{
  "name": "meeting-notes",
  "content": "# Meeting Notes\n\nDate: {{date}}\nAttendees: {{attendees}}\n..."
}
```

**Rate Limit:** 120 requests per 60 seconds

### Create Note from Template
```http
POST /api/templates/create-note
Content-Type: application/json

{
  "templateName": "meeting-notes",
  "notePath": "meetings/weekly-sync.md"
}
```

Creates a new note from a template with placeholder replacement.

**Supported placeholders:**
- `{{date}}` - Current date (YYYY-MM-DD)
- `{{time}}` - Current time (HH:MM:SS)
- `{{datetime}}` - Current datetime
- `{{timestamp}}` - Unix timestamp
- `{{year}}` - Current year (YYYY)
- `{{month}}` - Current month (MM)
- `{{day}}` - Current day (DD)
- `{{title}}` - Note name without extension
- `{{folder}}` - Parent folder name

**Response:**
```json
{
  "success": true,
  "notePath": "meetings/weekly-sync.md",
  "message": "Note created from template"
}
```

**Rate Limit:** 60 requests per 60 seconds

---

## 🔗 Sharing

Share notes publicly without requiring authentication.

### Create Share Link
```http
POST /api/share/{note_path}
Content-Type: application/json

{
  "theme": "dracula"
}
```
Creates a share token for the note. The `theme` is optional (defaults to "light").

**Response:**
```json
{
  "success": true,
  "token": "LRFEo86oSVeJ3Gju",
  "url": "http://localhost:9000/share/LRFEo86oSVeJ3Gju",
  "note_path": "folder/note.md"
}
```

**Rate Limit:** 30 requests per 60 seconds

### Get Share Status
```http
GET /api/share/{note_path}
```
Check if a note is currently shared.

**Response:**
```json
{
  "shared": true,
  "token": "LRFEo86oSVeJ3Gju",
  "url": "http://localhost:9000/share/LRFEo86oSVeJ3Gju",
  "theme": "dracula",
  "created": "2026-01-15T10:30:00+00:00"
}
```

**Rate Limit:** 120 requests per 60 seconds

### Revoke Share
```http
DELETE /api/share/{note_path}
```
Removes public access to the note.

**Response:**
```json
{
  "success": true,
  "message": "Share revoked"
}
```

**Rate Limit:** 30 requests per 60 seconds

### List Shared Notes
```http
GET /api/shared-notes
```

Returns paths of all currently shared notes.

**Response:**
```json
{
  "paths": ["folder/note.md", "another.md"]
}
```

**Rate Limit:** 60 requests per 60 seconds

### View Shared Note (Public)
```http
GET /share/{token}
```

Public endpoint - no authentication required. Returns the note as a standalone HTML page with the theme set when sharing was created. Optionally generate QR code for mobile access.

**Example:**
```bash
curl http://localhost:9000/share/LRFEo86oSVeJ3Gju
```

---

## 🔗 Backlinks

Find all notes that link to a specific note using wikilink syntax `[[link]]` and markdown links.

### Get Backlinks
```http
GET /api/backlinks/{note_path}
```

Returns source notes and where they link to the target note.

**Response:**
```json
{
  "success": true,
  "backlinks": [
    {
      "source_path": "other-note.md",
      "link_texts": ["note-name", "folder/note"],
      "links": [
        {
          "line": 10,
          "context": "See also [[note-name]]",
          "type": "wikilink"
        },
        {
          "line": 15,
          "context": "Check [this](folder/note) out",
          "type": "markdown"
        }
      ]
    }
  ],
  "count": 1
}
```

**Rate Limit:** 60 requests per 60 seconds

---

## 📊 Statistics

Get comprehensive statistics for a note including word count, reading time, and structure analysis.

### Get Note Statistics
```http
GET /api/stats/{note_path}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "words": 150,
    "sentences": 12,
    "characters": 800,
    "total_characters": 1000,
    "reading_time_minutes": 1,
    "lines": 50,
    "paragraphs": 10,
    "list_items": 5,
    "tables": 1,
    "links": 8,
    "internal_links": 6,
    "external_links": 2,
    "wikilinks": 3,
    "code_blocks": 2,
    "inline_code": 5,
    "headings": {
      "h1": 1,
      "h2": 3,
      "h3": 2
    },
    "tasks": {
      "total": 4,
      "completed": 2,
      "pending": 2
    },
    "images": 1,
    "blockquotes": 2
  }
}
```

**Rate Limit:** 60 requests per 60 seconds

---

## 🎨 Themes

### List Themes
```http
GET /api/themes
```

Returns all available themes from the `themes` directory.

### Get Theme CSS
```http
GET /api/themes/{theme_id}
```

Returns the CSS content for the specified theme.

**Example:**
```bash
curl http://localhost:9000/api/themes/dark
```

---

## 🌍 Locales

### List Available Languages
```http
GET /api/locales
```

Returns all available locale (translation) files.

### Get Locale File
```http
GET /api/locales/{language_code}
```

Returns locale content for the specified language (e.g., `en-US`, `zh-CN`).

---

## 📡 WebSocket

### Real-time Updates
```http
GET /ws
```

WebSocket endpoint for receiving real-time notifications. When authentication is enabled, requires valid session.

**Events:**
- `notes_updated` - Broadcast when background note scan completes (new/updated/deleted notes detected)

**Example message:**
```json
{
  "event": "notes_updated",
  "timestamp": "2025-11-26T11:00:00Z"
}
```

---

## ⚙️ System

### Get Config
```http
GET /api/config
```

Returns application configuration (may filter sensitive fields like secret_key).

**Response:**
```json
{
  "name": "GoNote",
  "version": "0.25.0",
  "searchEnabled": true,
  "demoMode": false,
  "alreadyDonated": false,
  "authentication": {
    "enabled": false
  }
}
```

**Rate Limit:** 120 requests per 60 seconds

### Health Check
```http
GET /health
```

Returns system health status.

**Response:**
```json
{
  "status": "ok",
  "app": "GoNote",
  "version": "0.25.0"
}
```

---

## 🔒 Security Headers

GoNote includes the following security middleware:

- **CSRF Protection**: Double Submit Cookie pattern applied to state-changing operations
- **Rate Limiting**: Per-endpoint rate limits to prevent abuse
- **CORS**: Configurable allowed origins
- **Secure Cookies**: Automatically enabled when HTTPS detected

**CSRF Token**: The CSRF token is set in a cookie named `csrf_`. For state-changing requests, include the token in the `X-CSRF-Token` header. The token is readable by JavaScript (HTTPOnly=false) to facilitate this.

---

## 📝 Response Format

All endpoints return JSON responses:

**Success (generic):**
```json
{
  "success": true,
  "message": "Operation successful"
}
```

**Success (specific):**
Some endpoints use custom response structures tailored to their data (see individual endpoint docs).

**Error:**
```json
{
  "detail": "Error message"
}
```

---

## 🔢 Rate Limiting

### Global Rate Limiter
A global rate limiter is applied to all requests. The default configuration (when `RATE_LIMIT_ENABLED=true`) is:
- 30 requests per 1 second window

Configure via:
```yaml
rate_limit:
  enabled: true
  max_requests: 30
  window_seconds: 1
```

Or environment variables:
```
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=30
RATE_LIMIT_WINDOW=1
```

### Per-Endpoint Rate Limiting

Some endpoints have custom limits:

| Endpoint | Limit | Window |
|----------|-------|--------|
| `POST /login` | 10 | 60s |
| `POST /api/notes/move` | 30 | 60s |
| `DELETE /api/notes/*` | 30 | 60s |
| `POST /api/folders` | 30 | 60s |
| `POST /api/folders/rename` | 30 | 60s |
| `POST /api/share/*` | 30 | 60s |
| `DELETE /api/share/*` | 30 | 60s |
| `POST /api/media/move` | 30 | 60s |
| `POST /api/upload-media` | 20 | 60s |
| `POST /api/templates/create-note` | 60 | 60s |
| `GET /api/backlinks/*` | 60 | 60s |
| `GET /api/stats/*` | 60 | 60s |
| `GET /api/shared-notes` | 60 | 60s |
| `GET /api/templates` | 120 | 60s |
| `GET /api/templates/*` | 120 | 60s |
| `GET /api/share/*` | 120 | 60s |
| `GET /api/config` | 120 | 60s |

When rate limit is exceeded, server returns `429 Too Many Requests`.

---

## 💡 Usage Tips

- Use `/api/config` to discover current runtime configuration (except secrets)
- All responses are case-sensitive; response fields use camelCase (e.g., `deletedCount`, `searchEnabled`)
- For drag & drop media uploads, the `note_path` parameter indicates which note the media is attached to
- All paths are URL-encoded and case-sensitive
- When authentication is enabled, include session cookies in requests (browser handles this automatically; for API calls, ensure you're logged in)

---

**Tip:** For password-protected deployments, log in first via `POST /login` to obtain session cookies, then all `/api/*` endpoints will be accessible.
