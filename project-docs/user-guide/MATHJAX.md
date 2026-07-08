# 🧮 LaTeX 与 MathJax 指南

GoNote 支持使用 **LaTeX 语法**在笔记中编写精美数学公式，由 MathJax 3 驱动。

---

## 🎯 快速开始

### 行内公式（与文本同行）

使用 `$...$` 包裹：

```markdown
爱因斯坦方程：$E = mc^2$
```

渲染结果：爱因斯坦方程：$E = mc^2$

---

### 独立公式（居中独占一行）

使用 `$$...$$` 包裹：

```markdown
$$
x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}
$$
```

渲染结果：

$$
x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}
$$

---

## 📘 基础语法

### 上标与下标

**上标**（`^`）：

```markdown
$x^2$ → $x^2$
$e^{i\pi}$ → $e^{i\pi}$
```

**下标**（`_`）：

```markdown
$x_1$ → $x_1$
$a_{ij}$ → $a_{ij}$
```

**组合使用**：

```markdown
$x_1^2$ → $x_1^2$
$\sum_{i=1}^{n} i^2$ → $\sum_{i=1}^{n} i^2$
```

---

### 分数与根号

**分数：**

```markdown
$\frac{a}{b}$ → $\frac{a}{b}$
$\frac{\frac{1}{x}+\frac{1}{y}}{x+y}$ → 复杂分数示例
```

**平方根：**

```markdown
$\sqrt{2}$ → $\sqrt{2}$
$\sqrt[3]{8}$ → $\sqrt[3]{8}$（立方根）
$\sqrt{x^2 + y^2}$ → $\sqrt{x^2 + y^2}$
```

---

## 🔤 希腊字母

### 小写希腊字母

```markdown
$\alpha, \beta, \gamma, \delta, \epsilon, \zeta, \eta, \theta$
→ $\alpha, \beta, \gamma, \delta, \epsilon, \zeta, \eta, \theta$

$\lambda, \mu, \pi, \sigma, \tau, \phi, \chi, \psi, \omega$
→ $\lambda, \mu, \pi, \sigma, \tau, \phi, \chi, \psi, \omega$
```

---

### 大写希腊字母

```markdown
$\Gamma, \Delta, \Theta, \Lambda, \Xi, \Pi, \Sigma, \Phi, \Psi, \Omega$
→ $\Gamma, \Delta, \Theta, \Lambda, \Xi, \Pi, \Sigma, \Phi, \Psi, \Omega$
```

---

## ∫ 微积分符号

### 积分

**定积分：**

$$
\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

**多重积分：**

$$
\iiint_V f(x,y,z) \, dx \, dy \, dz
$$

---

### 导数

**一阶导数：** `$\frac{df}{dx}$` → $\frac{df}{dx}$

**偏导数：** `$\frac{\partial f}{\partial x}$` → $\frac{\partial f}{\partial x}$

**梯度：**

$$
\nabla f = \frac{\partial f}{\partial x}\mathbf{i} + \frac{\partial f}{\partial y}\mathbf{j} + \frac{\partial f}{\partial z}\mathbf{k}
$$

---

### 极限

$$
\lim_{x \to \infty} \frac{1}{x} = 0
$$

---

## Σ 求和与乘积

### 求和

$$
\sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

$$
\sum_{k=1}^{\infty} \frac{1}{k^2} = \frac{\pi^2}{6}
$$

---

### 乘积

$$
\prod_{i=1}^{n} i = n!
$$

---

## 🔢 矩阵

### 基础 2×2 矩阵

$$
\begin{bmatrix}
a & b \\
c & d
\end{bmatrix}
$$

**代码：**
```markdown
$$
\begin{bmatrix}
a & b \\\
c & d
\end{bmatrix}
$$
```

**注意：** 使用 `\\\ `（三个反斜杠 + 空格）作为行分隔符，确保正确渲染。

---

### 3×3 矩阵

$$
A = \begin{bmatrix}
1 & 2 & 3 \\
4 & 5 & 6 \\
7 & 8 & 9
\end{bmatrix}
$$

---

### 向量与行列式

**单位矩阵：**

$$
I = \begin{pmatrix}
1 & 0 & 0 \\
0 & 1 & 0 \\
0 & 0 & 1
\end{pmatrix}
$$

**行列式：**

$$
\det(A) = \begin{vmatrix}
a & b \\
c & d
\end{vmatrix} = ad - bc
$$

---

## 🔗 高级结构

### 方程组

$$
\begin{cases}
x + y = 5 \\
2x - y = 1
\end{cases}
$$

---

### 对齐方程

$$
\begin{aligned}
f(x) &= (x+1)^2 \\
&= x^2 + 2x + 1
\end{aligned}
$$

---

### 连分数

$$
\phi = 1 + \frac{1}{1 + \frac{1}{1 + \frac{1}{1 + \cdots}}}
$$

---

## 🔣 常用数学符号

### 运算符

| 符号 | LaTeX 命令 | 渲染结果 |
|------|------------|----------|
| ± | `\pm` | $\pm$ |
| × | `\times` | $\times$ |
| ÷ | `\div` | $\div$ |
| ≠ | `\neq` | $\neq$ |
| ≤ | `\leq` | $\leq$ |
| ≥ | `\geq` | $\geq$ |
| ≈ | `\approx` | $\approx$ |
| ∞ | `\infty` | $\infty$ |

---

### 集合论

| 符号 | LaTeX 命令 | 渲染结果 |
|------|------------|----------|
| ∈ | `\in` | $\in$ |
| ∉ | `\notin` | $\notin$ |
| ⊂ | `\subset` | $\subset$ |
| ∪ | `\cup` | $\cup$ |
| ∩ | `\cap` | $\cap$ |
| ∅ | `\emptyset` | $\emptyset$ |

---

### 逻辑符号

| 符号 | LaTeX 命令 | 渲染结果 |
|------|------------|----------|
| ∧ | `\land` | $\land$ |
| ∨ | `\lor` | $\lor$ |
| ¬ | `\neg` | $\neg$ |
| ⇒ | `\implies` | $\implies$ |
| ⇔ | `\iff` | $\iff$ |
| ∀ | `\forall` | $\forall$ |
| ∃ | `\exists` | $\exists$ |

---

## 🌟 著名公式示例

### 欧拉恒等式

$$
e^{i\pi} + 1 = 0
$$

---

### 勾股定理

$$
a^2 + b^2 = c^2
$$

---

### 薛定谔方程

$$
i\hbar\frac{\partial}{\partial t}\Psi(\mathbf{r},t) = \hat{H}\Psi(\mathbf{r},t)
$$

---

### 麦克斯韦方程组

$$
\begin{aligned}
\nabla \cdot \mathbf{E} &= \frac{\rho}{\epsilon_0} \\
\nabla \cdot \mathbf{B} &= 0 \\
\nabla \times \mathbf{E} &= -\frac{\partial \mathbf{B}}{\partial t} \\
\nabla \times \mathbf{B} &= \mu_0\mathbf{J} + \mu_0\epsilon_0\frac{\partial \mathbf{E}}{\partial t}
\end{aligned}
$$

---

## 💡 实用提示

### 1. 预览模式

始终使用**分栏视图**或**预览模式**实时查看公式渲染效果。

---

### 2. 转义美元符号

如果需要显示字面美元符号（非数学公式），请转义：

```markdown
市场价格为 $100 美元
→ 添加反斜杠：\$100 美元

显示：市场价格为 \$100 美元
```

---

### 3. 多行公式格式化

**关键技巧**：矩阵和多行公式中，使用 **3 个反斜杠 + 空格**（`\\\ `）换行：

```markdown
✅ 正确（可读的多行格式）：
$$
\begin{bmatrix}
a & b \\\
c & d
\end{bmatrix}
$$

❌ 错误（只有 2 个反斜杠 - 不会正确换行）：
$$
\begin{bmatrix}
a & b \\
c & d
\end{bmatrix}
$$
```

**秘诀：** 每行末尾使用 `\\\ `（三个反斜杠 + 尾部空格），然后换行。

---

### 4. 调试未渲染的公式

如果公式未正确渲染：

1. ✅ 检查 `$...$` 或 `$$...$$` 是否配对完整
2. ✅ 确保反斜杠正确（`\frac` 不是 `/frac`）
3. ✅ 查找未转义的特殊字符
4. ✅ 矩阵使用 `\\\ ` 而非 `\\`
5. ✅ 确保 `\\\` 后有空格再换行

---

### 5. 性能考虑

MathJax 渲染效率高，但公式密集的笔记（100+ 公式）可能需要稍长时间排版。建议：

- 分屏预览实时检查
- 导出 HTML 前确认所有公式正常

---

## 📚 学习资源

### 官方文档

- [MathJax 官方文档](https://docs.mathjax.org/) — 完整语法参考
- [LaTeX 数学符号大全](http://tug.ctan.org/info/symbols/comprehensive/symbols-a4.pdf) — PDF 速查表（约 14,000 个符号）

---

### 实用工具

- **[Detexify](http://detexify.kirelabs.org/classify.html)** — 画符号查找 LaTeX 命令
- **[MathJax 实时编辑器](https://www.mathjax.org/#demo)** — 在线测试公式

---

## 🔄 与其他功能配合

MathJax 可与以下功能无缝配合：

- ✅ **Mermaid 图表** — 同一笔记中同时使用
- ✅ **代码高亮** — 公式包围在代码块中也能渲染
- ✅ **主题切换** — 公式颜色适配当前主题
- ✅ **HTML 导出** — 导出后仍可正常显示

---

**🚀 开始书写您的数学笔记吧！**

如需更多 LaTeX 命令，请参考上方学习资源。多数常用数学符号已被覆盖，复杂公式可查阅 LaTeX 数学模式完整文档。
