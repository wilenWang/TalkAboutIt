// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/wilenwang/talkaboutit/internal/config"
)

// resolveAPIKey 按优先级为指定 provider 解析 API Key。
// 1. 使用配置中已设置的 APIKey；
// 2. 若为空，按 provider 名称匹配常见环境变量（如 DEEPSEEK_API_KEY、OPENAI_API_KEY）；
// 3. 仍为空，尝试 TALKABOUTIT_LLM_{NAME}_API_KEY；
// 4. 若全部为空，返回错误。
func resolveAPIKey(name string, provCfg config.ProviderConfig) (string, error) {
	if provCfg.APIKey != "" {
		return provCfg.APIKey, nil
	}

	// 按名称匹配常见环境变量
	lowerName := strings.ToLower(name)
	switch lowerName {
	case "deepseek":
		if v := os.Getenv("DEEPSEEK_API_KEY"); v != "" {
			return v, nil
		}
	case "openai":
		if v := os.Getenv("OPENAI_API_KEY"); v != "" {
			return v, nil
		}
	}

	// 通用环境变量 fallback
	envKey := "TALKABOUTIT_LLM_" + strings.ToUpper(name) + "_API_KEY"
	if v := os.Getenv(envKey); v != "" {
		return v, nil
	}

	return "", fmt.Errorf("provider %q 缺少 API Key（请配置 api_key 或设置对应环境变量）", name)
}

// NewProvider 根据配置创建默认的 LLM Provider。
// 优先使用 llm.default 指定的 provider；若未配置或失败，则尝试第一个可用的 provider。
func NewProvider(cfg config.Config) (Provider, error) {
	defaultName := cfg.LLM.Default
	if defaultName == "" {
		defaultName = "deepseek"
	}

	provCfg, ok := cfg.LLM.Providers[defaultName]
	if !ok {
		// 尝试回退到第一个可用的 provider
		for name, p := range cfg.LLM.Providers {
			defaultName = name
			provCfg = p
			break
		}
	}

	// 若仍无配置，使用环境变量构造 DeepSeek 默认配置
	if provCfg.Type == "" {
		provCfg = config.ProviderConfig{
			Type:    "openai",
			BaseURL: "https://api.deepseek.com/v1",
			Model:   "deepseek-chat",
		}
	}

	// 统一补齐 API Key
	apiKey, err := resolveAPIKey(defaultName, provCfg)
	if err != nil {
		return nil, err
	}
	provCfg.APIKey = apiKey

	switch provCfg.Type {
	case "openai":
		baseURL := provCfg.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
		model := provCfg.Model
		if model == "" {
			model = "deepseek-chat"
		}
		return NewOpenAIProvider(defaultName, model, provCfg.APIKey, baseURL, nil), nil
	case "anthropic":
		model := provCfg.Model
		if model == "" {
			model = "claude-sonnet-4-20250514"
		}
		return NewAnthropicProvider(defaultName, model), nil
	default:
		return nil, fmt.Errorf("不支持的 LLM provider 类型: %s", provCfg.Type)
	}
}
