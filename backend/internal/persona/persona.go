// Package persona 提供 Persona Schema v1 的数据模型与 system prompt 构建能力。
package persona



// Persona 是 TalkAboutIt v1 的统一角色描述结构体（Persona Schema v1）。
type Persona struct {
	SchemaVersion   string          `json:"schema_version"`
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	DisplayName     string          `json:"display_name"`
	Avatar          string          `json:"avatar"`
	RoleTitle       string          `json:"role_title"`
	Description     string          `json:"description"`
	Tags            []string        `json:"tags"`
	Language        Language        `json:"language"`
	Stance          Stance          `json:"stance"`
	CoreBeliefs     []CoreBelief    `json:"core_beliefs"`
	SpeakingStyle   SpeakingStyle   `json:"speaking_style"`
	KnowledgeScope  KnowledgeScope  `json:"knowledge_scope"`
	InteractionRules InteractionRules `json:"interaction_rules"`
	DebateGoal      DebateGoal      `json:"debate_goal"`
	Prompting       Prompting       `json:"prompting"`
	Examples        Examples        `json:"examples"`
}

// Language 定义角色的语言偏好。
type Language struct {
	Primary       string   `json:"primary"`
	Allowed       []string `json:"allowed"`
	DefaultOutput string   `json:"default_output"`
	StyleHint     string   `json:"style_hint"`
}

// Stance 定义角色的默认立场与倾向。
type Stance struct {
	DefaultPosition string   `json:"default_position"`
	Intensity       int      `json:"intensity"`
	Biases          []string `json:"biases"`
	Taboos          []string `json:"taboos"`
}

// CoreBelief 定义角色的核心信念。
type CoreBelief struct {
	Belief    string `json:"belief"`
	Priority  int    `json:"priority"`
	Rationale string `json:"rationale"`
}

// SpeakingStyle 定义角色的说话风格。
type SpeakingStyle struct {
	Tone              string   `json:"tone"`
	Cadence           string   `json:"cadence"`
	Verbosity         int      `json:"verbosity"`
	SignaturePatterns []string `json:"signature_patterns"`
	Do                []string `json:"do"`
	Dont              []string `json:"dont"`
}

// KnowledgeScope 定义角色的知识边界。
type KnowledgeScope struct {
	Domains           []string       `json:"domains"`
	ExpertiseLevel    map[string]int `json:"expertise_level"`
	TimeCutoff        string         `json:"time_cutoff"`
	AllowedInference  string         `json:"allowed_inference"`
	UnknownHandling   string         `json:"unknown_handling"`
	ForbiddenClaims   []string       `json:"forbidden_claims"`
}

// InteractionRules 定义角色在讨论中的互动规则。
type InteractionRules struct {
	AddressOthers      string   `json:"address_others"`
	DisagreementStyle  string   `json:"disagreement_style"`
	InterruptionPolicy string   `json:"interruption_policy"`
	QuestionPolicy     string   `json:"question_policy"`
	ConcessionPolicy   string   `json:"concession_policy"`
	Avoid              []string `json:"avoid"`
}

// DebateGoal 定义角色的辩论目标。
type DebateGoal struct {
	PrimaryGoal    string   `json:"primary_goal"`
	SecondaryGoals []string `json:"secondary_goals"`
	WinCondition   string   `json:"win_condition"`
	LossCondition  string   `json:"loss_condition"`
}

// Prompting 定义额外的提示词约束。
type Prompting struct {
	SystemPreamble    string   `json:"system_preamble"`
	ReplyConstraints  []string `json:"reply_constraints"`
}

// Examples 提供角色的发言示例。
type Examples struct {
	OpeningLine     string `json:"opening_line"`
	SampleRebuttal  string `json:"sample_rebuttal"`
}


