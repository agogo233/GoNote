# 🏷️ 标签系统

使用 YAML frontmatter 中定义的标签组织和过滤您的笔记。

---

## 📋 基本用法

在笔记顶部添加标签 frontmatter：

```markdown
---
tags: [python, tutorial]
---

# 您的笔记内容

这里是笔记正文...
```

---

## 🔤 语法格式

### 推荐：内联数组格式

```yaml
---
tags: [python, tutorial, backend]
---
```

### 多行列表格式

```yaml
---
tags:
  - python
  - tutorial
  - backend
---
```

### 单个标签

```yaml
---
tags: python
---
```

---

## 🎯 核心功能

### 过滤笔记

- 点击侧边栏中的任意标签以过滤笔记
- 选择**多个标签**组合过滤器（AND 逻辑）
- 仅显示**同时具有所有选中标签**的笔记
- 标签计数徽章显示每个标签的使用次数

### 搜索与标签组合

- 单独使用标签：按类别过滤
- 单独使用文本搜索：查找内容
- 两者结合：缩小结果范围（如：在标记为"python"的笔记中搜索"async"）

---

### 显示模式说明

| 过滤模式 | 显示结果 |
|---------|---------|
| 无过滤 | 完整文件夹树 |
| 仅标签 | 匹配笔记的扁平列表 |
| 仅文本搜索 | 带匹配项的搜索结果 |
| 标签 + 文本 | 组合过滤的结果 |

---

## 💡 使用建议

- **标签命名**：小写字母，用连字符代替空格（如 `python`、`work-notes`）
- **保持一致性**：跨笔记使用统一的标签名称
- **层级组织**：使用相关标签组合（如 `python`、`python-async`、`python-web`）
- **避免过度**：每个笔记 3-5 个标签通常足够

---

## 📝 Frontmatter 规则

Frontmatter 必须遵循 YAML 格式规范：

- 第一行必须是单独的三连字符 `---`
- 最后一行必须是单独的三连字符 `---`
- `---` 之间的内容为 YAML 格式
- Frontmatter 在预览模式中自动隐藏

---

## 🌰 示例

### 项目笔记

```markdown
---
tags: [project, backend, api]
---

# API 文档

本文档描述后端 API...
```

### 知识库

```markdown
---
tags: [tutorial, beginner, docker]
---

# Docker 入门指南

学习 Docker 基础知识...
```

### 工作笔记

```markdown
---
tags: [meeting, q4-2024, planning]
---

# Q4 规划会议

会议记录...
```

---

## 🔗 相关文档

- [功能概览（FEATURES_CN.md）](../user-guide/FEATURES_CN.md) — 完整功能列表
- [笔记模板（TEMPLATES_CN.md）](../user-guide/TEMPLATES_CN.md) — 使用模板快速创建带标签的笔记

---

**文档版本**：v1.0  
**最后更新**：2025 年 1 月
