# GoNote 代码审计问题清单

> 审计日期：2026-07-03
> 审计方法：4 个并行子 agent 深度审查 + 2 个验证 agent 逐条复核真实代码
> 复核结论：25 项断言中 24 项完全成立，1 项部分成立（I-12 分母口径修正）
> 优先级分级：【严重】【重要】【建议】
>
> ## 修复状态
> 已修复（commit 1d9b898）：S-4, S-5, S-6, S-7, S-8, S-9 — 编译零错误，799 测试全量通过
> 已修复（commit 82a22d5）：I-1（per-path mutex + mtime 乐观锁 + 409 冲突响应）、I-15（多槽草稿 + dirty 追踪 + beforeunload + 409 冲突 banner）— 编译零错误，799 测试全量通过
> 已修复（commit 5815276）：I-6（原子写）、I-7（路径校验统一）、I-8（共享 token TTL + 原子写 + 坏文件备份告警）— 编译零错误，全量测试通过
> 已修复（commit 3df026c）：I-9（scannerDone 超时等待）、S-10（/healthz + /readyz 拆分，notes_dir/scanner/search index 三项检查）— 编译零错误，全量测试通过
> 已修复（commit 2f2dd77）：I-2（Cache.Get RLock 优化）、I-3（合并双 tag 缓存）、I-10（反向索引 + 锁外读盘）— 编译零错误，全量测试通过
> 已修复（commit 75ce890）：I-4（Fiber ProxyHeader 使 c.IP() 支持 X-Forwarded-For）、I-5（WebSocket Origin 校验 + ReadLimit/ReadDeadline + 连接上限）— 编译零错误，811 测试全量通过
> 已修复（commit fb0f1f4）：S-2（EndpointLimiter 与全局开关解耦）、S-3（共享页 title EscapeString + DOMPurify 清洗 marked 输出）— 编译零错误，811 测试全量通过
> 待修复：S-1, S-11, S-12, I-11~I-14, I-16, W-1~W-12

## 复核结果总览

| 类别 | 已核验 | 完全成立 | 部分成立 | 不成立 |
|------|:---:|:---:|:---:|:---:|
| 安全/稳定类 | 11 | 11 | 0 | 0 |
| 架构/性能/功能类 | 14 | 13 | 1 | 0 |
| **合计** | **25** | **24** | **1** | **0** |

所有断言均经实证取证，含 `file:line` + 代码片段证据。本文件为后续修复的权威依据。

---

## 一、【严重】共 12 项

### S-1 默认无认证 + 监听 0.0.0.0，公网部署完全开放

- **状态**：确认成立
- **证据**：
  - `go/config.yaml:21` `host: "0.0.0.0"`
  - `go/config.yaml:65` `enabled: false`
  - `go/internal/middleware/auth.go:13-15` `if !authEnabled { return c.Next() }`
  - `go/cmd/server/main.go:281` `api := app.Group("/api", middleware.AuthRequired(cfg.Authentication.Enabled))`
- **影响**：默认配置下所有 `/api/*`（读写、删除、搜索、媒体、分享）对任意访客开放。公网暴露即等于全部笔记泄露+任意文件写。
- **修复建议**：`authentication.enabled` 默认改为 `true`；自检默认密码/secret_key 仍为占位值时 `Fatal` 拒绝启动；绑定 `0.0.0.0` 且 `enabled=false` 时拒绝启动。

### S-2 速率限制默认关闭，登录爆破/API 滥用无防护 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：
  - `go/config.yaml:103` `enabled: false`
  - `go/internal/middleware/limiter.go:35-42`
    ```go
    if cfg != nil && !cfg.RateLimit.Enabled {
        return func(c *fiber.Ctx) error { return c.Next() }
    }
    ```
  - `go/cmd/server/main.go:274` `app.Post("/login", middleware.EndpointLimiterSimple(10), authHandler.Login)`
- **影响**：`EndpointLimiter` 在全局关闭时短路为 noop，`/login` 等敏感端点完全无限流，可暴力破解默认密码 `admin`。
- **修复说明**：`EndpointLimiter` 移除全局开关检查，始终返回真实限流器。`/login`、`/upload`、API 保存端点等独立于全局开关强制执行端点级限流。`TestEndpointLimiterWithGlobalDisabled` 验证全局禁用下端点限流仍生效。

### S-3 共享页存储型 XSS（title 未转义 + marked 默认渲染 HTML） ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：
  - `go/internal/services/export.go:107` `<title>` + title + `</title>`（title 未 `html.EscapeString`）
  - `go/internal/services/export.go:262-264`
    ```js
    document.getElementById('content').innerHTML = marked.parse(markdown);
    ```
  - `marked.setOptions` 仅设 `gfm/breaks/headerIds/mangle`，无 `sanitize`/`sanitizer`
  - escapeJS 仅转义 `\\\"`、`\n\r\t`、`</`，不阻止 `<script>...</script>`
- **影响**：`/share/:token` 公开头，笔记标题含 `</title><script>...</script>` 或 markdown 含 `<img onerror>` 即可执行任意 JS。
- **修复说明**：① title 使用 `html.EscapeString` 转义；② 共享页内联 DOMPurify（`shared/frontend/libs/dompurify/3.2.4/purify.min.js`），`marked.parse` 输出经 `DOMPurify.sanitize()` 清洗后注入。

### S-4 首次后台扫描 panic 后 ready 永不关闭 → 所有 list 请求永久阻塞 ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：`close(s.ready)` 移入 `defer`（无论如何都关闭）；`WaitForReady` 加 30s 超时 fallback
- **证据**：`go/internal/services/notes.go:597-607`
  ```go
  func() {
      defer func() { if r := recover(); r != nil { ... } }()
      s.performScan()
      close(s.ready) // panic 时此行不执行
  }()
  ```
- **影响**：`WaitForReady`（notes.go:691-696 `<-s.ready`）永久阻塞，所有 list/search/tag/graph/backlink 请求挂死，服务"假活"。
- **修复建议**：用 `defer close(s.ready)` 确保无论如何关闭；或在 recover 中也关闭；`WaitForReady` 加 30s 超时 fallback。

### S-5 Fiber 未注册 recover 中间件 ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：注册 `fiberrecover.New()` 在中间件链最前；import 用别名避免遮蔽内置 `recover()`
- **证据**：
  - `go/cmd/server/main.go:3-25` import 块无 `fiber/v2/middleware/recover`
  - `go/cmd/server/main.go:90-117` 注册顺序：compress→logger→CORS→RateLimiter→csrf，无 recover
- **影响**：任一 handler panic（断言失败、nil 解引用）会冒泡到 ErrorHandler 导致连接异常断开。
- **修复建议**：`app.Use(recover.New())` 显式注册。

### S-6 后台扫描不更新搜索索引，外部写入的笔记搜索不到 ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：NoteService 注入 searchIndex 引用；performScan 末尾新增 syncSearchIndex()，基于 mtime 快照做增量同步（新增/变更→UpdateIndex，消失→RemoveFromIndex）
- **证据**：
  - `go/internal/services/notes.go:668-678` `performScan` 仅调 `scanAndUpdate`（设 list cache）+ `onScanComplete`，无 `searchIndex.UpdateIndex`
  - `UpdateIndex` 仅在 handlers/note.go:152,192,236 经 HTTP API 调用
  - `NoteService` 结构体（notes.go:25-34）无 `searchIndex` 字段；backlink 独立于 NoteService（main.go:235）
- **影响**：外部编辑器/Sync 写入的笔记搜索索引永不更新，重启前永久缺失；用户通过搜索查不到外部新增内容。
- **修复建议**：`NoteService` 注入 `SearchIndex` 引用；`performScan` 检测 mtime 变化或新文件时调 `UpdateIndex`；或引入 fsnotify watch。

### S-7 搜索实现退化为 O(全 term 线性扫描) ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：新增 sortedTerms/sortedTitleTerms 有序切片 + atomic.Bool dirty 标志；findNotesWithPrefix/noteContainsTermWithPrefix/searchTitleByPrefix 改用 sort.SearchStrings 二分定位前缀区间；查询入口 prepareSortedTerms 持写锁重建一次后转 RLock
- **证据**：`go/internal/services/search_index.go:446-485`
  ```go
  func (si *SearchIndex) findNotesWithPrefix(prefix string) map[string]bool {
      for term, entries := range si.index {  // 遍历整个 map
          if strings.HasPrefix(term, prefix) {...}
      }
  }
  ```
  `noteContainsTermsWithPrefix`（L462-485）嵌套 O(terms × index size)，持 RLock
- **影响**：倒排索引的 O(1) 前缀查找能力未利用；CJK bigram 使 term 量级膨胀，万级笔记搜索延迟陡增并阻塞其它读。
- **修复建议**：用有序 term 切片 + `sort.Search` 做前缀二分，或 Trie/前缀树；先拿倒排 postings 再求交集，不重新 tokenize。

### S-8 MoveNote 先更新 backlinks 再 rename，失败则链接全断 ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：MoveNote/MoveFolder 统一改为"先 rename 成功 → 再改 backlinks → 失败回滚 rename"；MoveFolder 新增 UpdateFolderBacklinks 调用（此前完全未更新 folder 内 wikilinks）
- **证据**：`go/internal/services/notes.go:310-339`
  ```go
  backlinkService.UpdateAllBacklinks(oldPath, newPath)  // 改写其它 .md
  s.invalidateNoteCache(oldPath); s.invalidateNoteCache(newPath)
  return os.Rename(oldFull, newFull)                    // 失败则已无可回滚
  ```
- **影响**：跨设备/权限/目标已存在导致 Rename 失败时，其它笔记已把 `[[old]]` 改为 `[[new]]`，源文件仍在 old，反向链接不可恢复地错误。
- **修复建议**：调整为"先 Rename 成功再 UpdateAllBacklinks"；或失败时记录反向操作并补偿回滚；`MoveFolder`（folder.go:113-118）同一问题。

### S-9 Docker 容器以 root 运行 ✅已修复（commit 1d9b898）

- **状态**：确认成立 → 已修复
- **修复说明**：Dockerfile 运行阶段创建 gonote 用户（uid 10001）+ `USER gonote` + chown /app；docker-compose 显式 `user: "10001:10001"` 双保险 + 数据卷权限说明
- **证据**：`docker/go/Dockerfile` 全文无 `USER` 指令
- **影响**：容器逃逸即 root；data 卷文件属主 root，宿主调权限麻烦。
- **修复建议**：运行阶段创建 `gonote` 用户（uid 10001），`USER gonote`，chown `/app/data`。

### S-10 `/health` 不检查任何依赖，scanner 死锁时仍返回 200 ✅已修复

- **状态**：确认成立 → 已修复
- **修复说明**：拆分 `/healthz`（存活探针，保持原逻辑）与 `/readyz`（就绪探针，检查三项：notes_dir 可写 + scanner 首次扫描完成 + search index 已构建）；`SearchIndex` 新增 `buildDone atomic.Bool` + `IsReady()`，`NoteService` 新增 `IsScannerReady()`；Docker compose healthcheck 改用 `/readyz`，检测到依赖未就绪时返回 503
- **证据**：`go/internal/handlers/system.go:22-29`
  ```go
  return c.JSON(models.HealthResponse{
      Status: "ok", App: h.config.App.Name, Version: h.config.App.Version,
  })
  ```

### S-11 前端 Markdown 渲染不 sanitize，存在 XSS

- **状态**：确认成立
- **证据**：
  - `shared/frontend/app.js:5263-5278` `marked.setOptions({ breaks, gfm, renderer, tokenizer, highlight })` 无 sanitize
  - `grep DOMPurify shared/frontend` 零命中
  - `app.js:5286` `tempDiv.innerHTML = html` 直接注入 marked 输出
- **影响**：笔记内容含 `<script>`、`<img onerror>` 等原始 HTML 会被浏览器执行。
- **修复建议**：引入 DOMPurify 在 `marked.parse` 后清洗：`tempDiv.innerHTML = DOMPurify.sanitize(marked.parse(content))`。

### S-12 回收站/版本历史/定时备份声称但完全未实现

- **状态**：确认成立
- **证据**：
  - `go/internal/services/notes.go:292-307` `DeleteNote` 直接 `os.Remove(fullPath)`
  - `grep recycle|trash|softDelete|deleted_at|tombstone go/internal` 零命中
  - `go/config.yaml:86-91` scheduler 注释自承"尚未实现"
  - `config.go` Config 结构体无 Scheduler 字段
- **影响**：删除即永久丢失；CHANGELOG/README 的"永不丢失"宣传与实际不符。
- **修复建议**：实现软删除（移动到 `.trash/`，提供 `/api/notes/trash` 列出恢复端点）；或在文档显式声明"删除不可恢复"。

---

## 二、【重要】共 16 项

### I-1 同一笔记并发写入无锁，最后写胜出 ✅已修复（commit 82a22d5）

- **状态**：确认成立 → 已修复
- **修复说明**：NoteService 新增 `pathMu sync.Map` 惰性建锁 + `SaveNoteWithCheck(path, content, knownMtime)`；handler 扩展请求体 `modified` 字段，冲突返回 409 + 服务端 mtime；`NoteSaveResponse.Modified` 返回保存后权威 mtime；前端 `loadNote` 存服务端 mtime，`saveNote` 携带并在成功后用服务端值替换；旧前端不传 modified 时跳过乐观锁保持向后兼容
- **证据**：`go/internal/services/notes.go:262-289` SaveNote 无 mutex/版本号/ETag 校验
- **修复建议**：per-path mutex（`sync.Map[string]*sync.Mutex`）串行化写；保存时校验客户端携带的 mtime，冲突返回 409。

### I-2 Cache.Get 全程写锁，热路径串行 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：`go/internal/services/cache.go:96-118` `c.mu.Lock()` 而非 RLock（因需 MoveToBack）
- **影响**：list notes 高频请求在 Cache.Get 全局串行；QPS 高时锁竞争。
- **修复说明**：Get 方法改为 RLock 读路径，仅过期条目删除时持写锁；去掉读路径中的 MoveToBack（对 note 缓存影响极小，LRU 在 Set 时仍正常维护），并发 Get 不再串行阻塞

### I-3 双 tag 缓存重复实现 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：
  - `go/internal/services/notes.go:28` `tagCache *Cache`（基于 services.Cache LRU+TTL）
  - `go/internal/services/tags.go:16` `tagCache map[string]models.TagCacheEntry`（独立第二套 + RWMutex）
- **修复说明**：删除 TagService 的 `tagCache`/`tagMutex` 字段，`GetTagsCached` 和 `ClearCache` 委托 NoteService 的统一缓存，消除重复

### I-4 IP 限流键不走 X-Forwarded-For，反代后失效 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：`go/internal/middleware/limiter.go:24,48` `return c.IP()`（走 RemoteAddr）
- **修复说明**：ServerConfig 新增 `proxy_header`/`trusted_proxy_check`/`trusted_proxies` 配置；main.go 创建 Fiber 时应用 `fiber.Config{ProxyHeader, EnableTrustedProxyCheck, TrustedProxies, EnableIPValidation}`；`c.IP()` 在配置后自动读取 `X-Forwarded-For`；env 覆盖 `PROXY_HEADER`/`TRUSTED_PROXY_CHECK`/`TRUSTED_PROXIES`

### I-5 WebSocket 无读超时/消息上限/Origin 校验 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：
  - `go/internal/handlers/websocket.go` 全文无 `SetReadDeadline`/`SetReadLimit`/`CheckOrigin`
  - `main.go:262-269` `for { c.ReadJSON(&msg) }` 无超时无大小限制
  - `main.go:249` `websocket.New(...)` 未传 `Config{Origins:...}`
- **修复说明**：`websocket.New(handler, websocket.Config{Origins: allowedOrigins})` 自动校验 Origin；handler 内 `SetReadLimit(4096)` + `SetReadDeadline(60s)` + `SetPongHandler` 续期；WSManager 新增 `maxConnections` 字段，`Register` 检查容量上限（默认 100）；config 新增 `ws_max_connections`

### I-6 文件写入非原子，崩溃产生半成品 ✅已修复

- **状态**：确认成立 → 已修复
- **修复说明**：在 `files.go` 新增公共 `AtomicWrite(path, data, perm)`：写入同目录临时文件后 `os.Rename` 原子替换，失败时清理临时文件、目标文件保持原状。`SaveNoteWithCheck`（notes.go）与 `saveTokens`（share.go）统一改用 `AtomicWrite`，杜绝崩溃/断电产生半成品文件
- **证据**：
  - `go/internal/services/notes.go:281` `os.WriteFile(fullPath, content, 0644)`
  - `go/internal/services/share.go:62-76` token 文件同样非原子写
  - `share.go:53-57` 损坏时 `loadTokens` 静默返回空 map → 所有 share 链接默默失效
- **修复建议**：写临时文件（`full+".tmp"`） + `os.Rename` 原子替换。

### I-7 `SaveUploadedImage` 路径校验与 `ValidatePathSecurity` 不一致 ✅已修复

- **状态**：确认成立 → 已修复
- **修复说明**：新增公共 `IsPathInside(absTarget, absParent string)` 并将 `ValidatePathSecurityAbs` 改为内部委托它；`notes.go` `SaveUploadedImage` 与 `handlers/media.go` 上传路径校验统一改用 `ValidatePathSecurityAbs`；`export.go` `readLibFile` 改用 `IsPathInside`，消除三处各自的 `strings.HasPrefix(absPath, absNotesDir)` 缺分隔符隐患
- **证据**：
  - `go/internal/services/notes.go:494` `if !strings.HasPrefix(absPath, absNotesDir)` 缺 `+Separator`
  - 对比 `go/internal/services/files.go:23-29` `HasPrefix(absTarget, absNotesDir+sep)` 是标准
- **影响**：理论上 `/app/data-backup/x` 误匹配为 `/app/data` 子路径；`export.go:72` 同缺陷。
- **修复建议**：统一使用 `ValidatePathSecurityAbs`；`SanitizeFilename` 应把 `/` 也替换。

### I-8 共享 token 无 TTL、非原子写、损坏静默清空 ✅已修复

- **状态**：确认成立 → 已修复
- **修复说明**：`models.ShareToken` 增 `ExpiresAt string \`json:"expires_at,omitempty"\``（RFC3339，空 = 永久）；新增 `CreateShareTokenWithTTL(notePath, theme string, ttl time.Duration)`，`CreateShareToken` 保留为永久版本向后兼容；`loadTokens` 解析时即时过滤已过期项并惰性清理（下次 mutation 写盘时丢弃）；`saveTokens` 改用 I-6 的 `AtomicWrite`；损坏的 tokens 文件被重命名为 `.share-tokens.json.broken.<unix>` 备份并 `log.Printf` 告警后回退空 map（服务可用），取代静默清空
- **证据**：`go/internal/services/share.go:62-76`（非原子写）+ L53-57 静默空 map + `models.ShareToken` 无 `expires_at`
- **修复建议**：加 `expires_at` 字段，加载时过滤过期；原子写；损坏时备份并告警而非静默清空。

### I-9 `StopBackgroundScanner` 不等待 goroutine 退出 ✅已修复

- **状态**：确认成立 → 已修复
- **修复说明**：NoteService 新增 `scannerDone chan struct{}`，goroutine 退出时 `defer close`；`StopBackgroundScanner` 在 `close(s.stopScanner)` 后以 5s 超时等待 `<-s.scannerDone`，与 `cache.go` `StopCleanup` 的 `cleanupDone` 模式一致；新增 `IsScannerReady()` 供 `/readyz` 使用
- **证据**：
  - `go/internal/services/notes.go:643-651` 仅 `close(s.stopScanner)`，无 done channel
  - 对比 `cache.go:255-279` `StopCleanup` 有 `cleanupDone` + 5s 超时等待

### I-10 SearchIndex 写锁内做磁盘 IO，锁持有时间过长 ✅已修复

- **状态**：确认成立 → 已修复
- **证据**：`go/internal/services/search_index.go:294-301`
  ```go
  si.mu.Lock()
  si.removeNoteFromIndex(notePath)
  return si.indexNoteFresh(notePath)  // 含 os.ReadFile + tokenize
  ```
- **影响**：每次 SaveNote 更新索引持写锁数毫秒~秒级，期间所有 Search RLock 排队。
- **修复说明**：UpdateIndex 改为两阶段 — 锁外 readFile + tokenize，仅锁内做 removeNoteFromIndex（利用反向索引 noteTerms，O(terms) 而非 O(all_terms)）+ indexNoteFresh（写入预解析数据）；`indexNoteFresh` 改为接收预解析参数不再持锁 IO；删除死代码 `indexNote`（已被 UpdateIndex 替代）；全量测试通过

### I-11 多个配置项为死代码，yaml 键被静默丢弃

- **状态**：确认成立
- **证据**：
  - `config.go` 中 `StorageConfig{NotesDir}`、`LogConfig{Enabled}`、`SearchConfig{Enabled}` 字段不全
  - `config.yaml:45/48/51/16/59` 声明的 `cache_dir/temp_dir/backup_dir/log_dir/index_cache_dir` 无对应字段，Unmarshal 静默丢弃
  - `cfg.Search.Enabled` 在 `main.go` 零引用，搜索路由无条件注册
- **修复建议**：删除未用键或落地实现；启动时扫描 yaml key 与结构体不符即告警。

### I-12 CI 仅跑 3 个 E2E spec（实际覆盖率约 7%）

- **状态**：部分成立（分子精确成立，分母实际为 42~43 而非初报的 23）
- **证据**：
  - `.github/workflows/e2e-test.yml:77-79` `npx playwright test ... auth/ notes/crud.spec.ts search/search.spec.ts`
  - `.github/workflows/core-e2e-test.yml:54-55` 同样 3 个 spec
  - `tests/e2e/` 全量扫描共 42~43 个 `*.spec.ts`
- **影响**：CI 真实覆盖率约 7%，远低于 CHANGELOG 声明的"扩展 E2E 覆盖所有 major feature"。
- **修复建议**：CI 改为 `npx playwright test`（全量，按目录分片并行）；据实修正 CHANGELOG。

### I-13 `core-e2e-test.yml` 环境变量时机 bug

- **状态**：确认成立
- **证据**：`.github/workflows/core-e2e-test.yml`
  - L41-44 「Start server」步骤无 env
  - L50-55 「Run Core Tests」步骤才设 `AUTHENTICATION_ENABLED=true`
  - server 已用默认 `enabled:false` 启动，测试步骤的 env 对其零影响
  - 对比 `e2e-test.yml:45-75` 在同一 shell 块内先 export 再启 server（正确）
- **修复建议**：将 env 合并到 Start server 步骤。

### I-14 API 响应字段名与 `API.md` 多处不一致

- **状态**：确认成立
- **证据**：
  - `go/internal/models/types.go:6-14` `Note` 仅有 `modified` 字段
  - `project-docs/developer-guide/API.md:53-66` 示例含 `created_at`、`updated_at`、`title`
  - 列表无 `size/type/tags`，文档未列；搜索 API 用 `limit` 文档写 `per_page`
- **修复建议**：统一字段名（实现真实用 `modified`/`limit`，修文档）；补全文档列全字段；分页响应结构固定。

### I-15 前端无未保存离开提醒，自动保存无草稿恢复 ✅已修复（commit 82a22d5）

- **状态**：确认成立 → 已修复
- **修复说明**：注册 `beforeunload` + dirty 状态追踪；草稿系统从单槽改为多槽 localStorage（`gonote_draft:<path>` key，7 天 TTL 自动清理，配额满时删最旧 10 条重试）；2s 定时落草稿；启动时弹多笔记恢复列表；恢复草稿时检测服务端是否有更新版本，触发二级决策 modal；409 冲突 banner（加载服务器版本 / 保留我的版本，30s 自动收起）；DeleteNote/MoveNote 成功后清理对应草稿；补齐 en-US/zh-CN 11 个新 i18n key
- **证据**：`grep beforeunload shared/frontend` 零命中；无 dirty 状态追踪；无 Service Worker sync/IndexedDB 写回放
- **修复建议**：注册 `beforeunload` + dirty 追踪；localStorage 草稿 + 重入恢复提示。

### I-16 前端大列表无虚拟滚动，全量渲染 DOM

- **状态**：确认成立（前一轮分析）
- **修复建议**：笔记列表虚拟滚动或分页加载；Markdown 渲染加 debounce。

---

## 三、【建议】共 12 项

| # | 问题 | 证据 | 建议 |
|---|------|------|------|
| W-1 | 大量 `regexp.MustCompile` 在热路径每次重新编译 | `go/internal/services/statistics.go:84-168`, `backlink.go` 多处 | 预编译为包级 var |
| W-2 | 多个配置默认值在 `applyDefaults` 中缺失 | `config.go:130-151` 缺 Cache 默认 | applyDefaults 完整化，与文档对齐 |
| W-3 | 前端 Markdown 渲染无 debounce/throttle | `app.js` auto save 路径 | 输入节流后渲染 |
| W-4 | 前端搜索弹层焦点管理不当，TAB 会跳出至背景 | `index.html` 搜索组件 | 管理焦点循环 |
| W-5 | 前端多处硬编码字符串未走 `__()` | locales 对比 | 统一 i18n |
| W-6 | 编辑器快捷键在 IME 激活/macOS 未兼容 | `app.js` 快捷键 | compositionstart/end 检测 |
| W-7 | `performScan` 每 30s 两次完整 WalkDir | `notes.go:669-686` scanAndUpdate(true)+scanAndUpdate(false) | 一次扫描后派生 |
| W-8 | `GetNoteContent` 完全无缓存 | `notes.go:245-259` 每请求 ReadFile | 加 content cache（mtime 校验） |
| W-9 | 无 pprof/metrics/结构化日志 | 无 pprof 路由 | debug 模式暴露 `/debug/pprof/*`；用 slog |
| W-10 | docker compose 无资源限制/只读 fs | `docker/compose/production.yml` | 加 `mem_limit`/`read_only`/tmpfs |
| W-11 | 主题文档数量滞后（声称 8 实有 16） | `FEATURES.md:85` vs `shared/themes/` | 更新文档 |
| W-12 | i18n 文档声称 4 语言实仅 en/zh | `FEATURES.md:337` | 修正或实现 |

---

## 四、修复优先级路线图

### 第 1 批 · 止血（安全+稳定，1-2 周）

1. **S-1** 默认启用认证 + 自检默认密码/secret 拒绝启动
2. **S-2** 限流与全局开关解耦，登录/上传/保存端点强制启用 ✅已修复
3. **S-3** 共享页 XSS（title EscapeString + DOMPurify） ✅已修复
4. **S-5** 注册 `recover.New()` 中间件 ✅已修复
5. **S-4** scanner panic 加 `defer close(s.ready)` ✅已修复
6. **S-9** 容器改非 root 用户 ✅已修复
7. **S-11** 前端 DOMPurify 清洗 marked 输出
8. **I-6** 文件写入原子化 ✅已修复
9. **I-7** `SaveUploadedImage` 路径校验统一 ✅已修复
10. **I-13** core-e2e env 时机修复

### 第 2 批 · 核心完善（3-6 周）

11. **S-7** 搜索倒排改造（前缀 tree / sort.Search 二分）✅已修复
12. **S-6** 后台扫描同步刷新 SearchIndex（注入引用）✅已修复
13. **S-8** MoveNote 改为"先 rename 再 update backlinks"+补偿回滚 ✅已修复
14. **S-10** /healthz + /readyz 拆分 ✅已修复
15. **I-1** 并发写入加 mutex + mtime 乐观锁 ✅已修复
16. **I-2** Cache Get RLock 优化 ✅已修复
17. **I-9** scanner 引入 WaitGroup + done channel ✅已修复
18. **I-10** SearchIndex 锁外读盘 + 反向索引 ✅已修复
19. **I-4** IP 限流键支持 X-Forwarded-For ✅已修复
20. **I-5** WebSocket 读超时 + Origin + ReadLimit ✅已修复

### 第 3 批 · 体验质量（6-12 周）

21. **S-12** 回收站/软删除实现
22. **I-3** 合并双 tag 缓存 ✅已修复
23. **I-11** 清理死配置或落地实现
24. **I-12** CI E2E 全量 + 修正 CHANGELOG
25. **I-14** 统一 API 文档与实现
26. **I-15** 前端 beforeunload + 草稿 localStorage ✅已修复
27. **I-16** 前端虚拟滚动
28. 其余 W-1~W-12 建议项

> 完整证据链（含代码片段）复核报告见会话记录。修复时建议遵循：每条 issue 一次 PR + 附 regression 测试，避免大改次生风险。
