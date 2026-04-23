# 🎨 主题与自定义

## 内置主题

GoNote 自带 **16 种精美主题**：

### 暗色主题

- 🌙 **Dark**——中性、安全的默认主题，使用 Tailwind 调色板
- 🧛 **Dracula**——流行的紫 tint 暗色主题
- ❄️ **Nord**——凉爽、北极-inspired 霜蓝调色板
- 🎨 **Monokai**——充满活力的青色强调色，Sublime Text 经典
- 💚 **Vue High Contrast**——深对比度，带绿色 tint，无障碍导向
- 🌊 **Cobalt2**——深海军蓝，配 vibrant 黄色高亮
- 🟫 **Gruvbox Dark**——温暖的复古橙色，Vim-inspired
- 🌃 **Tokyo Night**——深蓝紫色，配柔和霓虹点缀
- 🟣 **Catppuccin Mocha**——舒缓的薰衣草色，配柔和彩虹色
- 🤖 **One Dark**——平衡的蓝绿橙色，Atom 编辑器经典
- 🌹 **Rosé Pine**——深紫色，配温暖的玫瑰金和薰衣草色

### 亮色主题

- 🌞 **Light**——干净的白色配蓝色强调色，Tailwind 默认
- 🔷 **VS Blue**——经典 Visual Studio 蓝色，企业风格
- 🍵 **Matcha Light**——充满活力的绿色单色主题
- ☀️ **Solarized Light**——精确米色，减少眼睛疲劳
- ⚪ **One Light**——温暖白色配多色强调色，Atom 亮色

随时从侧边栏下拉菜单切换主题。您的偏好会自动保存！

## 创建自定义主题

### 分步指南

#### 1. 在主题目录中创建 CSS 文件

文件名将成为主题 ID（使用小写加连字符）：

```bash
cd gonote/themes
touch my-awesome-theme.css
```

#### 2. 定义主题 CSS 变量

**⚠️ 重要**：`data-theme` 属性**必须与**您的文件名（不含 `.css`）**匹配**。

如果您的文件名为 `my-awesome-theme.css`，请使用 `data-theme="my-awesome-theme"`：

```css
/* My Awesome Theme - 精美的自定义主题 */
/* 您的主题描述 */

:root[data-theme="my-awesome-theme"] {
    /* 背景颜色 */
    --bg-primary: #ffffff;       /* 主背景 */
    --bg-secondary: #f6f6f6;     /* 侧边栏/次要区域 */
    --bg-tertiary: #eeeeee;      /* 三级背景 */
    --bg-hover: #e5e5e5;         /* 悬停状态 */
    --bg-active: #d4d4d4;        /* 激活/按下状态 */

    /* 文本颜色 */
    --text-primary: #1a1a1a;     /* 主要文本 */
    --text-secondary: #4a4a4a;   /* 次要文本 */
    --text-tertiary: #6b6b6b;    /* 弱化/三级文本 */

    /* 边框颜色 */
    --border-primary: #d1d5db;   /* 主要边框 */
    --border-secondary: #e5e7eb; /* 微妙边框 */

    /* 强调色 */
    --accent-primary: #3b82f6;   /* 链接、按钮、高亮 */
    --accent-hover: #2563eb;     /* 强调色悬停状态 */
    --accent-light: rgba(59, 130, 246, 0.1); /* 强调色背景 */

    /* 状态颜色 */
    --success: #10b981;          /* 成功消息 */
    --error: #ef4444;            /* 错误消息 */
    --warning: #f59e0b;          /* 警告消息 */

    /* 阴影 */
    --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.15);
}
```

#### 3. 添加主题类型元数据（推荐）

在 CSS 文件**顶部**添加注释，说明主题是亮色还是暗色：

```css
/* @theme-type: light */
/* 或 */
/* @theme-type: dark */
```

**示例：**
```css
/* @theme-type: dark */
/* My Awesome Theme - 精美的自定义主题 */

:root[data-theme="my-awesome-theme"] {
    /* ... 您的 CSS 变量 ... */
}
```

**为什么需要这个？**
某些功能（如 Mermaid 图表、Chart.js）需要知道背景是亮色还是暗色，以便调整渲染颜色。此元数据会被应用自动解析。

**默认行为：** 如果不添加此元数据，主题将默认为 `dark` 以保持向后兼容。

#### 4. （可选）添加自定义 Emoji 图标

编辑主题加载逻辑，为您的主题添加自定义 emoji：

```python
# 在主题配置中添加
theme_icons = {
    # ... 现有主题 ...
    "my-awesome-theme": "🚀"  # 您的自定义 emoji
}
```

如果跳过此步骤，您的主题将使用 🎨 作为默认图标。

#### 5. 重启应用

```bash
# 如果使用 Docker：
docker-compose restart

# 如果本地运行：
# 停止服务器（Ctrl+C）并重新运行：
go run cmd/server/main.go
```

您的新主题将以下拉菜单中的 **"🚀 My Awesome Theme"** 出现！

---

## 主题开发提示

### ✅ 必需变量
所有这些 CSS 变量**必须**定义，主题才能正常工作：
- 背景：`bg-primary`、`bg-secondary`、`bg-tertiary`、`bg-hover`、`bg-active`
- 文本：`text-primary`、`text-secondary`、`text-tertiary`
- 边框：`border-primary`、`border-secondary`
- 强调色：`accent-primary`、`accent-hover`、`accent-light`
- 状态：`success`、`error`、`warning`
- 阴影：`shadow-sm`、`shadow-md`、`shadow-lg`

### 📋 快速开始
1. 复制现有主题文件（如 `dracula.css`）
2. 重命名为您的主题名称
3. 更新 `data-theme` 属性以匹配
4. 修改颜色
5. 重启应用

### 🔍 测试
- 使用浏览器 DevTools 实时实验颜色
- 测试亮色和暗色系统偏好
- 检查对比度比率以确保无障碍性（使用 [WebAIM 对比度检查器](https://webaim.org/resources/contrastchecker/)）
- 查看不同内容类型：代码块、表格、链接等

---

🎨 **提示：** 使用浏览器 DevTools 在创建主题之前实时实验颜色！
