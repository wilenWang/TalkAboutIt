// Package config 提供 TalkAboutIt 的配置加载与管理。
// 支持从 YAML 文件读取，并通过环境变量覆盖指定字段。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 是应用全局配置的根结构体。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	LLM      LLMConfig      `yaml:"llm"`
	Personas PersonasConfig `yaml:"personas"`
	Session  SessionConfig  `yaml:"session"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig 是 HTTP 服务相关配置。
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// LLMConfig 是 LLM Gateway 相关配置。
type LLMConfig struct {
	Default   string                    `yaml:"default"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig 是单个 LLM Provider 的配置。
type ProviderConfig struct {
	Type    string `yaml:"type"`
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

// PersonasConfig 是 Persona 资产目录配置。
type PersonasConfig struct {
	Dir string `yaml:"dir"`
}

// SessionConfig 是会话与持久化相关配置。
type SessionConfig struct {
	DBPath     string `yaml:"db_path"`
	MaxRounds  int    `yaml:"max_rounds"`
}

// DatabaseConfig 是数据库相关配置（别名，兼容不同命名习惯）。
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// DefaultConfig 返回带有默认值的 Config 实例。
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		LLM: LLMConfig{
			Default:   "deepseek",
			Providers: make(map[string]ProviderConfig),
		},
		Personas: PersonasConfig{
			Dir: "personas",
		},
		Session: SessionConfig{
			DBPath:    "data/sessions.db",
			MaxRounds: 3,
		},
	}
}

// Load 从指定路径加载 YAML 配置文件，并应用环境变量覆盖。
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 环境变量覆盖
	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides 通过环境变量覆盖配置字段。
// 支持的环境变量：
//   - TALKABOUTIT_SERVER_PORT
//   - TALKABOUTIT_SERVER_HOST
//   - TALKABOUTIT_LLM_DEFAULT
//   - TALKABOUTIT_LLM_{PROVIDER}_API_KEY
//   - TALKABOUTIT_LLM_{PROVIDER}_BASE_URL
//   - TALKABOUTIT_LLM_{PROVIDER}_MODEL
//   - TALKABOUTIT_PERSONAS_DIR
//   - TALKABOUTIT_SESSION_DB_PATH
//   - TALKABOUTIT_SESSION_MAX_ROUNDS
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("TALKABOUTIT_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("TALKABOUTIT_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}

	if v := os.Getenv("TALKABOUTIT_LLM_DEFAULT"); v != "" {
		cfg.LLM.Default = v
	}

	// 遍历所有已配置的 provider，检查对应的环境变量
	for name, prov := range cfg.LLM.Providers {
		prefix := "TALKABOUTIT_LLM_" + strings.ToUpper(name)
		if v := os.Getenv(prefix + "_API_KEY"); v != "" {
			prov.APIKey = v
		}
		if v := os.Getenv(prefix + "_BASE_URL"); v != "" {
			prov.BaseURL = v
		}
		if v := os.Getenv(prefix + "_MODEL"); v != "" {
			prov.Model = v
		}
		cfg.LLM.Providers[name] = prov
	}

	if v := os.Getenv("TALKABOUTIT_PERSONAS_DIR"); v != "" {
		cfg.Personas.Dir = v
	}

	if v := os.Getenv("TALKABOUTIT_SESSION_DB_PATH"); v != "" {
		cfg.Session.DBPath = v
	}
	if v := os.Getenv("TALKABOUTIT_SESSION_MAX_ROUNDS"); v != "" {
		if mr, err := strconv.Atoi(v); err == nil {
			cfg.Session.MaxRounds = mr
		}
	}
}
