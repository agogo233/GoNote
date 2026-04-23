# 🏷️ 标签

## 概述

使用 YAML frontmatter 中定义的标签来组织和过滤您的笔记。

## 基本用法

在笔记顶部添加标签：

```markdown
---
tags: [python, tutorial]
---

# 您的笔记内容

其余笔记内容在此...
```

## 语法格式

### 内联数组（推荐）
```yaml
---
tags: [python, tutorial, backend]
---
```

### 多行列表
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

## 功能

### 过滤
- 点击侧边栏中的任意标签以过滤笔记
- 选择多个标签以组合过滤器（AND 逻辑）
- 仅显示具有**所有**选中标签的笔记
- 标签计数徽章显示每个标签的笔记数量

### 组合搜索
- 单独使用标签按类别过滤
- 单独使用文本搜索查找内容
- 结合两者缩小结果范围（例如，在标记为"python"的笔记中搜索"async"）

### 显示模式

| 过滤类型 | 显示 |
|------------|---------|
| 无 | 完整文件夹树 |
| 仅标签 | 匹配笔记的扁平列表 |
| 仅文本 | 带匹配项的搜索结果 |
| 标签 + 文本 | 组合过滤的结果 |

## 提示

- **标签名称**：小写，无空格（如 `python`、`work-notes`）
- **一致性**：跨笔记使用一致的标签名称
- **层级**：使用相关标签（如 `python`、`python-async`、`python-web`）
- **不要过度**：每个笔记 3-5 个标签通常足够

## Frontmatter 规则

- 必须在第一行以 `---` 开始
- 必须在单独一行以 `---` 结束
- 标记之间的内容为 YAML 格式
- Frontmatter 在预览中隐藏

## 示例

### 项目组织
```markdown
---
tags: [project, backend, api]
---

# API 文档
```

### 知识库
```markdown
---
tags: [tutorial, beginner, docker]
---

# Docker 入门指南
```

### 工作笔记
```markdown
---
tags: [meeting, q4-2024, planning]
---

# Q4 规划会议笔记
```
