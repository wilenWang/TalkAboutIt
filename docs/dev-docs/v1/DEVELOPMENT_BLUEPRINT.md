# TalkAboutIt 开发总纲

> 版本：v1 Blueprint  
> 日期：2026-05-12  
> 适用范围：TalkAboutIt 后续全部开发、评审、联调、验收  
> 决策优先级：`codex-design.md` > `technical-spec.md` / `implementation-plan.md` > 原型稿

---

## 1. 项目概述

**一句话定义：** TalkAboutIt v1 是一个“名人视角问题碰撞机”，用户给出一个问题，选择 2 到 4 个 Persona，让系统自动生成一段有观点冲突和信息增量的多轮圆桌讨论。

**目标用户：**
- 经常需要多视角快速碰撞的中文 AI 重度用户
- AI 创业者
- 产品经理
- 内容策划者 / 知识工作者

**核心场景：**
- 用户输入一个讨论题
- 选择 2 到 4 个预置 Persona
- 系统自动跑完 2 到 3 轮讨论
- 用户实时观看流式输出
- 用户在讨论结束后查看完整记录与回放

**MVP 第一原则：**
- 一切以“能跑通的 vertical slice”优先
- 先做预置 Persona + 自动讨论 + 流式播放 + 持久化
- 不做任何会显著拉高复杂度但不影响首个闭环的功能

---

## 2. 最终技术选型

### 2.1 总体决策

| 领域 | 最终方案 | 说明 |
|---|---|---|
| 前后端通信 | `REST + SSE` | `Codex 建议替代 WebSocket`，v1 不保留 WS |
| Persona 数据模型 | `Persona Schema v1` | `Codex 建议替代旧角色卡格式` |
| 产品边界 | 只做预置 Persona 选择 | `Codex 建议砍掉自定义 Persona 编辑器` |
| 前端信息架构 | `Talk 单主界面` | 不保留独立 `Role Tab` 作为 v1 功能 |

### 2.2 后端技术栈

| 组件 | 选型 |
|---|---|
| 语言 | Go 1.22+ |
| Router | `chi` v5 |
| DB | SQLite（`modernc.org/sqlite`） |
| 配置 | `yaml.v3` + 环境变量覆盖 |
| LLM Gateway | 统一 Provider 接口，默认 DeepSeek |
| OpenAI 兼容 Provider | `net/http` |
| Anthropic Provider | `net/http` |
| 流式协议 | HTTP SSE |

### 2.3 前端技术栈

| 组件 | 选型 |
|---|---|
| 框架 | React 18 + TypeScript |
| 构建 | Vite |
| 样式 | Tailwind CSS + 少量组件级 CSS Token |
| 流式订阅 | 浏览器原生 `EventSource` |
| 状态管理 | React hooks + reducer，本地状态优先 |

### 2.4 关键替代项

- `WebSocket -> SSE`：最终统一为 SSE。原因是 v1 只有服务端到客户端单向事件流，不需要双向 socket。
- `旧 CharacterCard -> Persona Schema v1`：所有 persona 文件、后端模型、前端展示、评测脚本全部切到同一 Schema。
- `默认 Provider -> DeepSeek`：MVP 默认统一为 DeepSeek，走 OpenAI 兼容接口以降低成本并保留切换灵活性。
- `Role Tab 编辑器 -> 取消`：v1 不做 Persona 新建、编辑、上传。Persona 为仓库内预置资产。
- `主持人模式 / 用户插话 -> 取消`：v1 不做双向控制，不做 `_user` 消息注入。
- `讨论摘要机制 -> 延后到 v2`：因为 v1 固定 2 到 3 轮，不需要为了上下文压缩引入额外摘要链路。

---

## 3. 最终目录结构

```text
TalkAboutIt/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── api/
│   │   │   ├── router.go
│   │   │   ├── persona_handler.go
│   │   │   ├── roundtable_handler.go
│   │   │   └── sse_handler.go
│   │   ├── config/
│   │   ├── engine/
│   │   ├── llm/
│   │   ├── persona/
│   │   └── session/
│   ├── personas/
│   │   ├── steve-jobs.json
│   │   ├── elon-musk.json
│   │   ├── naval-ravikant.json
│   │   ├── zhang-yiming.json
│   │   └── zhang-xiaolong.json
│   ├── data/
│   ├── config.example.yaml
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── api/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── pages/
│   │   ├── types/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── docs/
│   ├── package.json
│   └── vite.config.ts
└── docs/
    ├── technical-spec.md
    ├── implementation-plan.md
    ├── codex-design.md
    ├── DEVELOPMENT_BLUEPRINT.md
    └── talkaboutit-architecture.drawio
```

**目录原则：**
- `backend/`、`frontend/`、`docs/` 严格三层分离
- Persona JSON 只放 `backend/personas/`
- 前端不直接编辑 persona 文件
- 文档层只描述，不承载运行逻辑

---

## 4. 最终 Persona Schema

最终统一采用 `codex-design.md` 提出的 `persona.v1`。

### 4.1 顶层字段

```json
{
  "schema_version": "persona.v1",
  "id": "steve-jobs",
  "name": "Steve Jobs",
  "display_name": "Steve Jobs",
  "avatar": "🍎",
  "role_title": "Apple Co-founder",
  "description": "人物简介与世界观摘要",
  "tags": ["product", "design"],
  "language": {},
  "stance": {},
  "core_beliefs": [],
  "speaking_style": {},
  "knowledge_scope": {},
  "interaction_rules": {},
  "debate_goal": {},
  "prompting": {},
  "examples": {}
}
```

### 4.2 必填字段

- `schema_version`
- `id`
- `name`
- `description`
- `language`
- `stance`
- `core_beliefs`
- `speaking_style`
- `knowledge_scope`
- `interaction_rules`
- `debate_goal`

### 4.3 结构定义

#### `language`
- `primary`: `zh-CN | en-US | mixed`
- `allowed`: 允许输出语言数组
- `default_output`: `follow_user | primary_only`
- `style_hint`: 语言风格提示

#### `stance`
- `default_position`: 默认立场
- `intensity`: 1-5
- `biases`: 倾向列表
- `taboos`: 禁忌观点列表

#### `core_beliefs`
- 数组项：
  - `belief`
  - `priority`（1-5）
  - `rationale`

#### `speaking_style`
- `tone`: 语气标签数组
- `cadence`: `short_punchy | balanced | long_form`
- `verbosity`: 1-5
- `signature_patterns`
- `do`
- `dont`

#### `knowledge_scope`
- `domains`
- `expertise_level`
- `time_cutoff`
- `allowed_inference`: `low | medium | high`
- `unknown_handling`
- `forbidden_claims`

#### `interaction_rules`
- `address_others`
- `disagreement_style`
- `interruption_policy`: `never | rare | allowed | aggressive`
- `question_policy`
- `concession_policy`
- `avoid`

#### `debate_goal`
- `primary_goal`
- `secondary_goals`
- `win_condition`
- `loss_condition`

#### `prompting`
- `system_preamble`
- `reply_constraints`

#### `examples`
- `opening_line`
- `sample_rebuttal`

### 4.4 落地规则

- Persona 文件格式统一为 JSON，不再兼容旧版自由结构角色卡。
- 后端构造 system prompt 时，按模板展开字段，不把整段 JSON 原样塞给模型。
- v1 Persona 资产由仓库维护，至少预置 5 个角色：
  - Steve Jobs
  - Elon Musk
  - Naval Ravikant
  - 张一鸣
  - 张小龙

---

## 5. 最终 API 设计

最终统一采用 `REST + SSE`。

### 5.1 REST API

#### `GET /api/v1/personas`
- 输入：无
- 输出：Persona 列表摘要

#### `GET /api/v1/personas/{id}`
- 输入：Persona ID
- 输出：完整 Persona Schema

#### `POST /api/v1/roundtables`
- 输入：
```json
{
  "topic": "AI 会取代产品经理吗？",
  "personas": ["steve-jobs", "elon-musk", "naval-ravikant"],
  "max_rounds": 3,
  "language": "zh-CN"
}
```
- 输出：创建后的 roundtable 元数据

#### `POST /api/v1/roundtables/{id}/start`
- 输入：无
- 输出：启动结果
- 语义：只允许从 `pending` 进入 `running`

#### `GET /api/v1/roundtables/{id}`
- 输入：Roundtable ID
- 输出：讨论快照
- 用途：
  - 首屏恢复
  - SSE 中断兜底
  - 讨论结束后的详情页回放

#### `GET /api/v1/roundtables/{id}/events`
- 输入：HTTP SSE 订阅
- Header：支持 `Last-Event-ID`
- 输出：8 种事件流

### 5.2 SSE 事件协议

统一 envelope：

```text
id: 43
event: message_chunk
data: {"roundtable_id":"rt_xxx","round":1,"speaker_index":0,"persona_id":"steve-jobs","chunk":"..."}
```

### 5.3 8 种事件类型

#### 1. `stream_start`
- 事件流建立成功
- 前端进入 `streaming` 准备态

#### 2. `round_start`
- 某一轮开始
- 关键字段：`round`, `total_rounds`

#### 3. `speaking`
- 某个 Persona 开始发言
- 关键字段：`round`, `speaker_index`, `persona_id`

#### 4. `message_chunk`
- 流式增量文本
- 只用于临时渲染

#### 5. `message_done`
- 单条发言完成
- 必须带完整 `content`
- 前端与 DB 都以 `message_done` 为最终准

#### 6. `round_end`
- 一轮完成

#### 7. `stream_done`
- 整场讨论完成
- 关键字段：`total_rounds`, `total_messages`, `finished_at`

#### 8. `error`
- 流式错误
- 只区分：
  - `recoverable: true`
  - `recoverable: false`

### 5.4 状态机

#### Roundtable 状态
- `pending -> running -> completed`
- `pending -> failed`
- `running -> failed`

#### v1 明确不做
- `paused`
- `cancelled`
- `waiting_for_user`
- `user_input`

### 5.5 重连与去重

- 服务端为每个 roundtable 维护单调递增 `event_id`
- 客户端断线后使用 `Last-Event-ID` 恢复
- 服务端按 `roundtable_events` 日志补发缺失事件
- 前端按 `event_id` 去重
- `message_chunk` 丢失可接受，最终以 `message_done.content` 收敛一致

---

## 6. 最终 DB Schema

最终只保留 3 张表：`roundtables`、`messages`、`roundtable_events`。

### 6.1 `roundtables`

```sql
CREATE TABLE roundtables (
  id TEXT PRIMARY KEY,
  topic TEXT NOT NULL,
  personas_json TEXT NOT NULL,
  max_rounds INTEGER NOT NULL DEFAULT 3,
  language TEXT NOT NULL DEFAULT 'zh-CN',
  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at DATETIME,
  finished_at DATETIME,
  last_event_id INTEGER NOT NULL DEFAULT 0
);
```

### 6.2 `messages`

```sql
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  roundtable_id TEXT NOT NULL,
  round INTEGER NOT NULL,
  speaker_index INTEGER NOT NULL,
  persona_id TEXT NOT NULL,
  content TEXT NOT NULL,
  event_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id),
  UNIQUE(roundtable_id, round, speaker_index),
  UNIQUE(roundtable_id, event_id)
);
```

### 6.3 `roundtable_events`

```sql
CREATE TABLE roundtable_events (
  roundtable_id TEXT NOT NULL,
  event_id INTEGER NOT NULL,
  event_type TEXT NOT NULL CHECK (
    event_type IN (
      'stream_start',
      'round_start',
      'speaking',
      'message_chunk',
      'message_done',
      'round_end',
      'stream_done',
      'error'
    )
  ),
  round INTEGER,
  speaker_index INTEGER,
  persona_id TEXT,
  message_id TEXT,
  payload_json TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (roundtable_id, event_id),
  FOREIGN KEY (roundtable_id) REFERENCES roundtables(id)
);
```

### 6.4 写入原则

- 先写 DB，再广播 SSE
- 每次事件写入流程：
  1. `BEGIN`
  2. 读取 `last_event_id`
  3. 生成 `next_event_id`
  4. 插入 `roundtable_events`
  5. 更新 `roundtables.last_event_id`
  6. 若是 `message_done`，同步 upsert `messages`
  7. `COMMIT`
  8. 向在线订阅者广播

---

## 7. 实施阶段

原则：先打通一条 vertical slice，再做增量增强。所有 task 都必须有明确输入、输出、完成标准。

### Phase 1：项目骨架与基础资产

#### Task 1.1 目录与工程初始化
- 输入：总纲、现有仓库
- 输出：`backend/`、`frontend/`、`docs/` 三层结构齐备
- 完成标准：后端可 `go build`，前端可 `npm run dev`

#### Task 1.2 配置系统
- 输入：`config.example.yaml` 结构定义
- 输出：配置加载器、环境变量覆盖逻辑
- 完成标准：可读取 server / db / llm / personas 配置

#### Task 1.3 Persona Schema 与 5 个预置角色
- 输入：Persona Schema v1
- 输出：Go 结构体、JSON 校验规则、5 个 persona 文件
- 完成标准：`GET /personas`、`GET /personas/{id}` 正常工作

### Phase 2：Vertical Slice MVP

#### Task 2.1 Roundtable 建表与仓储层
- 输入：3 张表 Schema
- 输出：SQLite 初始化、CRUD、WAL 模式
- 完成标准：可创建 roundtable、写入事件、读取快照

#### Task 2.2 讨论编排器最小版
- 输入：topic、persona 列表、轮数
- 输出：串行 round-robin engine
- 完成标准：能按轮次驱动每个 Persona 依次生成内容

#### Task 2.3 SSE 事件总线
- 输入：engine 生成过程
- 输出：`stream_start -> ... -> stream_done` 全链路事件流
- 完成标准：浏览器可通过 `EventSource` 实时看到讨论推进

#### Task 2.4 Talk 页最小前端
- 输入：Style F Notion Warm 原型
- 输出：单页 MVP UI
- 完成标准：
  - 可选择 Persona
  - 可输入 topic
  - 可选轮数
  - 可点击开始讨论
  - 可实时显示消息流

### Phase 3：接入真实 LLM

#### Task 3.1 LLM Provider 抽象
- 输入：Provider 接口设计
- 输出：统一 `GenerateStream(...)` 能力
- 完成标准：engine 不感知具体模型厂商

#### Task 3.2 OpenAI 兼容 Provider
- 输入：OpenAI / DeepSeek 配置
- 输出：流式调用实现
- 完成标准：可产出 `message_chunk` 和最终完整文本

#### Task 3.3 Anthropic Provider
- 输入：Anthropic 配置
- 输出：Anthropic 流式实现
- 完成标准：与 OpenAI Provider 走同一编排路径

#### Task 3.4 Prompt 组装器
- 输入：Persona Schema v1 + topic + peers + round
- 输出：稳定的 system prompt 模板
- 完成标准：同一个 Persona 的输出风格显著稳定

### Phase 4：恢复能力与错误处理

#### Task 4.1 SSE 重连恢复
- 输入：`Last-Event-ID`
- 输出：历史事件补发 + live stream 衔接
- 完成标准：断线重连后最终消息不丢、不重

#### Task 4.2 快照兜底
- 输入：`GET /roundtables/{id}`
- 输出：完整讨论快照
- 完成标准：即使 SSE 中断，页面刷新后仍能恢复完整结果

#### Task 4.3 错误分类
- 输入：Provider 错误、超时、DB 错误
- 输出：统一 `error` 事件与后端错误码
- 完成标准：
  - 可恢复错误不会直接破坏已落库消息
  - 不可恢复错误会把 roundtable 状态置为 `failed`

### Phase 5：会话回放与前端边界态

#### Task 5.1 会话详情页
- 输入：roundtable 快照 API
- 输出：讨论完成态回放页面
- 完成标准：用户能复看历史讨论

#### Task 5.2 边界态补全
- 输入：Talk 页状态机
- 输出：空态、加载态、错误态、重连态、完成态
- 完成标准：
  - 未选满 2 人时不可开始
  - 断流时显示重连提示
  - 错误时显示明确错误文案

### Phase 6：测试、评测、部署

#### Task 6.1 后端单测
- 输入：config / persona / session / engine / api
- 输出：最小可维护测试集
- 完成标准：核心包单测通过

#### Task 6.2 集成测试
- 输入：真实 API + SQLite + SSE
- 输出：端到端闭环测试脚本
- 完成标准：创建、启动、收流、完成、快照恢复全部通过

#### Task 6.3 固定评测集回归
- 输入：10 个固定题 + 4 维评分
- 输出：`baseline.json`、`candidate.json`
- 完成标准：达到第 9 节门槛

#### Task 6.4 部署打包
- 输入：后端二进制、前端静态资源、persona JSON
- 输出：本地可部署构建物与 Dockerfile
- 完成标准：单机可运行

---

## 8. 前端组件规格

视觉基线采用 `frontend/docs/style-f-notion-warm.html` 的 Notion Warm 风格，但只继承视觉和布局语言，不继承双 Tab 产品结构。

### 8.1 最终信息架构

- v1 只保留 `Talk` 主界面
- 不保留独立 `Role Tab`
- Persona 详情以侧栏预览、抽屉或弹层展示
- 不提供 Persona 编辑、保存、删除、上传

### 8.2 风格规范

- 主色：暖白 + 浅灰背景 + 蓝色操作强调
- 圆角：`4 / 8 / 12px`
- 阴影：轻卡片阴影，不做厚重浮层
- 字体：沿用原型中的现代无衬线，保持克制与清晰
- 内容密度：偏编辑器式，不做营销化装饰

### 8.3 页面结构

#### Header
- Logo：`TalkAboutIt`
- 副标题：`Roundtable Discussion`
- 可选右侧状态区：连接状态 / 当前 roundtable 状态

#### Left Sidebar
- Persona 列表
- 多选态
- 已选数量显示
- Persona 预览入口

#### Main Panel
- 主题输入框
- 轮数选择器
- 开始讨论按钮
- Discussion 流区域
- 当前发言人 / 进行中状态

### 8.4 核心组件

- `PersonaSelector`
- `PersonaListItem`
- `PersonaPreviewCard`
- `TopicInput`
- `RoundSelect`
- `StartDiscussionButton`
- `DiscussionHeader`
- `MessageStream`
- `MessageCard`
- `TypingIndicator`
- `ConnectionBanner`
- `EmptyState`
- `ErrorState`

### 8.5 前端状态

- `idle`
- `creating`
- `starting`
- `streaming`
- `reconnecting`
- `completed`
- `failed`

### 8.6 必做边界态

- Persona 少于 2 人：禁用开始按钮
- Persona 超过 4 人：前端禁止继续选择
- 无 persona 资产：显示空态
- SSE 断开：显示“连接中断，正在恢复”
- `error(recoverable=false)`：进入失败态
- `stream_done`：显示完成态与消息总数

---

## 9. 质量评估机制

最终采用 `codex-design.md` 的固定评测机制。

### 9.1 10 个固定题

每次回归固定跑 10 题，不允许临时改题；中英文题目使用同一套基线。

### 9.2 固定评测配置

- Persona 组合固定：`Steve Jobs + Elon Musk + Naval Ravikant`
- 轮数固定：`3`
- 输出语言：跟随题目语言
- 温度固定：如 `0.7`
- 模型版本固定
- 判分方式：`LLM-as-judge + 人工 spot check`

### 9.3 四维评分

- `人物辨识度`
- `讨论连贯性`
- `信息增量`
- `口吻一致性`

每维 1 到 5 分。

### 9.4 综合分公式

```text
overall = 0.30 * 人物辨识度
        + 0.30 * 讨论连贯性
        + 0.25 * 信息增量
        + 0.15 * 口吻一致性
```

### 9.5 上线门槛

- 10 题平均综合分 `>= 3.8`
- 任一单题综合分 `>= 3.2`
- 任一维度全题平均分 `>= 3.5`
- 相比基线版本，平均综合分下降不得超过 `0.2`

### 9.6 硬失败条件

- Persona 明显串台
- 多角色连续输出几乎同一观点
- 第二轮以后大面积重复第一轮内容
- 中文题大量英文漂移
- 流中断后最终结果缺消息

### 9.7 执行流程

1. 跑 10 个固定题，收集完整 transcript
2. 用同一 judge prompt 做四维评分
3. 生成 `baseline.json` 与 `candidate.json`
4. 做 diff，标记退化项
5. 对低于阈值的样本做人审

---

## 10. v2 功能清单

以下能力明确不进入 v1：

- 用户自定义 Persona 编辑器
- 用户上传角色卡
- 独立 `Role Tab` 编辑工作台
- WebSocket 双向实时控制
- 用户插话 / 打断 / 手动追问
- 主持人 Persona
- 讨论摘要与多版本总结链路
- 联网检索、引用来源、事实校验
- 分享页 / 社区 / 点赞
- 语音合成 / 语音房间
- 图片头像生成
- 多语言 UI 国际化
- 团队工作区
- 付费系统
- A/B 测试平台
- 复杂 Prompt 配置面板

---

## 11. 最终裁决摘要

- **传输层：** 采用 `SSE`，废弃 `WebSocket`
- **Persona：** 全量采用 `Persona Schema v1`
- **产品结构：** v1 不保留独立 `Role Tab`
- **交互范围：** v1 仅支持“预置 Persona + 自动讨论 + 流式观看 + 历史回放”
- **实施顺序：** 先 vertical slice，再接真实 LLM，再补恢复和评测

本文件即 TalkAboutIt v1 的唯一开发总纲。后续若与旧文档冲突，以本文件为准。
