# 🧮 LaTeX/MathJax 参考

GoNote 支持由 MathJax 3 驱动的 **LaTeX 数学符号**。使用熟悉的 LaTeX 语法在笔记中编写精美的公式。

## 语法概述

### 行内数学（文本中）
使用 `$...$` 表示行内公式：

- `$E = mc^2$` 渲染为：$E = mc^2$
- `$x^2 + y^2 = r^2$` 渲染为：$x^2 + y^2 = r^2$

### 独立数学（居中，单独一行）
使用 `$$...$$` 表示独立公式：

```markdown
$$
x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}
$$
```

$$
x = \frac{-b \pm \sqrt{b^2-4ac}}{2a}
$$

---

## 基本示例

### 上标和下标

**上标**使用 `^`：
- `$x^2$` → $x^2$
- `$e^{i\pi}$` → $e^{i\pi}$

**下标**使用 `_`：
- `$x_1$` → $x_1$
- `$a_{ij}$` → $a_{ij}$

**组合**：
- `$x_1^2$` → $x_1^2$
- `$\sum_{i=1}^{n} i^2$` → $\sum_{i=1}^{n} i^2$

### 分数

简单分数：`$\frac{a}{b}$` → $\frac{a}{b}$

复杂分数：

$$
\frac{\frac{1}{x}+\frac{1}{y}}{x+y} = \frac{x+y}{xy(x+y)} = \frac{1}{xy}
$$

### 平方根

- `$\sqrt{2}$` → $\sqrt{2}$
- `$\sqrt[3]{8}$` → $\sqrt[3]{8}$（立方根）
- `$\sqrt{x^2 + y^2}$` → $\sqrt{x^2 + y^2}$

---

## 希腊字母

### 小写
`$\alpha, \beta, \gamma, \delta, \epsilon, \zeta, \eta, \theta, \lambda, \mu, \pi, \sigma, \tau, \phi, \chi, \psi, \omega$`

$\alpha, \beta, \gamma, \delta, \epsilon, \zeta, \eta, \theta, \lambda, \mu, \pi, \sigma, \tau, \phi, \chi, \psi, \omega$

### 大写
`$\Gamma, \Delta, \Theta, \Lambda, \Xi, \Pi, \Sigma, \Phi, \Psi, \Omega$`

$\Gamma, \Delta, \Theta, \Lambda, \Xi, \Pi, \Sigma, \Phi, \Psi, \Omega$

---

## 微积分

### 积分

**定积分：**
```
$$
\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$
```

$$
\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

**多重积分：**
```
$$
\iiint_V f(x,y,z) \, dx \, dy \, dz
$$
```

$$
\iiint_V f(x,y,z) \, dx \, dy \, dz
$$

### 导数

**一阶导数：** `$\frac{df}{dx}$` → $\frac{df}{dx}$

**偏导数：** `$\frac{\partial f}{\partial x}$` → $\frac{\partial f}{\partial x}$

**梯度：**
```
$$
\nabla f = \frac{\partial f}{\partial x}\mathbf{i} + \frac{\partial f}{\partial y}\mathbf{j} + \frac{\partial f}{\partial z}\mathbf{k}
$$
```

$$
\nabla f = \frac{\partial f}{\partial x}\mathbf{i} + \frac{\partial f}{\partial y}\mathbf{j} + \frac{\partial f}{\partial z}\mathbf{k}
$$

### 极限

```
$$
\lim_{x \to \infty} \frac{1}{x} = 0
$$
```

$$
\lim_{x \to \infty} \frac{1}{x} = 0
$$

---

## 求和与乘积

### 求和

**行内：** $\sum_{i=1}^{n} i = \frac{n(n+1)}{2}$

**独立：**
```
$$
\sum_{k=1}^{\infty} \frac{1}{k^2} = \frac{\pi^2}{6}
$$
```

$$
\sum_{k=1}^{\infty} \frac{1}{k^2} = \frac{\pi^2}{6}
$$

### 乘积

```
$$
\prod_{i=1}^{n} i = n!
$$
```

$$
\prod_{i=1}^{n} i = n!
$$

---

## 矩阵与向量

### 基础矩阵

```
$$
\begin{bmatrix}
a & b \\\
c & d
\end{bmatrix}
$$
```

$$
\begin{bmatrix}
a & b \\\
c & d
\end{bmatrix}
$$

### 大型矩阵

```
$$
A = \begin{bmatrix}
1 & 2 & 3 \\\
4 & 5 & 6 \\\
7 & 8 & 9
\end{bmatrix}
$$
```

$$
A = \begin{bmatrix}
1 & 2 & 3 \\\
4 & 5 & 6 \\\
7 & 8 & 9
\end{bmatrix}
$$

### 单位矩阵

```
$$
I = \begin{pmatrix}
1 & 0 & 0 \\\
0 & 1 & 0 \\\
0 & 0 & 1
\end{pmatrix}
$$
```

$$
I = \begin{pmatrix}
1 & 0 & 0 \\\
0 & 1 & 0 \\\
0 & 0 & 1
\end{pmatrix}
$$

### 行列式

```
$$
\det(A) = \begin{vmatrix}
a & b \\\
c & d
\end{vmatrix} = ad - bc
$$
```

$$
\det(A) = \begin{vmatrix}
a & b \\\
c & d
\end{vmatrix} = ad - bc
$$

---

## 高级功能

### 方程组

```
$$
\begin{cases}
x + y = 5 \\\
2x - y = 1
\end{cases}
$$
```

$$
\begin{cases}
x + y = 5 \\\
2x - y = 1
\end{cases}
$$

### 对齐方程

```
$$
\begin{aligned}
f(x) &= (x+1)^2 \\\
&= x^2 + 2x + 1
\end{aligned}
$$
```

$$
\begin{aligned}
f(x) &= (x+1)^2 \\\
&= x^2 + 2x + 1
\end{aligned}
$$

### 连分式

```
$$
\phi = 1 + \frac{1}{1 + \frac{1}{1 + \frac{1}{1 + \cdots}}}
$$
```

$$
\phi = 1 + \frac{1}{1 + \frac{1}{1 + \frac{1}{1 + \cdots}}}
$$

---

## 数学符号

### 运算符

| 符号 | LaTeX | 结果 |
|--------|-------|--------|
| 加减 | `$\pm$` | $\pm$ |
| 乘 | `$\times$` | $\times$ |
| 除 | `$\div$` | $\div$ |
| 不等 | `$\neq$` | $\neq$ |
| 小于/大于 | `$\leq, \geq$` | $\leq, \geq$ |
| 约等 | `$\approx$` | $\approx$ |
| 无穷 | `$\infty$` | $\infty$ |

### 集合论

| 符号 | LaTeX | 结果 |
|--------|-------|--------|
| 属于 | `$\in$` | $\in$ |
| 不属于 | `$\notin$` | $\notin$ |
| 子集 | `$\subset$` | $\subset$ |
| 并集 | `$\cup$` | $\cup$ |
| 交集 | `$\cap$` | $\cap$ |
| 空集 | `$\emptyset$` | $\emptyset$ |

### 逻辑

| 符号 | LaTeX | 结果 |
|--------|-------|--------|
| 与 | `$\land$` | $\land$ |
| 或 | `$\lor$` | $\lor$ |
| 非 | `$\neg$` | $\neg$ |
| 蕴含 | `$\implies$` | $\implies$ |
| 当且仅当 | `$\iff$` | $\iff$ |
| 任意 | `$\forall$` | $\forall$ |
| 存在 | `$\exists$` | $\exists$ |

---

## 著名公式

### 欧拉恒等式

$$ e^{i\pi} + 1 = 0 $$

### 爱因斯坦质能方程

$$ E = mc^2 $$

### 勾股定理

$$ a^2 + b^2 = c^2 $$

### 薛定谔方程

$$ i\hbar\frac{\partial}{\partial t}\Psi(\mathbf{r},t) = \hat{H}\Psi(\mathbf{r},t) $$

### 麦克斯韦方程组

```
$$
\begin{aligned}
\nabla \cdot \mathbf{E} &= \frac{\rho}{\epsilon_0} \\\
\nabla \cdot \mathbf{B} &= 0 \\\
\nabla \times \mathbf{E} &= -\frac{\partial \mathbf{B}}{\partial t} \\\
\nabla \times \mathbf{B} &= \mu_0\mathbf{J} + \mu_0\epsilon_0\frac{\partial \mathbf{E}}{\partial t}
\end{aligned}
$$
```

$$
\begin{aligned}
\nabla \cdot \mathbf{E} &= \frac{\rho}{\epsilon_0} \\\
\nabla \cdot \mathbf{B} &= 0 \\\
\nabla \times \mathbf{E} &= -\frac{\partial \mathbf{B}}{\partial t} \\\
\nabla \times \mathbf{B} &= \mu_0\mathbf{J} + \mu_0\epsilon_0\frac{\partial \mathbf{E}}{\partial t}
\end{aligned}
$$

---

## 提示

### 1. 预览模式
始终使用**分栏视图**或**预览模式**实时查看公式渲染效果。

### 2. 转义美元符号
如果需要字面美元符号（非数学），请转义：`$\\$100$` 渲染为 $\\$100$

### 3. 复杂表达式
对于非常长的公式，考虑使用 `aligned` 或 `split` 环境分成多行。

### 4. 矩阵和多行格式化
**重要**：使用 **3 个反斜杠 + 空格**（`\\\ `）换行以启用多行格式化：

```markdown
✅ 良好（可读的多行格式）：
$$
\begin{bmatrix}
a & b \\\
c & d
\end{bmatrix}
$$

❌ 错误（只有 2 个反斜杠 - 不工作）：
$$
\begin{bmatrix}
a & b \\
c & d
\end{bmatrix}
$$
```

**秘诀：** 在每行末尾使用 `\\\ `（三个反斜杠 + 尾部空格），然后添加换行。这样可以实现可读的多行公式！

### 5. 调试
如果公式未渲染：
- 检查匹配的分隔符（`$...$` 或 `$$...$$`）
- 确保反斜杠正确（`\frac` 不是 `/frac`）
- 查找未转义的特殊字符
- 对于矩阵/换行，使用 `\\\ `（三个反斜杠 + 空格）而不是 `\\`
- 确保 `\\\` 后面有空格再换行

### 6. 性能
MathJax 渲染效率高，但公式密集的笔记（100+ 公式）可能需要一点时间排版。

---

## 资源

更多 LaTeX 命令和符号，请参阅：
- [MathJax 文档](https://docs.mathjax.org/)
- [LaTeX 数学符号](http://tug.ctan.org/info/symbols/comprehensive/symbols-a4.pdf)
- [Detexify](http://detexify.kirelabs.org/classify.html)——绘制符号查找对应 LaTeX 命令

---

💡 **提示：** 复制此笔记中的示例快速开始在您自己的笔记中使用数学！
