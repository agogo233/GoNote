# 🧪 测试指南

本文档介绍 GoNote 项目的测试体系，包括 Playwright E2E 测试和 Go 单元测试。

---

## 📂 测试结构

```
tests/
└── e2e/                          # Playwright E2E 测试
    ├── auth/                     # 认证相关测试
    ├── bugs/                     # Bug 回归测试
    ├── encoding-fix/             # 字符编码测试
    ├── export/                   # 导出功能测试
    ├── folders/                  # 文件夹操作测试
    ├── graph/                    # 知识图谱测试
    ├── i18n/                     # 国际化测试
    ├── media/                    # 媒体文件测试
    ├── mobile/                   # 移动端响应式测试
    ├── notes/                    # 笔记 CRUD 测试
    ├── outline/                  # 大纲导航测试
    ├── search/                   # 搜索功能测试
    ├── security/                 # 安全测试
    ├── share/                    # 分享功能测试
    ├── shortcuts/                # 键盘快捷键测试
    ├── statistics/               # 统计功能测试
    ├── tags/                     # 标签功能测试
    ├── templates/                # 模板功能测试
    ├── themes/                   # 主题测试
    ├── view-modes/               # 视图模式测试
    └── homepage-fix.spec.ts      # 首页测试

go/internal/**/*_test.go          # Go 单元测试（与源码同目录）
```

---

## 🔬 测试类型

### E2E 测试（Playwright）

- **浏览器自动化** — 使用 Playwright 控制真实浏览器
- **全栈测试** — 覆盖前端 UI 和 API 层
- **真实用户场景** — 模拟实际使用流程
- **跨浏览器** — 支持 Chromium、Firefox、WebKit

### Go 单元测试

位于 `go/internal/**/*_test.go`：

- **Handler 测试** — HTTP 请求/响应验证
- **Service 测试** — 业务逻辑单元测试
- **Model 测试** — 数据结构和模型验证
- **Utility 测试** — 辅助函数测试

---

## 🚀 运行测试

### 环境准备

**必要条件：**

- Node.js 18+
- Playwright 浏览器
- Go 1.24+
- GoNote 服务器运行（E2E 测试需要）

**首次安装 Playwright 浏览器：**

```bash
npx playwright install

# 安装所有浏览器（可选）
npx playwright install --all
```

---

### 运行 E2E 测试

```bash
# 运行所有 E2E 测试（推荐先用此命令）
make test-e2e

# 或手动运行
npx playwright test
```

**常用选项：**

```bash
# UI 模式（交互式调试）
npx playwright test --ui

# 运行特定测试文件
npx playwright test tests/e2e/notes/create-note.spec.ts

# 运行特定标签的测试
npx playwright test --grep @smoke

# 运行特定浏览器
npx playwright test --project=chromium

# 查看 HTML 测试报告
npx playwright test --reporter=html
npx playwright show-report
```

---

### 运行 Go 单元测试

```bash
# 运行所有测试（推荐）
make test

# 手动运行（注意 STORAGE_NOTES_DIR 环境变量）
cd go
STORAGE_NOTES_DIR=../data go test ./...

# 竞争检测
go test ./... -race

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行特定包
go test ./internal/handlers/... -v

# 运行基准测试
go test -bench=. ./...
```

**注意：** Makefile 会自动设置 `STORAGE_NOTES_DIR=../data`，确保测试数据写入项目根目录的 `data/` 目录。

---

## 📝 编写测试

### E2E 测试示例（Playwright）

```typescript
import { test, expect } from '@playwright/test';

test('should create a new note', async ({ page }) => {
  await page.goto('http://localhost:9000');

  // 点击新建笔记按钮
  await page.click('[data-testid="new-note-btn"]');

  // 输入标题
  await page.fill('#note-title', 'Test Note');

  // 验证笔记已创建
  await expect(page.locator('.note-title')).toHaveText('Test Note');
});
```

**最佳实践：**

- 使用 `[data-testid]` 属性作为选择器（更稳定）
- 每个测试独立且可重复
- 测试前清理状态（夹具全局清理）
- 使用 `page` fixture 确保浏览器上下文隔离

---

### Go 测试示例

```go
func TestNoteHandler_GetNote(t *testing.T) {
    // 准备：创建 service 和 handler
    service := services.NewNoteService(notesDir)
    handler := handlers.NewNoteHandler(service, cfg)

    // 创建模拟请求
    req := httptest.NewRequest("GET", "/notes/test.md", nil)
    w := httptest.NewRecorder()

    // 执行：调用处理器
    handler.GetNote(w, req)

    // 验证：检查状态码
    if w.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", w.Code)
    }

    // 验证：解析响应
    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    if !resp["success"].(bool) {
        t.Error("expected success=true")
    }
}
```

**最佳实践：**

- 使用表驱动测试（table-driven tests）
- 模拟依赖（通过 interface）
- 测试边界条件和错误情况
- 保持测试独立（无共享状态）

---

## 📊 测试覆盖率

### E2E 测试覆盖范围

| 功能模块 | 覆盖率 |
|---------|--------|
| 认证 | ✅ |
| 笔记 CRUD | ✅ |
| 搜索 | ✅ |
| 标签 | ✅ |
| 模板 | ✅ |
| 图谱视图 | ✅ |
| 分享 | ✅ |
| 主题 | ✅ |
| 移动端 | ✅ |
| 安全 | ✅ |

---

### Go 代码覆盖率

```bash
# 生成覆盖率报告
cd go
go test ./... -coverprofile=coverage.out

# 查看 HTML 报告
go tool cover -html=coverage.out
```

---

## 🔄 持续集成

CI 配置位于 `.github/workflows/`：

- **每 push 到 main 分支** — 运行所有测试
- **每 PR** — 运行测试、linter、构建检查
- **每 release tag** — 运行所有检查 + 发布 Docker 镜像

### 本地模拟 CI

```bash
# 安装依赖
make deps

# 运行所有测试（像 CI 一样）
make test              # Go 单元测试
cd .. && npx playwright test  # E2E 测试
cd go && go vet ./...  # 静态检查
```

---

## 🧩 测试夹具（Fixtures）

测试数据位于 `tests/e2e/fixtures/`：

- **笔记样本** — 预创建的测试笔记（各种格式）
- **测试图片** — 用于媒体上传测试
- **配置** — 测试专用 `config.yaml`
- **模拟数据** — 特定场景数据

**全局清理：** `globalTeardown.ts` 负责所有测试完成后清理测试数据。

---

## 🐛 调试测试

### E2E 测试调试

```bash
# 启用慢动作播放
npx playwright test --debug

# 运行特定测试并显示日志
npx playwright test tests/e2e/notes/test.spec.ts --debug

# 使用 Playwright Inspector
PWDEBUG=1 npx playwright test

# 失败时自动截图（在 playwright.config.ts 中配置）
```

---

### Go 测试调试

```bash
# 详细输出
go test -v ./...

# 运行单个测试
go test -v -run TestNoteHandler_Get ./...

# 使用 Delve 调试器
dlv test ./internal/handlers
```

---

## ✅ 最佳实践

### 通用原则

1. **测试隔离** — 每个测试独立运行，不依赖其他测试状态
2. **描述性名称** — 如 `TestCreateNote_InvalidPath_ReturnsError`
3. **Arrange-Act-Assert** — 遵循 AAA 模式组织代码
4. **清理数据** — 测试后删除创建的笔记/文件
5. **有意义的断言** — 测试行为而非实现细节
6. **Page Object 模式** — E2E 使用页面对象封装 UI 交互
7. **测试数据固定** — 使用 fixtures 保证一致性

---

### 测试稳定性

**避免 Flaky Tests（随机失败）：**

- 使用 `await expect(...).toBeVisible()` 而非硬编码等待
- 确保测试间无状态残留
- 使用 `--retries`（CI）或 `test.describe.configure({ mode: 'serial' })`

**处理异步操作：**

- Playwright 自动等待元素，必要时 `await page.waitForLoadState()`
- Go 中使用 `t.Parallel()` 并行化独立测试

---

## 📚 相关文档

- [贡献指南](../../CONTRIBUTING.md）— PR 要求和测试准则
- [API 文档](../developer-guide/API.md）— 被测试的端点
- [安全专项](../../security/）— 安全测试说明

---

**💡 提交前检查：**

1. ✅ 新功能包含相应的 E2E 或单元测试
2. ✅ 所有测试通过（`make test && npx playwright test`）
3. ✅ Bug 修复包含回归测试防止问题重现

Happy testing! 🎉
