# TalkAboutIt 语言模式设计

## 问题

当前辩论中英混杂：Steve Jobs/Elon Musk 说英文，张小龙/张一鸣说中文。需要支持统一的语言模式。

## 设计

### 前端：语言模式切换器

在 TopicInput 和 RoundSelect 之间添加 LanguageToggle 组件：

```
┌──────────────────────────────────────┐
│ 讨论话题: [____________]             │
│ ┌──────────┐ ┌──────────┐ 轮次: [3] │
│ │ 🇨🇳 中文  │ │ 🇺🇸 English│          │
│ └──────────┘ └──────────┘           │
│              [✦ 开始讨论]            │
└──────────────────────────────────────┘
```

- 两个互斥按钮，选中态蓝色高亮 `bg-[#0075de] text-white`
- 未选中态灰色 `bg-gray-100 text-gray-500`
- 默认选中「中文」
- 状态提升到 App.tsx，通过 `language` prop 向下传递

### 组件文件

新建 `frontend/src/components/LanguageToggle.tsx`:

```tsx
interface Props {
  value: 'zh-CN' | 'en-US';
  onChange: (value: 'zh-CN' | 'en-US') => void;
}
```

### App.tsx 改动

1. 新增 state: `const [language, setLanguage] = useState<'zh-CN' | 'en-US'>('zh-CN')`
2. 在 Controls 区域插入 `<LanguageToggle value={language} onChange={setLanguage} />`
3. `handleStart` 中将 `language` 传给 API: `language: language`

### API 层

`CreateRoundtableRequest.Language` 字段已存在，无需改 API。

### 后端：语言强制 prompt

**文件**: `backend/internal/llm/prompt.go`

`BuildSystemPrompt` 新增 `language` 参数：

```go
func BuildSystemPrompt(p persona.Persona, topic string, peers []string, round int, language string) string {
```

在「发言规则」之后追加语言强制指令：

```go
// 语言模式强制
if language == "en-US" {
    b.WriteString("7. 你必须用英文发言。无论你的 persona 原本使用什么语言，本轮讨论统一使用英文。\n")
} else {
    b.WriteString("7. 你必须用中文发言。无论你的 persona 原本使用什么语言，本轮讨论统一使用中文（简体）。\n")
}
```

### Engine 改动

**文件**: `backend/internal/engine/engine.go`

`LLMGenerate` 需要接收 language:

```go
func LLMGenerate(provider llm.Provider, language string) GenerateFunc {
```

在内部调用 `BuildSystemPrompt` 时传入 language。

`Run` 方法中，创建 `LLMGenerate` 时从 `rt.Language` 读取:

```go
generate := LLMGenerate(provider, rt.Language)
```

```go
eng := engine.NewEngine(store, loader, generate)
```

### 测试文件更新

- `backend/internal/llm/prompt_test.go`: 添加语言强制测试
- `backend/internal/llm/llm_test.go`: 更新 BuildSystemPrompt 调用
- `backend/internal/engine/engine_test.go`: 更新 LLMGenerate 调用
- `backend/test/integration_test.go`: 更新 NewEngine 调用
- `backend/test/realllm_test.go`: 更新 LLMGenerate 调用
- `backend/test/eval` 下所有文件: 更新 LLMGenerate / NewEngine 调用

## 实施步骤

1. 新建 `frontend/src/components/LanguageToggle.tsx`
2. 修改 `frontend/src/App.tsx`：加 state + 插入组件 + 传 language 到 API
3. 修改 `backend/internal/llm/prompt.go`：加 language 参数 + 语言强制
4. 修改 `backend/internal/engine/engine.go`：LLMGenerate 接收 language
5. 修改 `backend/internal/api/sse_handler.go`：StartRoundtable 传 language 给 engine
6. 更新所有调用方（测试文件、eval、realllm 测试）
7. 运行 `go test ./...` 确保无编译错误

## 验证标准

- 选「中文」模式，Jobs 和 Musk 都用中文发言
- 选「English」模式，张小龙和张一鸣都用英文发言
- 语言切换器 UI 交互流畅
- 所有已有测试通过
