# Mermaid 图表

GoNote 支持直接在 Markdown 笔记中创建 **Mermaid** 图表！Mermaid 允许您使用基于文本的定义创建图表和可视化，使其易于版本控制和协作。

## 如何使用

只需创建一个代码块，语言设置为 `mermaid`：

````markdown
```mermaid
graph TD
    A[开始] --> B{能工作吗？}
    B -->|是| C[太好了！]
    B -->|否| D[调试]
    D --> B
```
````

## 基础示例

### 流程图

````markdown
```mermaid
graph LR
    A[方框] --> B((圆形))
    A --> C(圆角矩形)
    B --> D{菱形}
    C --> D
```
````

**预览：**

```mermaid
graph LR
    A[方框] --> B((圆形))
    A --> C(圆角矩形)
    B --> D{菱形}
    C --> D
```

---

### 时序图

````markdown
```mermaid
sequenceDiagram
    Alice->>John: 你好 John，你好吗？
    John-->>Alice: 很好！
    Alice-)John: 回头见！
```
````

**预览：**

```mermaid
sequenceDiagram
    Alice->>John: 你好 John，你好吗？
    John-->>Alice: 很好！
    Alice-)John: 回头见！
```

---

### 类图

````markdown
```mermaid
classDiagram
    Animal <|-- Duck
    Animal <|-- Fish
    Animal : +int age
    Animal : +String gender
    Animal: +isMammal()
    class Duck{
        +String beakColor
        +swim()
        +quack()
    }
    class Fish{
        -int sizeInFeet
        -canEat()
    }
```
````

**预览：**

```mermaid
classDiagram
    Animal <|-- Duck
    Animal <|-- Fish
    Animal : +int age
    Animal : +String gender
    Animal: +isMammal()
    class Duck{
        +String beakColor
        +swim()
        +quack()
    }
    class Fish{
        -int sizeInFeet
        -canEat()
    }
```

---

### 状态图

````markdown
```mermaid
stateDiagram-v2
    [*] --> Still
    Still --> [*]
    Still --> Moving
    Moving --> Still
    Moving --> Crash
    Crash --> [*]
```
````

**预览：**

```mermaid
stateDiagram-v2
    [*] --> Still
    Still --> [*]
    Still --> Moving
    Moving --> Still
    Moving --> Crash
    Crash --> [*]
```

---

### 甘特图

````markdown
```mermaid
gantt
    title 项目时间线
    dateFormat  YYYY-MM-DD
    section 规划
    调研           :a1, 2024-01-01, 30d
    设计             :after a1, 20d
    section 开发
    后端            :2024-02-01, 40d
    前端           :2024-02-15, 35d
    section 测试
    集成测试  :2024-03-20, 15d
```
````

**预览：**

```mermaid
gantt
    title 项目时间线
    dateFormat  YYYY-MM-DD
    section 规划
    调研           :a1, 2024-01-01, 30d
    设计             :after a1, 20d
    section 开发
    后端            :2024-02-01, 40d
    前端           :2024-02-15, 35d
    section 测试
    集成测试  :2024-03-20, 15d
```

---

### 实体关系图

````markdown
```mermaid
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    CUSTOMER }|..|{ DELIVERY-ADDRESS : uses

    CUSTOMER {
        string name
        string email
        string phone
    }
    ORDER {
        int orderNumber
        date orderDate
        string status
    }
```
````

**预览：**

```mermaid
erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains
    CUSTOMER }|..|{ DELIVERY-ADDRESS : uses

    CUSTOMER {
        string name
        string email
        string phone
    }
    ORDER {
        int orderNumber
        date orderDate
        string status
    }
```

---

### 饼图

````markdown
```mermaid
pie title 宠物偏好
    "狗" : 45
    "猫" : 30
    "鸟" : 15
    "鱼" : 10
```
````

**预览：**

```mermaid
pie title 宠物偏好
    "狗" : 45
    "猫" : 30
    "鸟" : 15
    "鱼" : 10
```

---

### Git 图

````markdown
```mermaid
gitGraph
    commit
    commit
    branch develop
    checkout develop
    commit
    commit
    checkout main
    merge develop
    commit
```
````

**预览：**

```mermaid
gitGraph
    commit
    commit
    branch develop
    checkout develop
    commit
    commit
    checkout main
    merge develop
    commit
```

---

### 用户旅程

````markdown
```mermaid
journey
    title 我的一天
    section 上班
      泡茶：5: 我
      上楼：3: 我
      工作：1: 我，猫
    section 回家
      下楼：5: 我
      坐下：5: 我
```
````

**预览：**

```mermaid
journey
    title 我的一天
    section 上班
      泡茶：5: 我
      上楼：3: 我
      工作：1: 我，猫
    section 回家
      下楼：5: 我
      坐下：5: 我
```

---

### 思维导图

````markdown
```mermaid
mindmap
  root((GoNote))
    功能
      Markdown
      主题
      搜索
      文件夹
    集成
      MathJax
      Mermaid
      语法高亮
    优势
      快速
      简单
      离线优先
```
````

**预览：**

```mermaid
mindmap
  root((GoNote))
    功能
      Markdown
      主题
      搜索
      文件夹
    集成
      MathJax
      Mermaid
      语法高亮
    优势
      快速
      简单
      离线优先
```

---

## 主题支持

Mermaid 图表自动适配当前 GoNote 主题：
- **亮色主题**使用默认 Mermaid 配色方案
- **暗色主题**使用暗色优化颜色，确保适当对比度
- 主题更改会自动重新渲染所有图表

## 提示

1. **保持简单**：从基础图表开始，按需增加复杂度
2. **使用注释**：在 Mermaid 代码中添加 `%%` 作为注释
3. **测试语法**：如果图表未渲染，检查 Mermaid [文档](https://mermaid.js.org/)
4. **导出**：导出笔记为 HTML 时包含图表

## 更多信息

完整 Mermaid 语法和更多图表类型，请访问官方文档：
- [Mermaid 文档](https://mermaid.js.org/)
- [在线编辑器](https://mermaid.live/)——在线测试您的图表

---

**专业提示**：将 Mermaid 图表与 LaTeX 数学表达式和代码块结合，实现全面的技术文档！📊
