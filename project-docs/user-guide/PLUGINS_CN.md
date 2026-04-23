# 🔌 插件系统

> **⚠️ 状态更新（2026-03-31）：**
> - **Python 后端（遗留）：** 插件系统功能完整
> - **Go 后端（当前）：** 插件系统**尚未实现**
>
> Python 插件系统使用事件钩子扩展功能。计划未来开发 Go 原生插件系统。

---

GoNote 包含强大的插件系统，允许您在不修改核心代码的情况下扩展功能。

## 插件工作原理

插件是位于 `plugins/` 目录中的 Python 文件。它们使用**事件钩子**响应应用中的操作：

### 可用钩子

| 钩子 | 触发时机 | 参数 | 可修改 |
|------|----------------|------------|------------|
| `on_note_create` | 创建新笔记 | `note_path`、`initial_content` | ✅ 是（返回修改后的内容） |
| `on_note_save` | 保存笔记时 | `note_path`、`content` | ✅ 是（返回转换后的内容，或 None） |
| `on_note_load` | 从磁盘加载笔记时 | `note_path`、`content` | ✅ 是（返回转换后的内容，或 None） |
| `on_note_delete` | 删除笔记时 | `note_path` | ❌ 否 |
| `on_search` | 执行搜索时 | `query`、`results` | ❌ 否 |
| `on_app_startup` | 应用启动时 | 无 | ❌ 否 |

## 创建插件

### 1. 创建 Python 文件

```bash
cd gonote/plugins
touch my_plugin.py
```

### 2. 定义插件类

每个插件必须包含一个 `Plugin` 类，包含：
- `name`——显示名称
- `version`——版本字符串
- `enabled`——是否激活（默认：`True`）

### 3. 实现事件钩子

为您要处理的事件添加方法。

## 基础示例：笔记日志

这个简单的插件将笔记活动记录到 Docker 日志（使用 `docker-compose logs -f` 可见）：

```python
"""
笔记日志插件
记录所有笔记操作到 Docker 日志，用于监控
"""

class Plugin:
    def __init__(self):
        self.name = "笔记日志"
        self.version = "1.0.0"
        self.enabled = True

    def on_note_save(self, note_path: str, content: str) -> str | None:
        """记录笔记保存"""
        word_count = len(content.split())
        print(f"💾 笔记已保存：{note_path}（{word_count} 字）")
        return None  # 不修改内容，仅观察

    def on_note_delete(self, note_path: str):
        """记录笔记删除"""
        print(f"🗑️  笔记已删除：{note_path}")

    def on_search(self, query: str, results: list):
        """记录搜索查询"""
        print(f"🔍 搜索：'{query}' → {len(results)} 结果")
```

### 如何查看日志

```bash
# 实时查看日志
docker-compose logs -f

# 查看特定服务的日志
docker-compose logs -f gonote
```

## 激活您的插件

1. **将文件放入** `plugins/` 目录
2. **重启应用**：`docker-compose restart`
3. **插件自动加载**：`enabled = True` 的插件将自动加载

### 通过 API 启用/禁用插件

使用 API 切换插件开关：

**Linux/Mac：**
```bash
# 启用插件
curl -X POST http://localhost:9000/api/plugins/note_logger/toggle \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'

# 禁用插件
curl -X POST http://localhost:9000/api/plugins/note_logger/toggle \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

**Windows PowerShell：**
```powershell
# 启用插件
curl.exe -X POST http://localhost:9000/api/plugins/note_logger/toggle -H "Content-Type: application/json" -d "{\"enabled\": true}"

# 禁用插件
curl.exe -X POST http://localhost:9000/api/plugins/note_logger/toggle -H "Content-Type: application/json" -d "{\"enabled\": false}"
```

**列出所有插件（所有平台）：**
```bash
curl http://localhost:9000/api/plugins
```

## 插件状态持久化

插件状态（启用/禁用）保存在 `plugins/plugin_config.json` 中，重启后持久化。

---

💡 **提示：** 在插件中使用 `print()` 语句记录到 Docker 日志，用于调试和监控！
