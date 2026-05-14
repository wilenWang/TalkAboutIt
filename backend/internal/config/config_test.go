package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfig 验证默认配置值是否正确。
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8080 {
		t.Errorf("默认端口期望 8080，得到 %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("默认主机期望 0.0.0.0，得到 %s", cfg.Server.Host)
	}
	if cfg.LLM.Default != "deepseek" {
		t.Errorf("默认 LLM 期望 deepseek，得到 %s", cfg.LLM.Default)
	}
	if cfg.Personas.Dir != "personas" {
		t.Errorf("默认 persona 目录期望 personas，得到 %s", cfg.Personas.Dir)
	}
	if cfg.Session.DBPath != "data/sessions.db" {
		t.Errorf("默认数据库路径期望 data/sessions.db，得到 %s", cfg.Session.DBPath)
	}
	if cfg.Session.MaxRounds != 3 {
		t.Errorf("默认最大轮数期望 3，得到 %d", cfg.Session.MaxRounds)
	}
}

// TestLoadFromYAML 验证从 YAML 文件加载配置。
func TestLoadFromYAML(t *testing.T) {
	// 创建临时目录与配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	content := `
server:
  port: 9090
  host: "127.0.0.1"
llm:
  default: deepseek
  providers:
    deepseek:
      type: openai
      api_key: "sk-test"
      base_url: "https://api.deepseek.com"
      model: deepseek-v4-pro
personas:
  dir: "custom_personas"
session:
  db_path: "custom.db"
  max_rounds: 5
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置文件失败: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("端口期望 9090，得到 %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("主机期望 127.0.0.1，得到 %s", cfg.Server.Host)
	}
	if cfg.LLM.Default != "deepseek" {
		t.Errorf("默认 LLM 期望 deepseek，得到 %s", cfg.LLM.Default)
	}
	if prov, ok := cfg.LLM.Providers["deepseek"]; !ok {
		t.Errorf("期望存在 deepseek provider")
	} else {
		if prov.APIKey != "sk-test" {
			t.Errorf("deepseek api_key 期望 sk-test，得到 %s", prov.APIKey)
		}
		if prov.Model != "deepseek-v4-pro" {
			t.Errorf("deepseek model 期望 deepseek-v4-pro，得到 %s", prov.Model)
		}
	}
	if cfg.Personas.Dir != "custom_personas" {
		t.Errorf("persona 目录期望 custom_personas，得到 %s", cfg.Personas.Dir)
	}
	if cfg.Session.DBPath != "custom.db" {
		t.Errorf("数据库路径期望 custom.db，得到 %s", cfg.Session.DBPath)
	}
	if cfg.Session.MaxRounds != 5 {
		t.Errorf("最大轮数期望 5，得到 %d", cfg.Session.MaxRounds)
	}
}

// TestEnvOverride 验证环境变量覆盖逻辑。
func TestEnvOverride(t *testing.T) {
	// 创建最小化配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "env_config.yaml")
	content := `
server:
  port: 8080
llm:
  default: openai
  providers:
    openai:
      type: openai
      api_key: "file-key"
      model: gpt-4o
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置文件失败: %v", err)
	}

	// 设置环境变量
	os.Setenv("TALKABOUTIT_SERVER_PORT", "7777")
	os.Setenv("TALKABOUTIT_LLM_DEFAULT", "claude")
	os.Setenv("TALKABOUTIT_LLM_OPENAI_API_KEY", "env-key")
	os.Setenv("TALKABOUTIT_LLM_OPENAI_MODEL", "gpt-4o-mini")
	defer func() {
		os.Unsetenv("TALKABOUTIT_SERVER_PORT")
		os.Unsetenv("TALKABOUTIT_LLM_DEFAULT")
		os.Unsetenv("TALKABOUTIT_LLM_OPENAI_API_KEY")
		os.Unsetenv("TALKABOUTIT_LLM_OPENAI_MODEL")
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	if cfg.Server.Port != 7777 {
		t.Errorf("端口期望被环境变量覆盖为 7777，得到 %d", cfg.Server.Port)
	}
	if cfg.LLM.Default != "claude" {
		t.Errorf("默认 LLM 期望被覆盖为 claude，得到 %s", cfg.LLM.Default)
	}
	if prov := cfg.LLM.Providers["openai"]; prov.APIKey != "env-key" {
		t.Errorf("openai api_key 期望被覆盖为 env-key，得到 %s", prov.APIKey)
	}
	if prov := cfg.LLM.Providers["openai"]; prov.Model != "gpt-4o-mini" {
		t.Errorf("openai model 期望被覆盖为 gpt-4o-mini，得到 %s", prov.Model)
	}
}

// TestLoadMissingFile 验证加载不存在的文件应返回错误。
func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("加载不存在的配置文件应返回错误")
	}
}
