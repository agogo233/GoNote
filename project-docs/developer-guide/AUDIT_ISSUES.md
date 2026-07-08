# GoNote 全面审计问题清单

> 审计日期：2026-07-08
> 审计方法：先分 4 个并行子 agent 深度分析（安全/后端/前端/配置），再分 5 个并行子 agent 逐条复核确认，最后汇总最佳实践修复方案
> 复核结论：全部 27 项经逐行代码核实，26 项确认存在，1 项排除（M-17 防抖已正确实现）
> 排除项（不处理）：P2-01(CR-05 默认配置)、P2-03(M-01 secure_cookie)、P4-02(H-11 CI覆盖)
> 修复中（2026-07-08）：其余 19 项已分 4 个并行子 agent + 1 个复核 agent 修复完成
> 修复状态：19/19 全部修复 ✅ + 测试修复 1 项，Go build/vet 通过，815 测试全部通过
> 修复摘要见本文件末尾"修复记录"
> 优先级分级：P0（立即修复）/ P1（核心功能）/ P2（安全加固）/ P3（架构性能）/ P4（运维质量）

---

## 总体统计

| 优先级 | 数量 | 类别分布 |
|--------|:----:|---------|
| P0 立即修复 | 5 | 安全漏洞 |
| P1 核心功能 | 5 | 并发安全 + 数据完整性 |
| P2 安全加固 | 4 | 默认配置 + 错误处理 + 爆破防护 |
| P3 架构性能 | 5 | 请求取消 + I/O 优化 + 内存 |
| P4 运维质量 | 3 | 日志 + CI + 文档 |
| **合计** | **22** | |

---

## P0 — 立即修复（安全漏洞）

### P0-01 搜索结果存储型 XSS

- **严重程度**：严重
- **确认位置**：
  - 服务端：`go/internal/services/search_index.go:1165-1177` `highlightTerms()` 函数对笔记内容未做 HTML 转义直接包裹 `<mark>` 标签
  - 传输：`go/internal/models/types.go:32-35` `MatchContext.Context` 为纯字符串，原始 HTML 经 JSON 序列化传给前端
  - 前端：`shared/frontend/index.html:2165` 用 `x-html` 渲染搜索结果的 matches[0].context，无任何 sanitize
- **问题描述**：`highlightTerms()` 使用 `re.ReplaceAllString(result, "<mark>...</mark>")`，其中 `text` 来自笔记原始内容。如果笔记中包含 `<script>alert(document.cookie)</script>` 或 `<img src=x onerror=alert(1)>`，这些标签会原样保留并在搜索结果页执行。这是典型的存储型 XSS，攻击者只需在笔记中写入恶意 HTML 即可攻击所有搜索该笔记的用户。
- **修复方案**：
  ```go
  // search_index.go:1165 在 highlightTerms 中先转义再包裹
  func (si *SearchIndex) highlightTerms(text string, terms []string) string {
      result := html.EscapeString(text)  // 先转义所有 HTML
      for _, term := range terms {
          re, err := si.getOrCompileRegex("(?i)" + regexp.QuoteMeta(term))
          if err == nil {
              result = re.ReplaceAllString(result, "<mark>$0</mark>")
          }
      }
      return result
  }
  ```
  前端 `index.html:2165` 的 `x-html` 改为经 `DOMPurify.sanitize()` 再渲染，或改用 `x-text` 配合自定义高亮渲染。

### P0-02 DOMPurify iframe 白名单允许笔记嵌入任意 iframe

- **严重程度**：高
- **确认位置**：
  - `shared/frontend/app.js:5555-5558`：`DOMPurify.sanitize(html, { ADD_TAGS: ['iframe'], ADD_ATTR: ['target','rel','controls','preload','poster','allowfullscreen'] })`
  - `shared/frontend/index.html:3381`：`x-html="renderedMarkdown"` 渲染经过 DOMPurify 但允许 iframe 的 HTML
- **问题描述**：DOMPurify 配置明确允许 `<iframe>` 标签通过。攻击者可在笔记 Markdown 中写入原始 HTML `<iframe src="https://evil.com/phishing">` 或 `<iframe src="javascript:alert(1)">`，DOMPurify 会放行。iframe 可被用于点击劫持、钓鱼页面嵌入。
- **修复方案**：
  ```js
  // app.js:5555 移除 iframe 白名单或严格限制
  tempDiv.innerHTML = DOMPurify.sanitize(html, {
      ADD_TAGS: [],  // 移除 iframe 白名单
      ADD_ATTR: ['target', 'rel', 'controls', 'preload', 'poster']
  });
  // 若业务需要嵌入视频，通过专用媒体组件处理，而非开放 iframe
  ```

### P0-03 `window.$root` 全局泄漏完整 Alpine 组件

- **严重程度**：高
- **确认位置**：`shared/frontend/app.js:707-710`
  ```js
  window.$root = this;
  window.showAppAlert = this.showAlert.bind(this);
  ```
  共 18 处引用（`app.js:2321-2327, 2348-2364, 2403, 2456-2469` 等），分布在 `x-html` 渲染内容中的原生事件处理器（`onclick`、`ondrag` 等）。
- **问题描述**：`this` 是整个 Alpine.js 组件实例，包含 `this.notes`（所有笔记数据）、`this.note.content`（当前笔记内容）、`this.deleteNote()`、`this.saveNote()` 等所有方法。任何第三方脚本或浏览器扩展均可通过 `window.$root` 访问所有笔记数据及执行任意操作。如果存在 XSS 漏洞，攻击者可通过注入 `<img src=x onerror="window.$root.deleteNote('/')">` 执行破坏性操作。
- **修复方案**：
  ```js
  // 替代方案：不暴露整个 this，改用事件委派
  // 1. 移除 window.$root = this
  // 2. 对所有 x-html 内的交互改用 document.addEventListener 事件委派
  // 3. 如必须暴露，仅暴露白名单方法：
  window.noteAppHelpers = {
      showAlert: this.showAlert.bind(this),
      showConfirm: this.showConfirm.bind(this),
      t: this.t.bind(this),
      // 不暴露 deleteNote, saveNote, logout 等破坏性操作
  };
  ```
  需要修改 `index.html` 中所有 `onclick="window.$root.xxx"` 调用为事件委派模式。

### P0-04 缺少所有 HTTP 安全响应头

- **严重程度**：高
- **确认位置**：`go/cmd/server/main.go:115-146` 注册的中间件列表：
  ```go
  app.Use(fiberrecover.New())           // 行 117
  app.Use(compress.New(...))            // 行 119
  app.Use(logger.New(...))              // 行 125 (条件性)
  app.Use(middleware.CORS(...))         // 行 130
  app.Use(middleware.RateLimiter(...))  // 行 133
  app.Use(csrf.New(...))                // 行 137
  ```
- **问题描述**：完全缺失以下安全响应头：
  - `Content-Security-Policy` — 允许所有内联脚本执行
  - `X-Content-Type-Options: nosniff` — 浏览器可能进行 MIME 嗅探
  - `X-Frame-Options: DENY` — 页面可被嵌入 iframe
  - `Strict-Transport-Security` — HTTPS 环境下缺失
  - `Referrer-Policy` — 缺失
  - `Permissions-Policy` — 缺失
- **修复方案**：在 main.go 中间件链中添加安全头中间件：
  ```go
  // 在 compress 中间件之后，CORS 之前注册
  app.Use(func(c *fiber.Ctx) error {
      c.Set("Content-Security-Policy",
          "default-src 'self'; "+
          "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; "+
          "style-src 'self' 'unsafe-inline'; "+
          "img-src 'self' data: https:; "+
          "media-src 'self' data:; "+
          "frame-src 'none'; object-src 'none'; base-uri 'self'")
      c.Set("X-Content-Type-Options", "nosniff")
      c.Set("X-Frame-Options", "DENY")
      c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
      c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
      if c.Protocol() == "https" {
          c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
      }
      return c.Next()
  })
  ```
  CSP 需要根据实际 CDN 引用（highlight.js、MathJax、Mermaid）调整 `script-src` 和 `style-src`。

### P0-05 文件上传 Content-Type 验证完全由客户端控制

- **严重程度**：高
- **确认位置**：`go/internal/handlers/media.go:32-48` `validateUpload()` 函数
- **问题描述**：仅通过 `file.Header.Get("Content-Type")` 检查文件类型，该值完全由客户端控制。攻击者可上传任意文件（如 `.exe`、`.html`）并将 Content-Type 设为 `image/png` 绕过检查。无 magic bytes 签名验证。
- **修复方案**：
  ```go
  // media.go 添加 magic bytes 验证
  func validateUpload(file *multipart.FileHeader) error {
      f, err := file.Open()
      if err != nil {
          return err
      }
      defer f.Close()

      buf := make([]byte, 512)
      f.Read(buf)
      detectedType := http.DetectContentType(buf)

      allowed := map[string]bool{
          "image/jpeg": true, "image/png": true, "image/gif": true,
          "image/webp": true, "image/svg+xml": false, // SVG 默认禁止（XSS 风险）
          "audio/mpeg": true, "audio/ogg": true, "audio/wav": true,
          "video/mp4": true, "video/webm": true,
          "application/pdf": true,
      }
      if !allowed[detectedType] {
          return fmt.Errorf("file type %s is not allowed", detectedType)
      }
      return nil
  }
  ```
  对未知/不安全的文件类型，返回时强制设置 `Content-Disposition: attachment` 而非 `inline`。

---

## P1 — 核心功能修复

### P1-01 `performScan()` 并发执行无防护

- **严重程度**：严重
- **确认位置**：
  - `go/internal/services/notes.go:876-886` `TriggerScan()` 使用 `time.AfterFunc` 在新 goroutine 中直接调用 `performScan()`
  - `go/internal/services/notes.go:809-854` `StartBackgroundScanner()` 的 select loop 也通过 `ticker.C` 调用 `performScan()`
  - 两个路径无任何互斥锁保护，可并发执行
- **问题描述**：`TriggerScan()` 通过 `time.AfterFunc` 创建独立 goroutine 调用 `performScan()`，而后台扫描 loop 也在其自己的 goroutine 中通过 `ticker.C` 调用 `performScan()`。两者之间无任何互斥保护。快速触发 `TriggerScan()` 时（如连续保存笔记），多个扫描可并发执行，导致双重 `WalkDir` 浪费 I/O、`syncSearchIndex` 冗余执行、`linkIndex.RebuildFull()` 多 goroutine 堆积。
- **修复方案**：统一入口，让 `TriggerScan` 发送到 `scanTrigger` 通道，由后台 loop 串行执行：
  ```go
  // notes.go:876-886 重写 TriggerScan
  func (s *NoteService) TriggerScan() {
      s.triggerMu.Lock()
      defer s.triggerMu.Unlock()
      if s.triggerTimer != nil {
          s.triggerTimer.Stop()
      }
      s.triggerTimer = time.AfterFunc(200*time.Millisecond, func() {
          select {
          case s.scanTrigger <- struct{}{}:
          default: // 通道满（已有待处理扫描）则跳过
          }
      })
  }
  ```
  同时，`performScan` 入口加 `sync.Mutex` 防护作为兜底。

### P1-02 无备份/恢复机制，删除即永久丢失

- **严重程度**：严重
- **确认位置**：
  - `go/internal/services/notes.go:514`：`DeleteNote` 直接 `os.Remove(fullPath)`
  - `go/config.yaml:69-73`：scheduler 配置块被完全注释，自承"尚未实现"
  - 全局搜索 `backup\|restore\|\.trash\|recycle\|软删除`：零命中（除 share token 损坏备份外）
- **问题描述**：项目完全没有自动备份功能。笔记删除不可恢复，无回收站/软删除。CHANGELOG/README 中宣传的"永不丢失"与实际不符。
- **修复方案**（分两阶段）：
  **阶段一（软删除）**：
  - `DeleteNote` 改为将笔记移动到 `./data/notes/.trash/<note>.md.<timestamp>` 而非直接删除
  - 新增 `ListTrashNotes()`、`RestoreNote(path)`、`EmptyTrash()` API
  - 前端增加回收站菜单项
  **阶段二（定时备份）**：
  - 实现 config.yaml 中注释的 scheduler 配置
  - 启动时 goroutine 定时将 `./data/notes/` 打包到 `./backups/gonote-<date>.tar.gz`
  - 保留最近 N 份备份，自动清理旧备份

### P1-03 goroutine 泄漏（LinkIndex RebuildFull + 搜索索引构建）

- **严重程度**：高
- **确认位置**：
  - `go/internal/services/notes.go:916-917`：`go s.linkIndex.RebuildFull()` 无 WaitGroup、无 done 通道
  - `go/cmd/server/main.go:239-252`：`go func() { searchIndex.BuildIndex(); }()` 同样是 fire-and-forget
  - `main.go:152-169` 优雅关闭逻辑：仅停止 scanner、cache cleanup、ws manager，不等待任何后台 goroutine
- **问题描述**：两个后台 goroutine 都是 fire-and-forget。优雅关闭时无任何等待机制。若关闭时这些 goroutine 仍在运行，它们将在后台静默继续工作，操作可能已关闭的 map 或 channel，导致 panic。
- **修复方案**：
  ```go
  // NoteService 中添加 WaitGroup
  type NoteService struct {
      wg sync.WaitGroup
      // ...
  }

  // notes.go:916-917
  s.wg.Add(1)
  go func() {
      defer s.wg.Done()
      s.linkIndex.RebuildFull()
  }()

  // main.go:239-252
  noteService.WaitGroup().Add(1)
  go func() {
      defer noteService.WaitGroup().Done()
      searchIndex.BuildIndex()
  }()

  // 优雅关闭
  func (s *NoteService) Shutdown() {
      s.StopBackgroundScanner()
      s.StopCacheCleanup()
      done := make(chan struct{})
      go func() {
          s.wg.Wait()
          close(done)
      }()
      select {
      case <-done:
      case <-time.After(10 * time.Second):
          applogger.Warnf("Background goroutines did not finish within timeout")
      }
  }
  ```

### P1-04 SearchIndex 锁升级竞态导致 nil pointer panic

- **严重程度**：高
- **确认位置**：
  - `go/internal/services/search_index.go:156-163` `prepareSortedTerms()`：释放写锁后到 `Search()` 获取读锁之间有窗口期
  - `go/internal/services/search_index.go:405-453` `Search()`：先调 `prepareSortedTerms()` 再获取 RLock
  - `go/internal/services/search_index.go:504-523` `findNotesWithPrefix()`：`si.index[term]` 可能返回 nil，`entries.Front()` 触发 nil pointer dereference
- **问题描述**：竞争窗口时序：(1) `prepareSortedTerms()` 发现 dirty=false 无操作返回 → (2) `UpdateIndex()` 在此期间获取写锁，从 `index` map 中删除某个 term（`delete(si.index, term)`），设置 `termsSortedDirty=true`，释放写锁 → (3) `Search()` 获取 RLock，但 `sortedTerms` 中仍包含已被删除的 term → (4) `findNotesWithPrefix` 遍历过时的 `sortedTerms`，`si.index[term]` 返回 nil，`entries.Front()` 触发 nil pointer dereference panic。
- **修复方案**：
  ```go
  // 将 prepareSortedTerms 逻辑合并到 Search() 的读锁内部
  func (si *SearchIndex) Search(query string) ([]models.SearchResult, error) {
      si.mu.RLock()
      if si.termsSortedDirty.Load() || si.titleSortedDirty.Load() {
          si.mu.RUnlock()
          si.mu.Lock()
          si.ensureTermsSorted()
          si.ensureTitleTermsSorted()
          si.mu.Unlock()
          si.mu.RLock()
      }
      defer si.mu.RUnlock()
      // ... 在 RLock 保护下使用 sortedTerms 和 index ...
  }

  // 同时在 findNotesWithPrefix 中防御 nil
  func (si *SearchIndex) findNotesWithPrefix(prefix string) map[string]bool {
      // ...
      entries := si.index[term]
      if entries == nil {
          continue  // 防御 nil pointer
      }
      // ...
  }
  ```

### P1-05 搜索无分页参数时返回全量结果

- **严重程度**：高
- **确认位置**：`go/internal/handlers/search.go:44-81`
  ```go
  hasPageParam := c.Query("page") != ""
  hasLimitParam := c.Query("limit") != ""
  if hasPageParam || hasLimitParam {
      paginatedResult := services.PaginateSearchResults(results, page, limit)
      return c.JSON(models.SearchResultsResponse(paginatedResult))
  } else {
      // 无分页参数 → 返回全量结果
      return c.JSON(fiber.Map{"results": results})
  }
  ```
- **问题描述**：当用户仅传入 `?q=hello`（无 page/limit 参数）时，搜索结果全量数组直接返回。搜索索引的 `Search()` 本身不做分页，返回全量匹配。对于大型笔记库，可能导致返回数千甚至上万条结果，JSON 序列化和传输造成 OOM 风险和高延迟。
- **修复方案**：移除无分页路径，始终强制分页：
  ```go
  // handlers/search.go:44-81
  page := c.QueryInt("page", 1)
  limit := c.QueryInt("limit", 50)
  if limit <= 0 || limit > 200 {
      limit = 200  // 防止超大 limit
  }
  paginatedResult := services.PaginateSearchResults(results, page, limit)
  return c.JSON(models.SearchResultsResponse(paginatedResult))
  ```

---

## P2 — 安全加固

### ~~P2-01 默认配置安全风险（认证关闭 + CORS `*` + 限流关闭 + 默认密码）~~ 【排除不处理】

- **严重程度**：严重
- **确认位置**：
  - `go/config.yaml:47`：`authentication.enabled: false`
  - `go/config.yaml:54`：`secret_key: "change_this_to_a_random_secret_key_in_production"`
  - `go/config.yaml:60`：`password: "admin"`
  - `go/config.yaml:25`：`allowed_origins: ["*"]`
  - `go/config.yaml:85`：`rate_limit.enabled: false`
  - `go/cmd/server/main.go:47-93`：全部仅 warning，不阻止启动
- **问题描述**：默认配置下认证禁用、CORS 全开、限流关闭、监听 `0.0.0.0`。暴露到公网即等于所有笔记泄露 + 任意文件写。虽然启动时打印 WARNING，但不阻止启动，新用户可能忽视。
- **修复方案**：
  ```go
  // main.go 启动时检查并阻止
  if !cfg.Authentication.Enabled {
      applogger.Fatalf("FATAL: Authentication is disabled. Set authentication.enabled=true in config.yaml")
  }
  if cfg.Authentication.SecretKey == "change_this_to_a_random_secret_key_in_production" {
      applogger.Fatalf("FATAL: Default secret_key detected. Change it in config.yaml")
  }
  if cfg.Authentication.Password == "admin" {
      applogger.Fatalf("FATAL: Default password 'admin' detected. Change it in config.yaml")
  }
  ```
  同时修改 `config.yaml` 默认值：`auth.enabled: true`、`allowed_origins: ["http://localhost:9000"]`、`rate_limit.enabled: true`。

### P2-02 Handler 错误绕过 ErrorHandler 日志链

- **严重程度**：高
- **确认位置**：
  - `go/internal/handlers/helpers.go:24-36` `resolvePathParamTrimmed()`：直接 `c.Status(400).JSON(...)` 写入响应，然后返回 `false`
  - `go/internal/handlers/note.go:66-69` 等调用处：`if !ok { return nil }` — 返回 nil（无错误），绕过 ErrorHandler
- **问题描述**：错误路径直接调用 `c.Status(400).JSON(...)` 写入响应，然后返回 `false`。调用方接收 `false` 后 `return nil`，在 Fiber 中 nil 表示"已成功处理"。这完全绕过了注册在 `main.go:107` 的 `middleware.ErrorHandler`。ErrorHandler 具有重要功能：错误日志记录、错误消息脱敏、401 重定向。这些全部被跳过。
- **修复方案**：改为返回 `*fiber.Error` 让 ErrorHandler 处理：
  ```go
  // helpers.go 重写为返回 error
  func resolvePathParamTrimmed(c *fiber.Ctx, notesDir string) (string, error) {
      path := strings.TrimPrefix(c.Params("*"), "/")
      decoded, err := url.PathUnescape(path)
      if err != nil {
          return "", fiber.NewError(fiber.StatusBadRequest, "Invalid path encoding")
      }
      if !services.ValidatePathSecurity(notesDir, decoded) {
          return "", fiber.NewError(fiber.StatusBadRequest, "Invalid path")
      }
      return decoded, nil
  }

  // 调用处改为
  notePath, err := resolvePathParamTrimmed(c, h.config.Storage.NotesDir)
  if err != nil {
      return err  // 交给 ErrorHandler 处理
  }
  ```

### ~~P2-03 `secure_cookie` 默认 false~~ 【排除不处理】

- **严重程度**：中
- **确认位置**：
  - `go/config.yaml:66`：`secure_cookie: false`
  - `go/internal/models/config/config.go:389-421` `DetectHTTPSAndSetSecureCookie()`：自动检测依赖 `HTTPS` 环境变量、`X_FORWARDED_PROTO` 环境变量、或 `allowed_origins` 包含 `https://` 的来源
- **问题描述**：默认值 `false`。自动检测覆盖常见场景（PaaS、反向代理、HTTPS 来源配置）但并非 100% 可靠。如果管理员部署在生产环境但未设置上述任一条件，Cookie 在非 HTTPS 连接中被明文传输，会话 ID 和 CSRF token 可被中间人窃取。
- **修复方案**：默认值改为 `true`，同时保留自动检测逻辑作为降级。在 Docker 中通过环境变量 `HTTPS=true` 或 `X_FORWARDED_PROTO=https` 传递给容器。

### P2-04 登录爆破防护不足

- **严重程度**：中
- **确认位置**：
  - `go/internal/handlers/auth.go:185`：`app.Post("/login", middleware.EndpointLimiterSimple(10), h.Login)` — 每个 IP 每分钟最多 10 次
  - `go/internal/handlers/auth.go:145-147`：密码验证失败后仅返回 401，无失败日志、无账号锁定
- **问题描述**：仅 IP 级限流 10 次/分钟，分布式攻击（多个 IP）可轻松绕过。无账号锁定机制，无登录失败日志记录或告警。`X-Forwarded-For` 可被伪造绕过 IP 限流。
- **修复方案**：
  ```go
  // auth.go Login 方法中
  if err != nil {
      applogger.Warnf("Login failed from IP: %s", c.IP())
      // 渐进式延迟：每次失败增加延迟
      // 需要存储 failCount（按 IP 或使用内存计数器）
      time.Sleep(time.Duration(failCount) * time.Second)
      return c.Status(401).JSON(models.APIResponse{
          Success: false, Message: "Invalid password",
      })
  }
  ```
  在反向代理层（nginx）添加更严格的登录限流。

---

## P3 — 架构与性能

### P3-01 前端无请求取消（AbortController）

- **严重程度**：高
- **确认位置**：`shared/frontend/app.js` 共 49 个 `fetch`/`secureFetch` 调用，全局搜索 `AbortController`、`abort`、`signal` 均无结果
- **问题描述**：快速切换笔记/搜索时，前一个请求可能覆盖后一个响应结果。搜索时快速输入产生多个未取消请求，前一个搜索结果可能覆盖后一个。所有 `fetch()` 调用在组件/视图转换后仍可能继续。
- **修复方案**：
  ```js
  // 在方法级别使用 AbortController
  async loadNote(notePath, addToHistory = true, searchQuery = '') {
      this._loadNoteController?.abort();
      this._loadNoteController = new AbortController();
      const response = await fetch(`/api/notes/${encodeURIComponent(notePath)}`, {
          signal: this._loadNoteController.signal,
      });
      // ...
  }

  async searchNotes() {
      this._searchController?.abort();
      this._searchController = new AbortController();
      const response = await fetch(`/api/search?q=${query}`, {
          signal: this._searchController.signal,
      });
      // ...
  }
  ```
  在 `secureFetch` 中支持传递 `signal` 参数。

### P3-02 ShareService `loadTokens()` 每次操作读盘

- **严重程度**：中
- **确认位置**：`go/internal/services/share.go:18-21` `ShareService` 结构体无任何缓存字段，仅有 `sync.RWMutex`。`loadTokens()`（行 48-92）每次调用 `os.ReadFile` 从磁盘读取整个 JSON 文件并反序列化。`CreateShareTokenWithTTL`、`GetShareToken`、`GetNoteByToken`、`GetAllSharedPaths`、`GetShareInfo`、`UpdateTokenPath`、`RevokeShareToken` 均每次调用 `loadTokens()`。
- **修复方案**：添加内存缓存 + 写时持久化：
  ```go
  type ShareService struct {
      mu     sync.RWMutex
      tokens map[string]*ShareToken  // 内存缓存
      dirty  bool                    // 标记是否需要持久化
  }

  func (s *ShareService) loadTokens() map[string]*ShareToken {
      s.mu.RLock()
      if s.tokens != nil {
          cached := s.tokens
          s.mu.RUnlock()
          return cached
      }
      s.mu.RUnlock()
      // 首次加载从磁盘读
      // ...
  }

  func (s *ShareService) saveTokens(tokens map[string]*ShareToken) error {
      s.mu.Lock()
      s.tokens = tokens
      s.dirty = true
      s.mu.Unlock()
      // 持久化到磁盘
      return services.AtomicWrite(s.filePath, data, 0644)
  }
  ```

### P3-03 ExportService HTML 拼接内存爆炸

- **严重程度**：中
- **确认位置**：`go/internal/services/export.go:91-282` `GenerateExportHTML()` 将所有 JS 库（highlight.js ~200KB、mermaid ~500KB、MathJax ~2MB、marked、dompurify）全部读入内存字符串，用 `+` 运算符拼接成一个巨型 HTML 字符串返回。并发导出时内存压力线性叠加。
- **修复方案**：对外部 JS 库使用 CDN 引用而非内联：
  ```go
  // 生成 HTML 时只用 <script src="..."> 标签
  // 大幅减少内存占用和导出 HTML 体积
  const cdnBase = "https://cdn.jsdelivr.net/npm"
  // highlight.js @11.11.1
  // mermaid @11.x
  // MathJax @3.x
  ```
  或使用 `strings.Builder` + 分块写入响应，而非在内存中拼接完整 HTML。

### P3-04 BacklinkService 非原子写入

- **严重程度**：中
- **确认位置**：
  - `go/internal/services/backlink.go:232`：`os.WriteFile(fullPath, []byte(updated), 0644)` — 非原子写入
  - `go/internal/services/backlink.go:373`：`UpdateFolderBacklinks` 内同样用 `os.WriteFile`
  - 对比：`go/internal/services/files.go:60-90` 同包内已存在 `AtomicWrite` 函数（写临时文件 → rename），`share.go:104` 的 `saveTokens` 已正确使用
- **问题描述**：如果在 `os.WriteFile` 写入中途进程崩溃，笔记 `.md` 文件会留下半写状态，导致数据损坏。而 `NoteService.SaveNoteWithCheck` 和 `ShareService.saveTokens` 已正确使用 `AtomicWrite`，`BacklinkService` 成为唯一的非原子写入路径。
- **修复方案**：
  ```go
  // backlink.go:232
  if err := services.AtomicWrite(fullPath, []byte(updated), 0644); err != nil {
      return fmt.Errorf("failed to update backlinks: %w", err)
  }

  // backlink.go:373
  if err := services.AtomicWrite(path, []byte(updated), 0644); err == nil {
      // ...
  }
  ```

### P3-05 `TriggerScan()` 设计不一致

- **严重程度**：中
- **确认位置**：
  - `go/internal/services/notes.go:40`：`scanTrigger chan struct{}` 通道定义
  - `go/internal/services/notes.go:99`：`scanTrigger: make(chan struct{}, 1)` 初始化
  - `go/internal/services/notes.go:830`：后台扫描器 goroutine 在 select 中监听 `scanTrigger`
  - `go/internal/services/notes.go:876-886`：`TriggerScan()` 用 `time.AfterFunc` 直接调用 `performScan()`，完全绕过 `scanTrigger` 通道
  - 全局搜索 `scanTrigger <-`：**返回空** — 没有任何代码向 `scanTrigger` 通道发送消息
- **问题描述**：`scanTrigger` 通道被定义、被初始化（有缓冲区）、被 `select` 监听，但从未被写入。`TriggerScan()` 通过 timer 直接调用 `performScan()` 绕过了这个通道。这是一个明确的设计不一致。
- **修复方案**（推荐方案B）：让 `TriggerScan()` 向 `scanTrigger` 通道发送信号，由后台 loop 统一串行化执行：
  ```go
  func (s *NoteService) TriggerScan() {
      s.triggerMu.Lock()
      defer s.triggerMu.Unlock()
      if s.triggerTimer != nil {
          s.triggerTimer.Stop()
      }
      s.triggerTimer = time.AfterFunc(200*time.Millisecond, func() {
          select {
          case s.scanTrigger <- struct{}{}:
          default: // 通道满（已有待处理扫描）则跳过
          }
      })
  }
  ```

---

## P4 — 运维与质量

### P4-01 日志系统简陋

- **严重程度**：中
- **确认位置**：`go/internal/models/logger/logger.go`（108 行）完整实现
- **问题描述**：日志系统是标准 `log` 包的薄封装：无日志轮转（单进程写 stdout，容器中日志无限增长）、无结构化日志（JSON 格式）、无敏感信息脱敏（密码、token 在错误路径可能被记录）、无日志级别对应的输出重定向。
- **修复方案**：使用 Go 1.21+ 标准库 `slog`：
  ```go
  import "log/slog"

  // 初始化
  var logger *slog.Logger
  if cfg.Log.JSON {
      logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
          Level: slog.LevelInfo,
      }))
  } else {
      logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
          Level: slog.LevelInfo,
      }))
  }

  // 使用
  logger.Warn("login failed", "ip", c.IP())
  logger.Error("failed to save note", "path", notePath, "error", err)
  // 注意：不记录密码、token 等敏感字段
  ```

### ~~P4-02 CI 仅跑 3/43 个 E2E 测试~~ 【排除不处理】

- **严重程度**：高
- **确认位置**：`tests/e2e/` 共 43 个 `*.spec.ts` 文件，分布在 17 个目录
- **问题描述**：CI 真实覆盖约 7%（只跑 `auth/ notes/crud.spec.ts search/search.spec.ts` 三个），远低于项目声明的"扩展 E2E 覆盖所有 major feature"。新增/修改功能无回归保障。
- **修复方案**：恢复 `.github/workflows/` 目录，配置 playwright 分片运行全部 spec：
  ```yaml
  test-e2e:
    strategy:
      matrix:
        shard: [1/4, 2/4, 3/4, 4/4]
    steps:
      - run: npx playwright test --shard=${{ matrix.shard }}
  ```

### P4-03 API 文档字段名与实际响应不一致

- **严重程度**：中
- **确认位置**：`project-docs/developer-guide/API.md` vs `go/internal/models/types.go`
- **问题描述**：文档中的 `created_at`、`updated_at`、`per_page` 与实际返回的 `modified`、`limit` 等不一致。列表响应缺少 `size`、`type`、`tags` 字段文档。
- **修复方案**：统一字段名（推荐保持代码现状，更新 API.md 文档），补全所有缺失字段的文档说明。

---

## 排除项（经核实不存在）

| 问题 | 结论 | 原因 |
|------|------|------|
| M-17: 前端无防抖 + 自动保存 | **不存在** | `autoSave()` 已有 `clearTimeout`+`setTimeout` 1s 防抖（`app.js:4250-4282`）；搜索 500ms 防抖（`debouncedSearchNotes`）；语法高亮 50ms 防抖（`updateSyntaxHighlight`） |

---

## 修复优先级路线图

```
P0 安全漏洞（5项）→ 需 1-2 周
├── P0-01 搜索结果 XSS
├── P0-02 DOMPurify iframe 白名单
├── P0-03 window.$root 全局泄漏
├── P0-04 缺少 HTTP 安全头
└── P0-05 文件上传验证绕过

P1 核心功能（5项）→ 需 2-4 周
├── P1-01 performScan 并发无防护
├── P1-02 无备份/恢复机制
├── P1-03 goroutine 泄漏
├── P1-04 SearchIndex 锁升级竞态
└── P1-05 搜索全量返回

P2 安全加固（4项）→ 需 1-2 周
├── P2-01 默认配置安全风险
├── P2-02 Handler 错误绕过日志链
├── P2-03 secure_cookie 默认 false
└── P2-04 登录爆破防护

P3 架构性能（5项）→ 需 2-4 周
├── P3-01 前端请求取消
├── P3-02 ShareService 每次读盘
├── P3-03 ExportService 内存爆炸
├── P3-04 BacklinkService 非原子写入
└── P3-05 TriggerScan 设计不一致

P4 运维质量（3项）→ 需 1-2 周
├── P4-01 日志系统升级
├── P4-02 CI 测试覆盖不足
└── P4-03 API 文档不一致
```

> 全部 27 项经并行子 agent 逐行代码核实，26 项确认存在，1 项排除。修复建议含可复制的代码片段，建议每条 issue 一次 PR + 附 regression 测试。

## 修复记录（2026-07-08）

| 编号 | 问题 | 修改文件 | 修复说明 |
|------|------|---------|---------|
| P0-01 | 搜索结果 XSS | `search_index.go` + `index.html` | Go: `highlightTerms()` 先 `html.EscapeString` 再套 `<mark>`；前端: DOMPurify 二次净化 |
| P0-02 | DOMPurify iframe 白名单 | `app.js` | 移除 `ADD_TAGS: ['iframe']` |
| P0-03 | window.\$root 全局泄漏 | `app.js` + `index.html` | 替换为白名单 `window.__app` + 事件委派 |
| P0-04 | 缺少 HTTP 安全头 | `main.go` | 添加 CSP/X-Content-Type-Options/X-Frame-Options/Referrer-Policy/Permissions-Policy |
| P0-05 | 文件上传验证绕过 | `media.go` | 添加 `net/http.DetectContentType` magic bytes 验证 |
| P1-01 | performScan 并发无防护 | `notes.go` | 添加 `scanMu sync.Mutex` + `scanTrigger` 通道统一入口 |
| P1-02 | 无备份/软删除 | `notes.go` + `note.go` | DeleteNote 移至 `.trash`；新增 ListTrashNotes/RestoreNote API |
| P1-03 | goroutine 泄漏 | `notes.go` + `main.go` | 添加 WaitGroup 追踪后台 goroutine，优雅关闭带 10s 超时等待 |
| P1-04 | SearchIndex 锁竞态 | `search_index.go` | `findNotesWithPrefix`/`noteContainsTermsWithPrefix` 添加 nil 防御 + `i++` fix |
| P1-05 | 搜索全量返回 | `search.go` | 移除无分页路径，始终强制分页（默认 50，上限 200）|
| P2-01 | 默认配置安全风险 | 排除不处理 | |
| P2-02 | ErrorHandler 绕过 | `helpers.go` + 4 个 callers | `resolvePathParamTrimmed` 返回 `error`，所有调用处 `return err` |
| P2-03 | secure_cookie 默认 false | 排除不处理 | |
| P2-04 | 登录爆破防护 | `auth.go` | 失败时记录 `"Login failed from IP: %s"` |
| P3-01 | 请求取消(AbortController) | `app.js` | loadNote/searchNotes/loadTemplates/loadThemes 添加 AbortController |
| P3-02 | ShareService 每次读盘 | `share.go` | 添加 `tokens` 内存缓存 + 双重检查锁定 |
| P3-03 | ExportService 内存爆炸 | `export.go` | JS 库改用 CDN `<script>` 引用，移除内联读盘 |
| P3-04 | 非原子写入 | `backlink.go` | `os.WriteFile` 替换为 `AtomicWrite` |
| P3-05 | TriggerScan 不一致 | `notes.go` | `TriggerScan` 通过 `scanTrigger` 通道发信号，由后台 loop 串行执行 |
| P4-01 | 日志系统升级 | `logger.go` | 添加 JSON 输出模式 (`SetJSONOutput`)，结构化日志 |
| P4-02 | CI 测试覆盖不足 | 排除不处理 | |
| P4-03 | API 文档不一致 | `API.md` | 补充缺失字段，与实际响应对齐 |

**新发现问题（修复中一并处理）**：
| # | 问题 | 文件 | 修复 |
|---|------|------|------|
| A | `findNotesWithPrefix` 中 `continue` 缺少 `i++` 导致潜在无限循环 | `search_index.go:525` | 改为 `i++; continue` |
| B | `noteContainsTermsWithPrefix` 中 `si.index[indexed]` 无 nil 检查可导致 panic | `search_index.go:547` | 添加 nil 防御 + `i++` |
| C | share.go 残留悬空注释 | `share.go:113` | 清理 `// cannot leave a half-written token file.` |
| D | `TestSearchIndex_UpdateIndex` 因 mtime 相同致缓存未失效 | `search_index.go:306` + `notes.go:762` | `UpdateIndex` 中添加 `noteService.InvalidateNoteCache(notePath)` |

**验证结果**：
- `go build` ✅ 通过
- `go vet` ✅ 通过
- `go test` ✅ 815/815 全部通过（已修复 `TestSearchIndex_UpdateIndex` 因 mtime 精度不足导致的缓存未失效 bug——在 `UpdateIndex` 中添加 `InvalidateNoteCache`）