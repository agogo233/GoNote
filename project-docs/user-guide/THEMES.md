# 🎨 主题与自定义 / Themes & Customization

GoNote 支持 16 种内置主题，并允许您创建完全自定义的主题 CSS。

---

## 🎭 内置主题 / Built-in Themes

### 暗色主题 / Dark Themes

| 主题名称 | 标识符 (ID) | 风格说明 |
|---------|-------------|----------|
| **Dark** | `dark` | 中性、安全的默认主题，使用 Tailwind 调色板 |
| **Dracula** | `dracula` | 流行的紫色调暗色主题 |
| **Nord** | `nord` | 凉爽的北极蓝调色板 |
| **Monokai** | `monokai` | 充满活力的青色强调色，Sublime Text 经典 |
| **Vue High Contrast** | `vue-high-contrast` | 高对比度，带绿色调，无障碍导向 |
| **Cobalt2** | `cobalt2` | 深海军蓝，配亮黄色高亮 |
| **Gruvbox Dark** | `gruvbox-dark` | 温暖的复古橙色，Vim 风格 |
| **Tokyo Night** | `tokyo-night` | 深蓝紫色，配柔和霓虹点缀 |
| **Catppuccin Mocha** | `catppuccin-mocha` | 舒缓的薰衣草色，配柔和彩虹色 |
| **One Dark** | `one-dark` | 平衡的蓝绿橙色，Atom 编辑器经典 |
| **Rosé Pine** | `rose-pine` | 深紫色，配温暖的玫瑰金和薰衣草色 |

---

### 亮色主题 / Light Themes

| 主题名称 | 标识符 (ID) | 风格说明 |
|---------|-------------|----------|
| **Light** | `light` | 干净的白色配蓝色强调色，Tailwind 默认 |
| **VS Blue** | `vs-blue` | 经典 Visual Studio 蓝色，企业风格 |
| **Matcha Light** | `matcha-light` | 充满活力的绿色单色主题 |
| **Solarized Light** | `solarized-light` | 精确米色，减少眼睛疲劳 |
| **One Light** | `one-light` | 温暖白色配多色强调色，Atom 亮色 |

---

**切换方法：** 侧边栏下拉菜单选择主题，偏好会自动保存。

---

## 🔧 创建自定义主题 / Creating Custom Themes

### 完整步骤 / Step-by-Step Guide

#### 步骤 1：创建主题 CSS 文件

1. 进入主题目录：
```bash
cd gonote/themes
```

2. 创建新 CSS 文件，文件名将成为主题 ID（小写加连字符）：
```bash
touch my-awesome-theme.css
```

---

#### 步骤 2：定义主题 CSS 变量

**重要：** `data-theme` 属性必须与您的文件名（不含 `.css`）完全匹配。

如果文件名为 `my-awesome-theme.css`，则使用 `data-theme="my-awesome-theme"`：

```css
/* My Awesome Theme - 精美的自定义主题 */
/* 主题描述 */

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

---

#### 步骤 3：添加主题类型元数据（推荐）

在 CSS 文件**顶部**添加注释，声明主题是亮色还是暗色：

```css
/* @theme-type: dark */
/* @theme-type: light */
```

**完整示例：**
```css
/* @theme-type: dark */
/* My Awesome Theme - 深邃的工作区主题 */

:root[data-theme="my-awesome-theme"] {
    /* ... 变量定义 ... */
}
```

**为什么需要这个？**  
某些功能（如 Mermaid 图表、Chart.js）需要根据背景亮度调整渲染颜色。此元数据会被应用自动解析。

**默认行为：** 如果不添加元数据，主题将默认为 `dark` 以保持向后兼容。

---

#### 步骤 4：添加自定义图标（可选）

编辑 `go/internal/themes/themes.go`，在 `ThemeIcons` 映射中添加您的主题：

```go
var ThemeIcons = map[string]string{
    // ... 现有主题图标 ...
    "my-awesome-theme": "🚀", // 您的自定义 emoji
}
```

如果跳过此步骤，主题将使用 🎨 作为默认图标。

---

#### 步骤 5：重启应用

```bash
# Docker 部署
docker-compose restart

# 本地运行（先停止再启动）
cd go && go run cmd/server/main.go
```

您的新主题将以下拉菜单中的 **"🚀 My Awesome Theme"** 出现！

---

## 📋 主题开发参考 / Theme Development Reference

### 必需 CSS 变量清单

以下所有变量**必须**定义，主题才能正常工作：

| 类别 | 变量名 | 说明 |
|------|--------|------|
| **背景** | `--bg-primary` | 主背景色 |
| | `--bg-secondary` | 侧边栏/次要区域背景 |
| | `--bg-tertiary` | 三级背景（如弹出菜单） |
| | `--bg-hover` | 悬停状态背景 |
| | `--bg-active` | 激活/按下状态背景 |
| **文本** | `--text-primary` | 主要文本颜色 |
| | `--text-secondary` | 次要文本颜色 |
| | `--text-tertiary` | 弱化/三级文本颜色 |
| **边框** | `--border-primary` | 主要边框颜色 |
| | `--border-secondary` | 微妙边框颜色 |
| **强调色** | `--accent-primary` | 链接、按钮、高亮 |
| | `--accent-hover` | 强调色悬停状态 |
| | `--accent-light` | 强调色背景（半透明） |
| **状态** | `--success` | 成功消息颜色 |
| | `--error` | 错误消息颜色 |
| | `--warning` | 警告消息颜色 |
| **阴影** | `--shadow-sm` | 小阴影 |
| | `--shadow-md` | 中等阴影 |
| | `--shadow-lg` | 大阴影 |

---

### 快速开始模板

复制以下模板，快速创建新主题：

```css
/* @theme-type: dark */
/* Theme Name - 简短描述 */

:root[data-theme="theme-id"] {
    /* 背景 */
    --bg-primary: #1a1a1a;
    --bg-secondary: #2d2d2d;
    --bg-tertiary: #3a3a3a;
    --bg-hover: #404040;
    --bg-active: #4a4a4a;

    /* 文本 */
    --text-primary: #e5e5e5;
    --text-secondary: #a3a3a3;
    --text-tertiary: #737373;

    /* 边框 */
    --border-primary: #404040;
    --border-secondary: #525252;

    /* 强调色 */
    --accent-primary: #3b82f6;
    --accent-hover: #2563eb;
    --accent-light: rgba(59, 130, 246, 0.1);

    /* 状态 */
    --success: #10b981;
    --error: #ef4444;
    --warning: #f59e0b;

    /* 阴影 */
    --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
    --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.15);
}
```

---

### 测试建议

- ✅ 使用浏览器 DevTools 实时实验颜色
- ✅ 测试亮色和暗色系统偏好切换
- ✅ 检查对比度比率确保无障碍性（使用 [WebAIM 对比度检查器](https://webaim.org/resources/contrastchecker/)）
- ✅ 查看不同内容类型：代码块、表格、链接、Mermaid 图表等
- ✅ 在不同浏览器测试（Chrome、Firefox、Safari、Edge）

---

### 调试提示

1. **主题不显示？** 检查：
   - `data-theme` 属性是否与文件名完全匹配
   - CSS 文件是否放置在 `themes/` 目录
   - 应用是否重启

2. **某些元素颜色不对？** 可能缺少对应变量的定义，检查浏览器控制台是否有 CSS 变量未定义的警告。

3. **图标不显示？** 确认在 `themes.go` 中添加了图标映射。

---

## 💡 实用技巧

- **从现有主题开始**：复制一个类似风格的主题文件，然后修改颜色值
- **使用配色工具**：Coolors.co、Adobe Color 可以帮助生成配色方案
- **保持对比度**：确保文本与背景的对比度至少达到 WCAG AA 标准（4.5:1）
- **测试代码高亮**：不同语言的代码块在不同主题下都需清晰可读
- **考虑使用场景**：暗色主题适合夜间，亮色主题适合白天；也可以提供两者

---

**🎨 提示：** 使用浏览器 DevTools 的 Styles 面板，在创建主题之前实时实验颜色！

---

## 📚 相关文档

- [开发者指南 - 环境变量](../developer-guide/ENVIRONMENT_VARIABLES.md) - 通过环境变量切换主题
- [API 文档](../developer-guide/API.md) - 主题相关 API 端点
- [配置文件参考](../developer-guide/DEPLOY.md#配置文件说明) - 默认主题配置

---

**文档版本**：v1.0  
**最后更新**：2025 年 1 月  
**适用版本**：GoNote v1.0+
