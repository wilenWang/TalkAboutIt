// Package llm_test 提供 TalkAboutIt 的 LLM Provider 抽象与实现测试。
package llm_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/wilenwang/talkaboutit/internal/config"
	"github.com/wilenwang/talkaboutit/internal/llm"
)

// TestFactory_CreateOpenAI 验证工厂能根据 openai 类型配置创建 OpenAIProvider。
func TestFactory_CreateOpenAI(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Default: "deepseek",
			Providers: map[string]config.ProviderConfig{
				"deepseek": {
					Type:    "openai",
					BaseURL: "https://api.deepseek.com",
					APIKey:  "test-key",
					Model:   "deepseek-v4-pro",
				},
			},
		},
	}

	prov, err := llm.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	if prov == nil {
		t.Fatal("expected provider, got nil")
	}

	openAIProv, ok := prov.(*llm.OpenAIProvider)
	if !ok {
		t.Fatalf("expected *OpenAIProvider, got %T", prov)
	}

	if openAIProv.Name() != "deepseek" {
		t.Errorf("expected name deepseek, got %s", openAIProv.Name())
	}
	if openAIProv.Model() != "deepseek-v4-pro" {
		t.Errorf("expected model deepseek-v4-pro, got %s", openAIProv.Model())
	}
}

// TestFactory_CreateAnthropic 验证工厂能根据 anthropic 类型配置创建 AnthropicProvider（占位）。
func TestFactory_CreateAnthropic(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Default: "claude",
			Providers: map[string]config.ProviderConfig{
				"claude": {
					Type:   "anthropic",
					APIKey: "test-key",
					Model:  "claude-sonnet-4-20250514",
				},
			},
		},
	}

	prov, err := llm.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	anthProv, ok := prov.(*llm.AnthropicProvider)
	if !ok {
		t.Fatalf("expected *AnthropicProvider, got %T", prov)
	}

	if anthProv.Name() != "claude" {
		t.Errorf("expected name claude, got %s", anthProv.Name())
	}
	if anthProv.Model() != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %s", anthProv.Model())
	}
}

// TestFactory_UnsupportedType 验证工厂对不支持的类型返回错误。
func TestFactory_UnsupportedType(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Default: "unknown",
			Providers: map[string]config.ProviderConfig{
				"unknown": {
					Type:   "gemini",
					APIKey: "test-key",
				},
			},
		},
	}

	_, err := llm.NewProvider(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported provider type")
	}
}

// TestFactory_DefaultFallback 验证无配置时工厂回退到默认 DeepSeek 配置。
func TestFactory_DefaultFallback(t *testing.T) {
	cfg := config.Config{
		LLM: config.LLMConfig{
			Default:   "",
			Providers: map[string]config.ProviderConfig{},
		},
	}

	// 空配置回退到 deepseek，需要环境变量提供 key
	t.Setenv("DEEPSEEK_API_KEY", "fallback-test-key")

	prov, err := llm.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	openAIProv, ok := prov.(*llm.OpenAIProvider)
	if !ok {
		t.Fatalf("expected *OpenAIProvider fallback, got %T", prov)
	}

	if openAIProv.Model() != "deepseek-v4-pro" {
		t.Errorf("expected default model deepseek-v4-pro, got %s", openAIProv.Model())
	}
}

// TestMockGenerateFallback 验证 engine 在 generate 为 nil 时回退到 mock generate。
// 由于 mock generate 逻辑在 engine 包中，这里仅验证工厂层面：当配置缺失时 provider 仍可创建（fallback）。
func TestMockGenerateFallback(t *testing.T) {
	// 空配置应触发 fallback，返回 OpenAIProvider（默认 deepseek）
	cfg := config.Config{}
	t.Setenv("DEEPSEEK_API_KEY", "fallback-test-key")
	prov, err := llm.NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider with empty config failed: %v", err)
	}
	if prov == nil {
		t.Fatal("expected fallback provider, got nil")
	}
}

// TestNewProviderError_Classification 验证 HTTP 状态码到 ProviderError 的分类映射。
func TestNewProviderError_Classification(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   string
		wantRec    bool
	}{
		{"401 认证失败", 401, "PROVIDER_AUTH", false},
		{"403 认证失败", 403, "PROVIDER_AUTH", false},
		{"429 速率限制", 429, "PROVIDER_RATE_LIMIT", true},
		{"500 服务端错误", 500, "PROVIDER_SERVER_ERROR", true},
		{"502 网关错误", 502, "PROVIDER_SERVER_ERROR", true},
		{"503 服务不可用", 503, "PROVIDER_SERVER_ERROR", true},
		{"418 未知错误", 418, "PROVIDER_UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := llm.NewProviderError(tt.statusCode, "test body", nil)
			if err.Code != tt.wantCode {
				t.Errorf("status %d: 期望 code %s，得到 %s", tt.statusCode, tt.wantCode, err.Code)
			}
			if err.Recoverable != tt.wantRec {
				t.Errorf("status %d: 期望 recoverable=%v，得到 %v", tt.statusCode, tt.wantRec, err.Recoverable)
			}
		})
	}
}

func TestChatMessage_JSONNameOmitEmpty(t *testing.T) {
	msg := llm.ChatMessage{Role: "assistant", Content: "hello"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(data) != `{"role":"assistant","content":"hello"}` {
		t.Fatalf("unexpected json: %s", data)
	}
}

func TestChatUsage_JSONFields(t *testing.T) {
	usage := llm.ChatUsage{
		PromptTokens:          10,
		CompletionTokens:      5,
		TotalTokens:           15,
		PromptCacheHitTokens:  8,
		PromptCacheMissTokens: 2,
	}
	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	got := string(data)
	for _, field := range []string{
		`"prompt_tokens":10`,
		`"completion_tokens":5`,
		`"total_tokens":15`,
		`"prompt_cache_hit_tokens":8`,
		`"prompt_cache_miss_tokens":2`,
	} {
		if !strings.Contains(got, field) {
			t.Fatalf("usage json should contain %s, got %s", field, got)
		}
	}
}
