// Package eval 提供 LLM-as-Judge 评测能力。
// 用独立 LLM 对圆桌讨论质量进行多维度评分，替代关键词匹配的粗粒度评分器。
package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ============================================================
// LLM Judge 评分维度（每项 0-20，总分 100）
// ============================================================

const judgeSystemPrompt = `你是一位资深的圆桌讨论质量评审专家。你将收到一场 AI 模拟名人辩论的完整记录，请从以下 5 个维度评分（每项 0-20）：

1. **角色逼真度** (0-20)：发言是否像本人的语气、立场、思维方式和标志性用语？是否有违背角色设定的内容？
2. **论证质量** (0-20)：论据是否扎实？逻辑链是否完整？是否有具体的例子或推理支撑？
3. **互动性** (0-20)：是否回应/反驳了对方观点？还是各说各话？有无真正交锋？
4. **表达自然度** (0-20)：是否口语化、有情感变化？还是充满 AI 模板味和列表化输出？
5. **洞察力** (0-20)：是否提出了非显而易见的视角？有没有让你觉得"这个角度有意思"？

评分标准：
- 0-4: 极差，完全不符合
- 5-9: 较差，有明显缺陷
- 10-13: 一般，基本合格但不出彩
- 14-16: 良好，有亮眼之处
- 17-20: 优秀，接近真人水平

你必须严格按照 JSON 格式输出，不要有任何额外文字：
{
  "dimensions": [
    {"name": "角色逼真度", "score": <int>, "reason": "<一句话理由>"},
    {"name": "论证质量", "score": <int>, "reason": "<一句话理由>"},
    {"name": "互动性",   "score": <int>, "reason": "<一句话理由>"},
    {"name": "表达自然度", "score": <int>, "reason": "<一句话理由>"},
    {"name": "洞察力",   "score": <int>, "reason": "<一句话理由>"}
  ],
  "overall_comment": "<一段话总结这场辩论的亮点和不足>"
}`

// ============================================================
// LLM Judge 核心结构
// ============================================================

// LLMJudge 用外部 LLM 对圆桌讨论进行多维度评分。
type LLMJudge struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewLLMJudge 创建 LLM 评测器。
func NewLLMJudge(apiKey string) *LLMJudge {
	return &LLMJudge{
		apiKey:  apiKey,
		baseURL: "https://api.deepseek.com",
		model:   "deepseek-v4-pro",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// JudgeResult LLM 评测结果。
type JudgeResult struct {
	Dimensions     []JudgeDimension `json:"dimensions"`
	OverallComment string           `json:"overall_comment"`
	Total          float64          `json:"-"`
}

// JudgeDimension 单个维度的评分。
type JudgeDimension struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// Evaluate 对一场圆桌讨论进行 LLM 评测。
// topic: 讨论话题
// messages: 格式化的发言记录（含发言人、轮次、内容）
// personaContext: 角色简介（帮助 Judge 理解应该期待什么样的表现）
func (j *LLMJudge) Evaluate(ctx context.Context, topic string, messages string, personaContext string) (*JudgeResult, error) {
	userPrompt := fmt.Sprintf(`## 讨论话题
%s

## 参与角色简介
%s

## 辩论记录
%s

请评分。`, topic, personaContext, messages)

	body := map[string]interface{}{
		"model": j.model,
		"messages": []map[string]string{
			{"role": "system", "content": judgeSystemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  1024,
		"temperature": 0.3,
		"stream":      false,
		"thinking":    map[string]string{"type": "disabled"},
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", j.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("构造 Judge 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Judge API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Judge API 返回 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("解析 Judge 响应失败: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("Judge 返回空响应")
	}

	content := apiResp.Choices[0].Message.Content
	// 提取 JSON（LLM 可能在 JSON 前后加了 markdown 代码块标记）
	content = extractJSON(content)

	var result JudgeResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析 Judge JSON 失败: %w\n原始内容: %s", err, truncateStr(content, 500))
	}

	// 计算总分
	for _, d := range result.Dimensions {
		result.Total += d.Score
	}

	return &result, nil
}

// PersonaCards 返回评测用的角色简介。
func PersonaCards() string {
	return `- Steve Jobs: Apple 联合创始人，追求极致简洁与用户体验，厌恶平庸和委员会决策。说话短促锐利，常用强判断句。时间截止 2011 年。
- Elon Musk: Tesla/SpaceX CEO，第一性原理思考者，强调物理约束和工程速度。说话直接，喜欢用数据和物理类比。`
}

// FormatMessages 将消息列表格式化为 Judge 可读的文本。
func FormatMessages(messages []Message) string {
	var buf strings.Builder
	for i, m := range messages {
		personaName := m.PersonaID
		switch m.PersonaID {
		case "steve-jobs":
			personaName = "Steve Jobs"
		case "elon-musk":
			personaName = "Elon Musk"
		case "naval-ravikant":
			personaName = "Naval Ravikant"
		case "zhang-xiaolong":
			personaName = "张小龙"
		case "zhang-yiming":
			personaName = "张一鸣"
		}
		fmt.Fprintf(&buf, "[第%d轮] %s：%s\n\n", m.Round, personaName, m.Content)
		_ = i
	}
	return buf.String()
}

// extractJSON 从 LLM 输出中提取 JSON 内容（处理 markdown 代码块包裹）。
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// 尝试去掉 ```json ... ``` 包裹
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
