# Phase 2 Vertical Slice 审查报告

## 结论

Phase 2 的 happy path 已经打通：后端测试可通过，前端可构建，基本链路 `create -> start -> stream -> snapshot` 能跑。

但按 Blueprint 要求审查，这个实现还不能算“完整且稳健”。当前存在 4 个会直接影响协议正确性或稳定性的严重问题，另外有若干状态管理、输入校验和测试覆盖缺口。

## 验证结果

- `go test ./...`：通过
- `npm run build`：通过
- `npm test`：失败，`frontend/package.json` 没有 `test` script

---

## 🔴 严重问题

### 1. `round_end` 被写入 DB，但没有广播到 SSE，事件序列不完整

**位置**

- `backend/internal/engine/engine.go:180-186`
- Blueprint 要求：`docs/DEVELOPMENT_BLUEPRINT.md:479-482`
- 事件协议：`docs/DEVELOPMENT_BLUEPRINT.md:327-332`

**问题**

`round_end` 事件只调用了 `store.AddEvent(...)`，但没有像其他事件一样 `broadcast`。这会导致：

- 数据库里有 `round_end`
- 实时 SSE 客户端却收不到 `round_end`
- 实际事件顺序变成 `... -> message_done -> next round_start / stream_done`

这与审查要求中的事件序列 `stream_start -> round_start -> speaking -> message_chunk -> message_done -> round_end -> stream_done` 不一致。

**修复建议**

```go
evt, err := e.store.AddEvent(ctx, tableID, "round_end", &r, nil, nil, nil, map[string]interface{}{
    "round":        round,
    "total_rounds": rt.MaxRounds,
})
if err != nil {
    return fmt.Errorf("发送 round_end 失败: %w", err)
}
e.broadcast(*evt)
sleepBetweenEvents()
```

---

### 2. SSE 重连补发存在竞态窗口，会丢事件

**位置**

- `backend/internal/api/sse_handler.go:114-127`
- Blueprint 重连要求：`docs/DEVELOPMENT_BLUEPRINT.md:353-359`

**问题**

当前顺序是：

1. 先 `GetEventsAfter(lastEventID)` 补发历史
2. 再 `Subscribe(id)` 订阅 live 事件

如果有新事件恰好发生在这两步之间，那么它既不在“历史补发”里，也不在“实时订阅”里，客户端会永久漏掉该事件。

这直接破坏了 `Last-Event-ID` 的恢复语义。

**修复建议**

至少要保证“补历史”和“接 live”之间无缝衔接。常见做法：

1. 先订阅 channel
2. 再读取当前 `last_event_id` 或补历史
3. 将历史和 live 合并，并按 `event_id` 去重

示例方向：

```go
ch, cancel := h.bus.Subscribe(id)
defer cancel()

history, err := h.store.GetEventsAfter(r.Context(), id, lastEventID)
if err != nil {
    http.Error(w, `{"error":"历史事件读取失败"}`, http.StatusInternalServerError)
    return
}

seen := make(map[int]struct{})
for _, evt := range history {
    seen[evt.EventID] = struct{}{}
    writeEvent(w, evt)
}
flusher.Flush()

for {
    select {
    case evt := <-ch:
        if _, ok := seen[evt.EventID]; ok {
            continue
        }
        writeEvent(w, evt)
        flusher.Flush()
```

更稳妥的方案是把“订阅 + 从某个 event_id 开始 drain backlog”做成总线层的一个原子操作。

---

### 3. `EventBus` 的取消订阅和广播之间有关闭 channel 的竞态，可能直接 panic

**位置**

- `backend/internal/api/sse_handler.go:41-50`
- `backend/internal/api/sse_handler.go:55-76`

**问题**

`Publish` 会先复制 `chans` 列表，之后在锁外发送。

`cancel()` 会：

1. 从 map 删除 `ch`
2. 解锁
3. `close(ch)`

如果时序是：

1. `Publish` 已经复制了某个 `ch`
2. `cancel()` 把这个 `ch` 关闭
3. `Publish` 再向这个 `ch` 发送

那么会触发 `send on closed channel` panic，直接打崩请求处理 goroutine，严重时可影响整个进程稳定性。

**修复建议**

不要关闭仍可能被并发发送的数据通道。更安全的方式：

- 只从订阅表中删除，不 `close(ch)`
- 或引入 subscriber 结构体，带独立 `done` 信号
- 或发送端统一持锁并保证不会对已关闭通道发送

一个简单改法：

```go
type subscriber struct {
    ch chan session.Event
}

func (b *EventBus) Subscribe(roundtableID string) (chan session.Event, func()) {
    ch := make(chan session.Event, 64)
    b.mu.Lock()
    if b.subscribers[roundtableID] == nil {
        b.subscribers[roundtableID] = make(map[chan session.Event]struct{})
    }
    b.subscribers[roundtableID][ch] = struct{}{}
    b.mu.Unlock()

    cancel := func() {
        b.mu.Lock()
        defer b.mu.Unlock()
        if subs, ok := b.subscribers[roundtableID]; ok {
            delete(subs, ch)
            if len(subs) == 0 {
                delete(b.subscribers, roundtableID)
            }
        }
        // 不主动 close(ch)
    }
    return ch, cancel
}
```

如果不 close 数据通道，`SSEHandler` 退出依赖 `r.Context().Done()` 即可。

---

### 4. `StartRoundtable` 不是原子状态切换，重复请求会产生双启动和脏状态

**位置**

- `backend/internal/api/sse_handler.go:273-296`
- `backend/internal/engine/engine.go:86-89`
- Blueprint 状态机：`docs/DEVELOPMENT_BLUEPRINT.md:279`
- Blueprint 状态流转：`docs/DEVELOPMENT_BLUEPRINT.md:342-345`

**问题**

当前流程是：

1. 读取 roundtable
2. 检查 `status == pending`
3. 启一个 goroutine，稍后才在 `Engine.Run()` 里 `UpdateStatus(..., "running")`

这会造成两个问题：

- 两个并发 `POST /start` 很可能都看到 `pending`，然后各自启动一个 goroutine
- 第二个 goroutine 进入 `Engine.Run()` 后发现状态已不是 `pending`，会报错；`StartRoundtable` 的错误处理又会写入 `error` 事件并把状态改成 `failed`

结果是：一个合法讨论明明已经在正常运行，另一个重复启动请求却可能把它最终打成 `failed`。

**修复建议**

把“从 `pending` 切到 `running`”做成单条原子 SQL，只有抢到状态的人才能真正启动引擎。

建议在 store 增加 compare-and-swap：

```go
func (s *Store) MarkRunning(ctx context.Context, id string) (bool, error) {
    res, err := s.db.ExecContext(ctx, `
        UPDATE roundtables
        SET status = 'running', started_at = CURRENT_TIMESTAMP
        WHERE id = ? AND status = 'pending'
    `, id)
    if err != nil {
        return false, err
    }
    n, err := res.RowsAffected()
    if err != nil {
        return false, err
    }
    return n == 1, nil
}
```

然后在 `StartRoundtable` 中：

```go
ok, err := h.store.MarkRunning(r.Context(), id)
if err != nil {
    http.Error(w, `{"error":"启动失败"}`, http.StatusInternalServerError)
    return
}
if !ok {
    http.Error(w, `{"error":"roundtable 不在 pending 状态"}`, http.StatusConflict)
    return
}
```

同时把 `Engine.Run()` 里的“切 running”逻辑移除，避免双写。

---

## 🟡 一般问题

### 1. `CreateRoundtable` 返回的 `created_at` 是零值

**位置**

- `backend/internal/api/sse_handler.go:236-257`

**问题**

`rt.CreatedAt` 在插入前没有赋值，`CreateRoundtable` 之后也没有从 DB reload，因此响应里的 `created_at` 实际上会是 Go 零值时间。

**修复建议**

创建后立即回读，或者在写入前显式设置时间：

```go
rt.CreatedAt = time.Now().UTC()
```

更稳妥是写库后 `GetRoundtable()` 再返回。

---

### 2. 空 `personas_json` 会导致 `Engine.Run()` 直接 panic

**位置**

- `backend/internal/engine/engine.go:72-95`

**问题**

`pid0 := personas[0].ID` 默认假设至少有一个 persona。虽然 `CreateRoundtable` 限制了最少 2 个，但数据库数据可能来自测试、迁移或未来接口，当前实现对空数组没有兜底。

这也是本轮测试缺失的边界场景之一。

**修复建议**

```go
if len(personaIDs) == 0 {
    return fmt.Errorf("roundtable 缺少 persona")
}
```

同时增加测试覆盖空 roundtable。

---

### 3. `StartRoundtable` 使用 `context.Background()`，请求取消和进程关闭都不会传播

**位置**

- `backend/internal/api/sse_handler.go:284-295`
- `backend/internal/engine/engine.go:50-53`

**问题**

后台 goroutine 脱离了请求上下文，而且 `sleepBetweenEvents()` 不检查 `ctx.Done()`。这意味着：

- 客户端断开连接后，后端仍然继续跑完整场讨论
- 服务关闭时，不容易优雅中止正在运行的 roundtable

**修复建议**

- 把 server 级别的生命周期 context 注入 handler/engine
- 在事件间隔和主循环中检查 `ctx.Done()`

示例：

```go
func sleepBetweenEvents(ctx context.Context) error {
    timer := time.NewTimer(50 * time.Millisecond)
    defer timer.Stop()
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-timer.C:
        return nil
    }
}
```

---

### 4. Store 层的状态更新没有校验行数，也没有保护非法流转

**位置**

- `backend/internal/session/store.go:168-183`

**问题**

`UpdateStatus()` 只执行 `UPDATE`，不检查：

- roundtable 是否存在
- 是否真的更新到 1 行
- 状态流转是否符合 `pending -> running -> completed/failed`

这会让调用方以为更新成功，但实际可能什么都没改。

**修复建议**

- 检查 `RowsAffected()`
- 需要时把旧状态带入 `WHERE`
- 为常用流转提供显式方法，例如 `MarkRunning`、`MarkCompleted`、`MarkFailed`

---

### 5. `CreateRoundtable` 的输入校验不完整

**位置**

- `backend/internal/api/sse_handler.go:220-245`
- Blueprint 范围：`docs/DEVELOPMENT_BLUEPRINT.md:488-491`

**问题**

当前只校验了：

- topic 非空
- personas 至少 2 个

但缺少：

- personas 最多 4 个
- persona ID 是否存在
- persona 是否重复
- topic 是否仅由空白组成

这会把明显无效的数据写进 DB，直到 `start` 才失败。

**修复建议**

```go
if strings.TrimSpace(req.Topic) == "" {
    http.Error(w, `{"error":"话题不能为空"}`, http.StatusBadRequest)
    return
}
if len(req.Personas) < 2 || len(req.Personas) > 4 {
    http.Error(w, `{"error":"参与者数量必须在 2 到 4 之间"}`, http.StatusBadRequest)
    return
}
```

然后用 `loader.LoadOne()` 逐个验证 persona 是否存在。

---

### 6. 前端完成一次讨论后无法再开始下一次

**位置**

- `frontend/src/App.tsx:10`
- `frontend/src/App.tsx:112`

**问题**

`canStart` 只允许 `status === 'idle'`。但成功结束后状态会变成 `completed`，页面里没有任何逻辑把它切回 `idle`，因此用户完成一次后就不能再次点击“开始讨论”。

**修复建议**

开始新一轮时允许 `completed`，或在用户修改 topic/personas 时重置为 `idle`。

例如：

```ts
const canStart =
  selectedPersonas.length >= 2 &&
  topic.trim().length > 0 &&
  (status === 'idle' || status === 'completed');
```

---

### 7. 前端没有消费 `message_chunk`，实时流只显示“正在输入”，不显示增量文本

**位置**

- `frontend/src/App.tsx:24-82`
- Blueprint：`docs/DEVELOPMENT_BLUEPRINT.md:318-325`
- Task 2.4 完成标准：`docs/DEVELOPMENT_BLUEPRINT.md:484-492`

**问题**

当前 UI 只在 `message_done` 时追加最终消息，对 `message_chunk` 完全忽略。用户看到的不是“增量消息流”，而只是一个 typing 指示器加最终整段落地。

这不影响 mock happy path，但与事件协议里 `message_chunk` 的临时渲染语义不一致。

**修复建议**

给当前 speaking persona 维护一条 `status: 'streaming'` 的临时消息，在 `message_chunk` 时追加内容，在 `message_done` 时替换为最终定稿。

---

### 8. SSE 历史补发失败会被静默吞掉

**位置**

- `backend/internal/api/sse_handler.go:115-122`

**问题**

`GetEventsAfter()` 出错时，handler 会直接忽略错误并继续进入 live stream。这样客户端会以为重连成功，但实际上历史事件已经丢了。

**修复建议**

补发失败时应直接返回 500，或者至少写一条 `error` 事件并断开，让客户端明确知道恢复失败。

---

## 🟢 建议

### 1. `message_done` 写 `messages` 时建议改为 idempotent upsert

**位置**

- `backend/internal/session/store.go:229-239`
- Blueprint 要求：`docs/DEVELOPMENT_BLUEPRINT.md:440`

**原因**

Blueprint 写的是“同步 upsert `messages`”。当前是 `INSERT`，一旦未来做重试、补写或恢复流程，就会因为唯一约束直接失败。

**建议**

使用 SQLite 的 `INSERT ... ON CONFLICT(id) DO UPDATE`。

---

### 2. CORS 只加在 SSE 路由上，其他 API 没有统一处理

**位置**

- `backend/internal/api/sse_handler.go:94-99`
- `backend/internal/api/router.go:34-40`

**原因**

如果前后端分域部署，`GET /personas`、`POST /roundtables`、`POST /start` 都缺少统一的 CORS/OPTIONS 处理。

**建议**

把 CORS 放到中间件层统一处理，不要只在 SSE handler 里单独设置。

---

### 3. API client 丢失了后端错误上下文

**位置**

- `frontend/src/api/client.ts:3-28`

**原因**

当前前端只抛固定文案，例如“创建讨论失败”。后端返回的具体错误体完全被忽略，不利于定位问题。

**建议**

解析 JSON 错误响应：

```ts
async function parseError(res: Response, fallback: string) {
  try {
    const body = await res.json() as { error?: string };
    return body.error || fallback;
  } catch {
    return fallback;
  }
}
```

---

### 4. 有一些无效或未完成的前端状态痕迹

**位置**

- `frontend/src/App.tsx:22`
- `frontend/src/components/MessageStream.tsx:19-22`

**原因**

- `currentRoundRef` 没有被使用
- `useMemo(() => messages, [messages])` 没有实际价值

**建议**

删掉无效状态和无意义 memo，降低噪音。

---

## 测试覆盖评估

### 已覆盖

- Store 的基础 CRUD、事件写入、NULL 字段读取
- Engine 的 mock happy path
- API 的基础 handler 和 start happy path

### 未覆盖的关键场景

- 并发 `POST /start`，验证不会双启动或错误打成 `failed`
- `EventBus` 的 `Subscribe/Publish/cancel` 并发安全
- SSE `Last-Event-ID` 重连窗口，验证“补历史 + live”不会漏事件
- `round_end` 是否真的被广播到客户端，而不只是写进 DB
- 空 `personas_json`、非法 persona ID、重复 persona、超过 4 个 persona
- `UpdateStatus` 对不存在 roundtable 的行为
- `CreateRoundtableResponse.created_at` 是否为真实时间
- 前端 `useSSE` / `App` 的状态流转、重连、chunk 渲染

### 测试基础设施缺口

- `frontend/package.json:12-16` 没有 `test` script
- 当前仓库没有前端单元测试或组件测试

---

## 对 Blueprint 的完成度判断

### Task 2.1 仓储层

**结论：基本完成，但不够稳健。**

已实现三表、WAL、基本 CRUD 和事件日志。主要缺口是：

- 事件写入缺少并发写保护/幂等策略
- 状态更新不校验行数和流转合法性
- `message_done -> messages` 不是 upsert

### Task 2.2 编排引擎

**结论：部分完成。**

happy path 的 round-robin 已跑通，但还存在：

- `round_end` 未广播
- 空 persona 会 panic
- context 取消不生效

### Task 2.3 SSE + API

**结论：部分完成。**

基本 SSE 链路已通，但有两个结构性问题：

- 补历史再订阅的重连竞态
- EventBus 的关闭通道竞态 panic

此外 `StartRoundtable` 的异步启动不是原子切换。

### Task 2.4 前端

**结论：UI 已有最小可用形态，但“实时流”和状态管理仍不完整。**

- 可选 persona、输入 topic、选轮数、点击开始
- 能看到最终消息和正在输入状态
- 但没有真正消费 `message_chunk`
- 完成后不能再次开始
- 没有前端测试

### 总判断

**Phase 2 已达到“可演示”的 vertical slice，但还没有达到 Blueprint 语义上的“完整实施”。**

如果只看 happy path，可以进入下一阶段；如果要把这套代码作为后续真实 LLM 接入的稳定基座，建议先修掉上面的 4 个严重问题。

