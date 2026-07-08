# 📦 共享资源目录结构

本文档介绍 GoNote 项目中 `shared/` 目录的结构和用途。

---

## 📂 目录总览

```
shared/
├── frontend/        # 前端源代码（HTML、JavaScript、CSS）
│   └── libs/        # 第三方库（Tailwind CSS 等）
├── locales/         # 国际化翻译文件（JSON）
├── themes/          # 用户自定义主题 CSS
└── README.md        # 本目录说明
```

---

## 💻 frontend/ — 前端源代码

包含应用的完整前端实现。

### 目录结构

```
shared/frontend/
├── index.html           # 单页应用（SPA）主页面
├── app.js               # 主应用逻辑（Alpine.js + 自定义 JS）
├── styles.css           # 自定义补充样式（如有）
├── components/          # 可复用组件（如有）
└── libs/
    ├── tailwind/
    │   └── tailwind.css # 编译后的 Tailwind CSS
    └── ...             # 其他第三方库
```

**服务方式：**  
前端静态文件直接由 Go 后端服务，从 `shared/frontend/` 目录提供（Docker 容器内路径 `/app/frontend/`）。

---

## 🌍 locales/ — 翻译文件

存放 JSON 格式的国际化翻译文件。

### 文件命名规范

```
<language-code>.json
# 示例：en-US.json, zh-CN.json, es-ES.json
```

### 文件结构示例

```json
{
  "navigation": {
    "home": "Home",
    "settings": "Settings",
    "notes": "Notes"
  },
  "buttons": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete"
  },
  "messages": {
    "note_saved": "Note saved successfully!"
  }
}
```

**添加新语言：**

1. 复制现有语言文件（如 `en-US.json`）
2. 翻译所有字符串值（保持 JSON 结构不变）
3. 重命名为目标语言代码（如 `fr-FR.json`）
4. 在应用设置中切换到新语言

---

## 🎨 themes/ — 主题资源

包含应用的 CSS 主题文件，用于定制界面外观。

### 主题结构

每个主题是一个独立的 CSS 文件，使用 CSS 变量覆盖默认样式：

```css
/* shared/themes/dark.css */
:root[data-theme="dark"] {
    --bg-primary: #1a1a1a;
    --bg-secondary: #2d2d2d;
    --text-primary: #ffffff;
    --text-secondary: #a3a3a3;
    --accent-primary: #7c3aed;
    --border-primary: #404040;
    /* ... 更多变量 */
}
```

### 内置主题 vs 自定义主题

- **内置主题** — 位于 `shared/themes/` 目录下
- **用户自定义主题** — 用户放置在 `gonote/themes/`（Docker 挂载或直接创建）

详见 [用户指南 - 主题定制](../user-guide/THEMES.md)。

---

## 📍 相关资源目录对比

| 目录 | 用途 | 示例内容 |
|------|------|---------|
| `shared/frontend/libs/` | 第三方库缓存 | Tailwind CSS |
| `shared/themes/` | 内置及自定义主题 | `dark.css`, `dracula.css` |
| `shared/locales/` | 翻译文件 | `en-US.json`, `zh-CN.json` |
| `data/` | 用户笔记数据（不在 shared/） | `notes/`, `cache/`, `temp/` |
| `build/` | 构建配置文件 | Tailwind 配置、PostCSS 配置 |

---

## 🔧 开发工作流

### 修改前端代码

1. 编辑 `shared/frontend/` 中的文件
2. 修改自动生效（静态文件直接由服务器提供）
3. 如果修改了 Tailwind 类名，需重新构建 CSS：
   ```bash
   npm run css-build
   ```

---

### 自定义主题

1. 复制现有主题文件到 `shared/themes/` 或 `gonote/themes/`
2. 修改文件名（如 `my-theme.css`）
3. 编辑 CSS 变量
4. 重启服务器，在应用设置中选择新主题

---

### 添加翻译

1. 复制 `locales/en-US.json` 为 `<lang>.json`
2. 翻译所有字符串值（保持 JSON 结构不变）
3. 重启服务器，在设置中切换到新语言

---

## 🔨 构建配置参考

构建配置文件位于 `build/` 目录：

- `build/tailwind/input.css` — Tailwind 输入源
- `build/tailwind/tailwind.config.js` — Tailwind 配置
- `build/tailwind/postcss.config.js` — PostCSS 配置

完整构建说明请参阅 [BUILD.md](./BUILD.md)。

---

## 🎯 关键要点

1. **前端静态文件** — 全部在 `shared/frontend/`
2. **第三方库** — 放在 `shared/frontend/libs/`
3. **主题** — 放在 `shared/themes/`（内置）或 `gonote/themes/`（用户自定义）
4. **翻译** — 放在 `shared/locales/`
5. **数据** — 用户笔记数据存储在 `data/`（不在 `shared/` 下）

---

**相关文档：**

- [BUILD.md](../developer-guide/BUILD.md) — CSS 构建配置详解
- [THEMES.md](../user-guide/THEMES.md) — 主题使用与开发指南
- [AUTHENTICATION.md](../security/AUTHENTICATION.md) — 认证配置说明
