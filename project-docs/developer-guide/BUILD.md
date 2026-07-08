# 🔨 构建配置

本文档介绍 GoNote 项目的构建配置文件，主要面向需要构建或修改前端样式的开发者。

---

## 📂 目录结构

```
build/
└── tailwind/
    ├── input.css           # Tailwind CSS 输入源
    ├── tailwind.config.js  # Tailwind 配置
    └── postcss.config.js   # PostCSS 配置
```

这些配置文件用于构建前端样式系统。

---

## 🎨 构建 CSS

### 开发模式

```bash
# 从项目根目录运行
npm install
npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --watch
```

该命令会启动监听模式，当 `input.css` 中的类引用变更时自动重新构建。

### 生产模式

```bash
# 从项目根目录运行
npm install
npx tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --minify
```

这会生成压缩后的 CSS 文件，删除未使用的样式以减小体积。

### 使用 Make

```bash
# 安装依赖（仅一次）
make deps

# 构建 CSS
make css-build

# 监听文件变化（开发时）
make css-watch
```

---

## ⚙️ 配置文件详解

### tailwind.config.js

Tailwind CSS 配置文件，定义：
- **content**: 扫描哪些文件中的 HTML/JS 以提取用到的 CSS 类
- **darkMode**: 暗色模式策略（`class` 策略）
- **theme**: 自定义主题变量（颜色、字体、间距等扩展）
- **plugins**: 启用的 Tailwind 插件

#### 当前配置要点

```javascript
module.exports = {
  content: [
    './shared/frontend/**/*.html',
    './shared/frontend/**/*.js',
    './shared/themes/**/*.css'
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 自定义颜色变量
        primary: 'var(--color-primary)',
        'bg-primary': 'var(--bg-primary)',
        'text-primary': 'var(--text-primary)',
        // ... 更多 CSS 变量映射
      }
    }
  },
  plugins: []
}
```

### input.css

Tailwind 的入口 CSS 文件，通常包含：
- `@tailwind base;` - 基础样式（重置、盒模型等）
- `@tailwind components;` - 组件样式（按钮、卡片等）
- `@tailwind utilities;` - 工具类（间距、颜色、定位等）
- 任何自定义 CSS 规则或 `@layer` 块

示例：
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* 自定义样式 */
.note-preview h1 {
  @apply text-2xl font-bold;
}
```

---

## 🔧 环境要求

- **Node.js**: 18+（建议与前端开发版本一致）
- **npm** 或 **yarn**: 包管理器
- **Make** (可选): 如果使用 Make targets

---

## 📦 依赖管理

项目依赖在 `package.json` 中定义：

```json
{
  "devDependencies": {
    "tailwindcss": "^3.x",
    "postcss": "^8.x",
    "autoprefixer": "^10.x"
  },
  "scripts": {
    "css-build": "tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --minify",
    "css-watch": "tailwindcss -i ./build/tailwind/input.css -o ./shared/frontend/libs/tailwind/tailwind.css --watch"
  }
}
```

运行 `npm install` 安装依赖到 `node_modules/`。

---

## 🚀 完整工作流示例

### 修改 UI 样式

1. **修改 Tailwind 类**  
   编辑 `shared/frontend/` 中的 HTML/JS 文件，例如：
   ```html
   <button class="bg-blue-500 text-white px-4 py-2 rounded">Click</button>
   ```

2. **重新构建 CSS**（如果添加了新类）
   ```bash
   npm run css-build
   ```
   
   或使用监听模式：
   ```bash
   npm run css-watch
   ```

3. **验证应用**  
   刷新浏览器，检查样式是否按预期应用。

### 添加新主题

1. 在 `shared/themes/` 创建新 CSS 文件（如 `my-theme.css`）
2. 在 HTML 中通过 `data-theme` 属性或设置界面选择主题
3. 主题机制基于 CSS 变量：重写 `:root` 中的 `--*` 变量

### 定制化构建（进阶）

如需深度定制（如修改 Tailwind 默认断点、颜色等），编辑 `build/tailwind/tailwind.config.js`，然后重建。

---

## 🔍 问题排查

| 问题 | 解决方案 |
|------|----------|
| 新增的 Tailwind 类不生效 | 确保类名在 `tailwind.config.js` 的 `content` 扫描路径中，并重新运行 `npm run css-build` |
| CSS 文件很大 | 生产构建使用 `--minify`；精简 `content` 路径；避免 `@apply` 滥用 |
| 暗色模式不工作 | 检查 `darkMode: 'class'` 配置；确保 HTML 有 `class="dark"` |
| 自定义颜色不生效 | 检查 `tailwind.config.js` 的 `theme.extend.colors` 中变量名是否拼写正确 |

---

## 📚 参考资源

- [Tailwind CSS 官方文档](https://tailwindcss.com/docs)
- [PostCSS 官方文档](https://postcss.org/docs/)
- [项目 README](../README.md) - 快速开始
- [ASSETS.md](./ASSETS.md) - 资源目录说明

---

**提示**：Docker 镜像已包含编译后的 CSS。开发时可直接修改 `shared/frontend/`，无需重复构建，除非使用了 Tailwind 工具类。构建步骤主要用于生成或更新 Tailwind 的样式表。
