// Package eval 提供 TalkAboutIt 圆桌讨论质量的规则评分器。
// v1 版本基于关键词匹配与启发式规则，不依赖外部 LLM。
package eval

import (
	"strings"
	"unicode/utf8"
)

// DimensionScore 代表单一维度的评分结果。
type DimensionScore struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Max    float64 `json:"max"`
	Reason string  `json:"reason"`
}

// EvaluationResult 代表一次 roundtable 的完整评测结果。
type EvaluationResult struct {
	Topic      string           `json:"topic"`
	Category   string           `json:"category"`
	Dimensions []DimensionScore `json:"dimensions"`
	Total      float64          `json:"total"`
	MaxTotal   float64          `json:"max_total"`
}

// ScoreRoundtable 对一组消息进行 4 维评分。
// 四个维度：内容深度、角色一致性、交互质量、表达自然度。
func ScoreRoundtable(topic string, category string, messages []Message) EvaluationResult {
	if len(messages) == 0 {
		return EvaluationResult{
			Topic:    topic,
			Category: category,
			Dimensions: []DimensionScore{
				{Name: "内容深度", Score: 0, Max: 25, Reason: "无消息"},
				{Name: "角色一致性", Score: 0, Max: 25, Reason: "无消息"},
				{Name: "交互质量", Score: 0, Max: 25, Reason: "无消息"},
				{Name: "表达自然度", Score: 0, Max: 25, Reason: "无消息"},
			},
			Total:    0,
			MaxTotal: 100,
		}
	}

	// 收集所有消息内容
	var contents []string
	for _, m := range messages {
		contents = append(contents, m.Content)
	}
	allText := strings.Join(contents, " ")

	// 1. 内容深度（0-25）
	depthScore := scoreDepth(allText, contents)

	// 2. 角色一致性（0-25）
	consistencyScore := scoreConsistency(contents)

	// 3. 交互质量（0-25）
	interactionScore := scoreInteraction(contents)

	// 4. 表达自然度（0-25）
	naturalnessScore := scoreNaturalness(contents)

	result := EvaluationResult{
		Topic:    topic,
		Category: category,
		Dimensions: []DimensionScore{
			{Name: "内容深度", Score: depthScore, Max: 25, Reason: depthReason(depthScore)},
			{Name: "角色一致性", Score: consistencyScore, Max: 25, Reason: consistencyReason(consistencyScore)},
			{Name: "交互质量", Score: interactionScore, Max: 25, Reason: interactionReason(interactionScore)},
			{Name: "表达自然度", Score: naturalnessScore, Max: 25, Reason: naturalnessReason(naturalnessScore)},
		},
		MaxTotal: 100,
	}
	for _, d := range result.Dimensions {
		result.Total += d.Score
	}
	return result
}

// Message 代表评测用的消息结构。
type Message struct {
	PersonaID string `json:"persona_id"`
	Content   string `json:"content"`
	Round     int    `json:"round"`
}

// scoreDepth 基于关键词密度、论证标志词和文本长度评估内容深度。
func scoreDepth(allText string, contents []string) float64 {
	score := 5.0 // 基础分

	// 深度关键词
	depthKeywords := []string{"因为", "所以", "因此", "然而", "但是", "首先", "其次", "最后",
		"本质", "核心", "根本", "原理", "逻辑", "证据", "数据", "事实", "第一性原理",
		"experience", "fundamental", "principle", "evidence", "data", "logic", "because", "therefore"}
	for _, kw := range depthKeywords {
		score += float64(strings.Count(allText, kw)) * 0.8
	}

	// 平均消息长度（中文字符）
	var totalLen int
	for _, c := range contents {
		totalLen += utf8.RuneCountInString(c)
	}
	avgLen := float64(totalLen) / float64(len(contents))
	if avgLen >= 150 {
		score += 8
	} else if avgLen >= 80 {
		score += 4
	}

	// 消息数量越多，覆盖越广
	if len(contents) >= 4 {
		score += 4
	} else if len(contents) >= 2 {
		score += 2
	}

	// 观点对比标志
	contrastWords := []string{"相反", "对比", "不同于", "相比之下", "相反地", "versus", "contrast", "unlike"}
	for _, cw := range contrastWords {
		if strings.Contains(allText, cw) {
			score += 2
			break
		}
	}

	return clamp(score, 0, 25)
}

// scoreConsistency 基于角色标志性表达和立场一致性评估。
func scoreConsistency(contents []string) float64 {
	score := 5.0

	// Steve Jobs 标志性表达
	jobsMarkers := []string{"simplicity", "simple", "experience", "product", "taste", "design",
		"简洁", "体验", "产品", "品味", "设计", "用户", "伟大", "平庸"}
	// Elon Musk 标志性表达
	muskMarkers := []string{"physics", "first principles", "manufacturing", "engineer", "speed", "timeline",
		"物理", "第一性原理", "工程", "制造", "速度", "时间线", "成本", "计算"}

	var jobsCount, muskCount int
	for _, c := range contents {
		lower := strings.ToLower(c)
		for _, m := range jobsMarkers {
			if strings.Contains(lower, m) {
				jobsCount++
			}
		}
		for _, m := range muskMarkers {
			if strings.Contains(lower, m) {
				muskCount++
			}
		}
	}

	// 只要有角色相关表达就加分
	if jobsCount > 0 {
		score += 6
	}
	if muskCount > 0 {
		score += 6
	}
	if jobsCount > 2 || muskCount > 2 {
		score += 4
	}

	// 立场强度词
	stanceWords := []string{"必须", "一定", "绝不", "根本", "显然", "绝对",
		"must", "never", "absolutely", "obviously", "fundamentally"}
	for _, sw := range stanceWords {
		if strings.Contains(strings.ToLower(allText(contents)), sw) {
			score += 2
			break
		}
	}

	return clamp(score, 0, 25)
}

// scoreInteraction 基于回应、提问和反驳标志评估交互质量。
func scoreInteraction(contents []string) float64 {
	score := 5.0
	allText := allText(contents)

	// 回应标志
	responseMarkers := []string{"同意", "不同意", "反驳", "回应", "针对", "你说", "你的观点",
		"agree", "disagree", "反驳", "response", "your point", "you said"}
	for _, rm := range responseMarkers {
		if strings.Contains(allText, rm) {
			score += 3
			break
		}
	}

	// 提问标志
	questionMarkers := []string{"？", "?", "为什么", "怎么", "如何", "what", "why", "how"}
	for _, qm := range questionMarkers {
		if strings.Contains(allText, qm) {
			score += 3
			break
		}
	}

	// 多轮次有内容
	if len(contents) >= 4 {
		score += 6
	} else if len(contents) >= 2 {
		score += 3
	}

	// 观点推进标志（不是重复）
	progressMarkers := []string{"进一步", "补充", "另外", "此外", "更重要的是",
		"furthermore", "additionally", "more importantly", "beyond"}
	for _, pm := range progressMarkers {
		if strings.Contains(allText, pm) {
			score += 4
			break
		}
	}

	// 对话感标志
	dialogueMarkers := []string{"直接", "点名", "回应", "对话", "交流",
		"directly", "address", "reply", "conversation"}
	for _, dm := range dialogueMarkers {
		if strings.Contains(allText, dm) {
			score += 4
			break
		}
	}

	return clamp(score, 0, 25)
}

// scoreNaturalness 基于口语化表达、避免模板化和流畅度评估。
func scoreNaturalness(contents []string) float64 {
	score := 5.0
	allText := allText(contents)

	// 口语化标志
	colloquial := []string{"其实", "说实话", "我觉得", "我认为", "坦白说", "说白了",
		"honestly", "I think", "I believe", "frankly", "basically", "actually"}
	for _, c := range colloquial {
		if strings.Contains(allText, c) {
			score += 3
			break
		}
	}

	// 避免过度模板化（检查列表化输出）
	listPatterns := []string{"1.", "2.", "3.", "一、", "二、", "三、", "首先", "其次", "最后"}
	listCount := 0
	for _, lp := range listPatterns {
		listCount += strings.Count(allText, lp)
	}
	if listCount <= 2 {
		score += 5 // 较少列表化，更自然
	} else if listCount <= 5 {
		score += 2
	}

	// 句子长度变化（简单估计：有短句有长句）
	var hasShort, hasLong bool
	for _, c := range contents {
		sentences := strings.Split(c, "。")
		for _, s := range sentences {
			runes := utf8.RuneCountInString(s)
			if runes > 0 && runes < 15 {
				hasShort = true
			}
			if runes > 40 {
				hasLong = true
			}
		}
	}
	if hasShort && hasLong {
		score += 5 // 有节奏变化
	} else if hasShort || hasLong {
		score += 2
	}

	// 情感表达
	emotionMarkers := []string{"！", "!", "有趣", "荒谬", "精彩", "糟糕", "amazing", "ridiculous", "terrible", "great"}
	for _, em := range emotionMarkers {
		if strings.Contains(allText, em) {
			score += 3
			break
		}
	}

	// 平均长度适中（200-500 字是期望范围）
	var totalLen int
	for _, c := range contents {
		totalLen += utf8.RuneCountInString(c)
	}
	avgLen := float64(totalLen) / float64(len(contents))
	if avgLen >= 100 && avgLen <= 600 {
		score += 4
	} else if avgLen >= 50 {
		score += 2
	}

	return clamp(score, 0, 25)
}

func allText(contents []string) string {
	return strings.Join(contents, " ")
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func depthReason(score float64) string {
	if score >= 20 {
		return "内容论证充分，包含多层次的逻辑展开和关键概念"
	}
	if score >= 15 {
		return "内容有一定深度，包含论证和关键概念"
	}
	if score >= 10 {
		return "内容深度一般，有基本观点但缺乏深入论证"
	}
	return "内容较为浅显，缺乏深度论证"
}

func consistencyReason(score float64) string {
	if score >= 20 {
		return "角色表达高度一致，标志性语言和立场清晰"
	}
	if score >= 15 {
		return "角色表达较为一致，能识别角色特征"
	}
	if score >= 10 {
		return "角色一致性一般，部分表达符合角色设定"
	}
	return "角色一致性较弱，难以识别特定角色特征"
}

func interactionReason(score float64) string {
	if score >= 20 {
		return "交互质量高，有明确回应、反驳和观点推进"
	}
	if score >= 15 {
		return "交互质量较好，存在回应和观点交流"
	}
	if score >= 10 {
		return "交互质量一般，有一定交流但缺乏深度互动"
	}
	return "交互质量较弱，讨论更像独白而非对话"
}

func naturalnessReason(score float64) string {
	if score >= 20 {
		return "表达非常自然，口语化、有情感、节奏变化丰富"
	}
	if score >= 15 {
		return "表达较为自然，有口语化特征和情感"
	}
	if score >= 10 {
		return "表达自然度一般，部分口语化但偏模板化"
	}
	return "表达较为机械，模板化或列表化倾向明显"
}
