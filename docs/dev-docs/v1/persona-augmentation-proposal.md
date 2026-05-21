# TalkAboutIt 人物复刻系统增广方案

> 作者：Hermes Agent (subagent)  
> 日期：2026-05-14  
> 目标：解决当前 Persona v1 的「死板收敛」问题，在不破坏人物一致性的前提下引入丰富性和不可预测性。

---

## 一、现状诊断

### 1.1 当前架构

```
Persona JSON (v1)
  └─> BuildSystemPrompt() 拼接为静态 system prompt
      └─> LLMGenerate() 发送给 DeepSeek API
          - Temperature: 固定 0.8
          - MaxTokens: 固定 512
          - User message: 第1轮是"请围绕主题发表开场观点"，后续轮是"现在是第N轮，请继续发言"
```

### 1.2 核心问题分析

| 问题 | 根因 | 后果 |
|------|------|------|
| **言论重复** | 每一轮注入的 system prompt 完全一样，user message 对上下文的描述也完全一样 | LLM 的 top sampling 在相同输入下倾向于生成相似输出 |
| **缺乏对话记忆** | `engine.Run()` 中，第 N 轮传给 LLM 的 user message 是固定的字符串模板，**不包含前几轮的实际对话历史** | 人物"失忆"，不知道之前说了什么，也不知道别人说了什么 |
| **静态人物模型** | Persona 是一次性的、不可变的 JSON 快照，没有动态状态 | 人物无法根据对话进程调整策略、情绪、或与其他人的关系 |
| **无多样性机制** | Temperature 0.8 是唯一的不确定性来源，且固定不变 | 没有风格变换、没有即兴发挥，所有发言「同质化」 |
| **缺少对手感知** | peers 只作为名字列表传给 prompt，但不感知对手的发言内容 | 对话变成自说自话，没有真正的针锋相对 |

### 1.3 最致命的问题

**第 2~N 轮的 user message 不包含对话历史。** 这是导致「来回来去就是那几句话」的直接原因——LLM 每次都从零开始构思发言，必然落入相似的思维轨迹和表达模式。

---

## 二、增广方案：三层架构

不是写更长的 JSON 或 prompt，而是在不同层级引入动态性：

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: 静态底座 (Static Base)                            │
│  - 现有的 Persona JSON + BuildSystemPrompt()                │
│  - 核心信念、语言风格、知识边界 —— 保证「一致性」            │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: 动态状态 (Dynamic State)          ← **重点建设**  │
│  - Per-Persona 运行时状态机                                  │
│  - 情感/注意力向量                                          │
│  - 关系图谱（对每个对手的好感/对抗/尊重）                     │
│  - 已用过的论点库 + 去重策略                               │
│  - 对话记忆摘要 + RAG 检索                                 │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: 随机即兴 (Aleatoric Improv)        ← **锦上添花**  │
│  - 温度/参数动态调度                                        │
│  - 偶发风格变体（粗暴日/哲思日/嘲讽日）                       │
│  - Few-shot 语料注入                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、Layer 2 详细设计（核心建设）

### 3.1 对话记忆注入（优先级：🔴 最高，立即可做）

**问题**：当前 `LLMGenerate()` 中 user message 不包含对话历史。

**方案**：在 engine 层维护每个 session 的消息队列，构建 user message 时注入上几轮的对话摘要。

```go
// engine.go 中新增
type ConversationMemory struct {
    Messages []MessageRecord // 最近 N 条发言
    Summary  string          // LLM 生成的摘要（超长对话时用）
}

type MessageRecord struct {
    Round     int
    SpeakerID string
    Content   string
}

// LLMGenerate 修改
func LLMGenerate(provider llm.Provider, language string, memory *ConversationMemory) GenerateFunc {
    return func(ctx context.Context, p persona.Persona, topic string, peers []string, round int) (<-chan llm.ChatChunk, error) {
        systemPrompt := llm.BuildSystemPrompt(p, topic, peers, round, language)

        var userContent string
        if round == 1 {
            userContent = fmt.Sprintf("请围绕主题「%s」发表你的开场观点。", topic)
        } else {
            // 🆕 注入对话历史
            history := memory.FormatForPersona(p.ID, round)
            userContent = fmt.Sprintf(
                "## 前几轮讨论摘要\n%s\n\n---\n现在是第 %d 轮。请针对上面的讨论继续发言，可以直接回应或反驳其他人的观点。",
                history, round,
            )
        }

        req := llm.ChatRequest{
            Messages: []llm.ChatMessage{
                {Role: "system", Content: systemPrompt},
                {Role: "user", Content: userContent},
            },
            MaxTokens:   512,
            Temperature: 0.8,
            Stream:      true,
        }
        return provider.Chat(ctx, req)
    }
}
```

**Go 实现成本**：约 50 行新代码 + engine.go 中的消息记录逻辑。

**预期效果**：人物知道前面说了什么，能接着话题深入，不会重复开场白。这是 ROI 最高的改动。

---

### 3.2 已用论点去重（优先级：🔴 高）

**问题**：即使知道对话历史，LLM 仍可能机械地重复自己的核心论点。

**方案**：维护每个 persona 的「已用论点集合」，在 prompt 中注入"避免重复以下表述"。

```go
// persona/state.go
type PerPersonaState struct {
    UsedArguments    []string          // 已用过的论点（取近 5 条）
    UsedPhrases      map[string]int   // 高频短语计数
    EmotionalTension float64          // 情绪张力 0.0~1.0
}

// prompt.go 的 BuildSystemPrompt 增加参数
func BuildSystemPrompt(p persona.Persona, topic string, peers []string, round int, language string, state *PerPersonaState) string {
    // ... 原有逻辑 ...

    if state != nil && len(state.UsedArguments) > 0 {
        b.WriteString("\n## 注意：避免重复\n")
        b.WriteString("你已经在之前的发言中使用过以下核心论点，请不要逐字重复：\n")
        for _, arg := range state.UsedArguments {
            b.WriteString(fmt.Sprintf("- %s\n", arg))
        }
        b.WriteString("请从新的角度或用新的案例来表达你的立场。\n")
    }
    // ...
}
```

**如何提取已用论点**：每次发言后，用一次轻量 LLM 调用（或简单启发式：取发言的前两句）作为「论点指纹」。

**Go 实现成本**：新增 struct + 在 `engine.Run()` 的发言完成回调中更新 state。

---

### 3.3 情感/注意力状态机（优先级：🟡 中）

**目标**：让人物在对话中「活」起来——愤怒、讽刺、不耐烦、退让等情绪随对话进程变化。

**方案**：维护一个简洁的五维情绪向量，每轮根据对话内容自动更新。

```go
// persona/state.go
type EmotionalState struct {
    Tension    float64 // 紧张度 0~1，被反驳/打断时上升
    Engagement float64 // 投入度 0~1，话题不感兴趣时下降
    Respect    float64 // 对当前讨论的尊重度 0~1
    Creativity float64 // 即兴程度 0~1，高时更愿意讲轶事/类比
    Defensiveness float64 // 防御性 0~1，被连续攻击时上升
}

// 更新规则（可配置的启发式规则，不需要 LLM）
func (es *EmotionalState) Update(event EventType, targetPersona string) {
    switch event {
    case EventDisagreed:
        es.Tension += 0.15
        es.Defensiveness += 0.1
    case EventAgreed:
        es.Tension -= 0.1
        es.Respect += 0.05
    case EventInterrupted:
        es.Tension += 0.2
        es.Defensiveness += 0.15
    case EventTopicShifted:
        es.Engagement = 0.5 // reset
    }
    // clamp & decay
    es.decay(0.95) // 每轮衰减 5% 回到基线
}
```

**注入 prompt**：根据情绪状态在 system prompt 末尾追加一两个简短的「当前心境」描述，例如：

```
当前心境：你开始对 Elon 的反复打断感到不耐烦。你的回复可以比平时更锐利。
```

**Go 实现成本**：约 100 行（状态机 + 事件分类逻辑）。

---

### 3.4 对手感知 prompt 动态调整（优先级：🟡 中）

**问题**：当前 system prompt 只列出了参与者名字，人物不了解对手的特征。

**方案**：在构建每个人的 system prompt 时，注入其他参与者的**简要摘要**（100字以内），让 LLM 知道"我在跟谁说话"。

```go
// 在 BuildSystemPrompt 中增加
func BuildSystemPrompt(p persona.Persona, topic string, peers []persona.Persona, round int, language string, state *PerPersonaState) string {
    // ... 原有逻辑 ...

    // 🆕 对手档案
    b.WriteString("## 你的对手\n")
    for _, peer := range peers {
        b.WriteString(fmt.Sprintf("- **%s**（%s）：%s\n",
            peer.Name,
            peer.RoleTitle,
            peer.Description[:min(100, len(peer.Description))],
        ))
    }
    b.WriteString("\n请在发言时有针对性地回应他们的观点，而不是泛泛而谈。\n")
    // ...
}
```

**Go 实现成本**：需要把 `engine.go` 的 `GenerateFunc` 签名从 `peers []string` 扩展为 `peers []persona.Persona` 或将对手档案预构建。

---

## 四、Layer 3 锦上添花

### 4.1 温度/参数动态调度

不固定 temperature=0.8。根据轮次和情绪状态动态调节：

| 场景 | Temperature | 效果 |
|------|-------------|------|
| 开场 (第1轮) | 0.9 | 鼓励新颖的切入角度 |
| 中途 (2~4轮) | 0.75 | 维持辩论深度 |
| 终场 (最后1轮) | 0.6 | 收敛结论 |
| 情绪高 (Tension > 0.7) | +0.1 | 允许更激烈的语言 |
| 情绪低 (Engagement < 0.3) | +0.15 | 打破无聊循环 |
| 已重复 (UsedArguments > 3) | +0.15 | 强制跳出既定模式 |

### 4.2 动态 Few-shot 注入

在 Persona JSON 中新增 `examples.examples_pool` 字段（可选），存放 5~10 条不同风格的发言示例。每轮从 pool 中随机抽取 1~2 条注入 system prompt 的示例区：

```json
{
  "examples": {
    "opening_line": "...",
    "sample_rebuttal": "...",
    "examples_pool": [
      {"style": "provocative", "text": "..."},
      {"style": "philosophical", "text": "..."},
      {"style": "anecdotal", "text": "..."}
    ]
  }
}
```

**Go 实现成本**：约 30 行（随机选择逻辑 + JSON schema 扩展）。

---

## 五、进阶方向（Phase 2+）

### 5.1 RAG 知识底座：真人语料库（优先级：🟢 低，长期）

为每个人物构建一个语料向量库（用 OpenAI embeddings + 本地向量存储），存放真人的演讲、访谈、推文、文章片段。

每轮发言前，用当前讨论的 topic + 上轮对手发言做相似度检索，召回 2~3 段相关语料作为「知识注入」拼入 system prompt。这样人物的表达会带出真人的措辞习惯和独特案例。

**Go 实现**：
- 使用 [pgvector](https://github.com/pgvector/pgvector) 或 [milvus-lite](https://milvus.io)
- embeddings 来自 DeepSeek 或本地 Ollama/sentence-transformers
- 语料预处理 pipeline：转录 → 分块 → embed → 存库

### 5.2 对抗自博弈（Adversarial Self-Play）

离线阶段：让同一人物的两个实例（不同温度/不同情绪配置）互相辩论同一主题。记录胜者（由 judge LLM 评分）的发言策略。将这些策略作为 few-shot 示例回灌到人物配置中。

### 5.3 多维度人格向量

将 Persona JSON 的各维度映射为一个数值向量（Big Five + 特定领域维度），在每轮辩论中让 LLM 先输出一个「内部思考块」（thinking block），标出当前状态在这些维度上的坐标，然后基于该坐标生成发言。这样可以量化和追溯人物状态变化。

---

## 六、实施优先级排序

| 优先级 | 方案 | 预期效果 | 实现成本 | 建议时间 |
|--------|------|---------|---------|---------|
| 🔴 P0 | 3.1 对话记忆注入 | 不再「失忆」，发言可衔接上下文 | ~50行 Go | 当天 |
| 🔴 P0 | 3.2 已用论点去重 | 避免同一个人重复同样的话 | ~80行 Go | 2天 |
| 🟡 P1 | 3.4 对手感知 prompt | 发言更具针对性，针锋相对 | ~30行 Go | 1天 |
| 🟡 P1 | 4.1 温度动态调度 | 在一致性和多样性间自动平衡 | ~20行 Go | 1天 |
| 🟡 P1 | 3.3 情感状态机 | 人物有「脾气」，对话有张力 | ~100行 Go | 3天 |
| 🟢 P2 | 4.2 动态 Few-shot | 每次发言风格微变但不违和 | ~40行 Go | 2天 |
| 🟢 P2 | 5.1 RAG 真人语料 | 人物措辞更接近真人 | ~300行 Go + infra | 2周 |
| 🟢 P3 | 5.2 对抗自博弈 | 发现更好的辩论策略 | 离线脚本 | 1周 |
| 🟢 P3 | 5.3 人格向量 | 可量化的人物状态分析 | 较大工程 | 2周 |

---

## 七、一个最小可行增广（MVP）的代码改动路径

如果你只想先做一个最小改动就看到效果，以下是具体步骤：

### Step 1: engine.go — 收集对话消息

```go
// 在 Run() 函数中，增加一个 slice
var conversation []MessageRecord

// 每次 message_done 后追加
conversation = append(conversation, MessageRecord{
    Round:     round,
    SpeakerID: p.ID,
    SpeakerName: p.Name,
    Content:   content,
})
```

### Step 2: engine.go — 修改 GenerateFunc 签名

```go
// 旧的
type GenerateFunc func(ctx context.Context, p persona.Persona, topic string, peers []string, round int) (...)

// 新的：增加 conversation history
type GenerateFunc func(ctx context.Context, p persona.Persona, topic string, peers []string, round int, history []MessageRecord) (...)
```

### Step 3: prompt.go — 注入对话历史

```go
// 在 BuildSystemPrompt 末尾（发言规则之后）增加
if len(history) > 0 {
    b.WriteString("\n## 此前的讨论\n")
    for _, msg := range history {
        b.WriteString(fmt.Sprintf("- **[%s]** %s\n", msg.SpeakerName, truncate(msg.Content, 150)))
    }
}
```

这三步改动不超过 **80 行代码**，但能让每个人物在每轮发言中看到之前的全部对话历史，从根本上解决「来回来去就是那几句话」的问题。

之后再逐步加入论点去重、情绪状态机、温度调度等增强。

---

## 八、总结

当前方案的根本问题**不是 JSON 不够长、prompt 不够细**，而是：

1. **对话上下文没有注入** — 每轮 LLM 都在「第一轮」的心态下重新思考
2. **没有跨轮次状态** — 人物不会生气、不会疲惫、不会换策略
3. **没有任何去重机制** — 相同输入 + 0.8 temperature = 相似输出

增广的核心思路是：**让生成过程「有记忆、有状态、有变化」**，而不是让描述更详细。P0 的两个改动（对话记忆 + 论点去重）就能带来质的飞跃。后续的情感状态机和温度调度可以在不增加 JSON 复杂度的情况下，通过运行时的轻量状态自动产生丰富的变化。
