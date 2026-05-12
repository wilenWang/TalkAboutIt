// Package llm 提供 TalkAboutIt 的 LLM Provider 错误分类。
package llm

import "fmt"

// ProviderError 是 LLM Provider 返回的结构化错误。
type ProviderError struct {
	Code          string // 错误码，用于前端与引擎识别
	Recoverable   bool   // 是否可恢复（可恢复则跳过当前发言，继续下一人）
	UserMessage   string // 用户可见的友好文案
	InternalError error  // 原始内部错误（可选，用于日志）
}

func (e *ProviderError) Error() string {
	if e.InternalError != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.UserMessage, e.InternalError)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.UserMessage)
}

func (e *ProviderError) Unwrap() error {
	return e.InternalError
}

// 预定义错误类型
var (
	// ErrProviderTimeout 请求 LLM API 超时，可恢复。
	ErrProviderTimeout = &ProviderError{
		Code:        "PROVIDER_TIMEOUT",
		Recoverable: true,
		UserMessage: "AI 服务响应超时，已跳过当前发言，继续下一位。",
	}

	// ErrProviderRateLimit 触发 LLM API 速率限制，可恢复。
	ErrProviderRateLimit = &ProviderError{
		Code:        "PROVIDER_RATE_LIMIT",
		Recoverable: true,
		UserMessage: "AI 服务繁忙（速率限制），已跳过当前发言，继续下一位。",
	}

	// ErrProviderAuth LLM API 认证失败，不可恢复。
	ErrProviderAuth = &ProviderError{
		Code:        "PROVIDER_AUTH",
		Recoverable: false,
		UserMessage: "AI 服务认证失败，请检查 API Key 配置。",
	}

	// ErrProviderServer LLM API 服务端错误（5xx），可恢复。
	ErrProviderServer = &ProviderError{
		Code:        "PROVIDER_SERVER_ERROR",
		Recoverable: true,
		UserMessage: "AI 服务暂时不可用，已跳过当前发言，继续下一位。",
	}

	// ErrDBError 数据库操作失败，不可恢复。
	ErrDBError = &ProviderError{
		Code:        "DB_ERROR",
		Recoverable: false,
		UserMessage: "系统内部错误，讨论已终止。",
	}
)

// NewProviderError 根据 HTTP 状态码创建对应的 ProviderError。
func NewProviderError(statusCode int, body string, internalErr error) *ProviderError {
	switch {
	case statusCode == 401 || statusCode == 403:
		return &ProviderError{
			Code:        ErrProviderAuth.Code,
			Recoverable: false,
			UserMessage: ErrProviderAuth.UserMessage,
			InternalError: internalErr,
		}
	case statusCode == 429:
		return &ProviderError{
			Code:        ErrProviderRateLimit.Code,
			Recoverable: true,
			UserMessage: ErrProviderRateLimit.UserMessage,
			InternalError: internalErr,
		}
	case statusCode >= 500:
		return &ProviderError{
			Code:        ErrProviderServer.Code,
			Recoverable: true,
			UserMessage: ErrProviderServer.UserMessage,
			InternalError: internalErr,
		}
	default:
		return &ProviderError{
			Code:        "PROVIDER_UNKNOWN",
			Recoverable: false,
			UserMessage: "AI 服务返回未知错误：" + body,
			InternalError: internalErr,
		}
	}
}
