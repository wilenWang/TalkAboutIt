package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wilenwang/talkaboutit/internal/api"
	"github.com/wilenwang/talkaboutit/internal/config"
	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("配置加载失败（使用默认）: %v", err)
		cfg = config.DefaultConfig()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 初始化 persona loader
	personaDir := cfg.Personas.Dir
	if personaDir == "" {
		personaDir = "personas"
	}
	// 支持从工作目录或二进制所在目录解析相对路径
	if !filepath.IsAbs(personaDir) {
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			alt := filepath.Join(exeDir, personaDir)
			if _, err := os.Stat(alt); err == nil {
				personaDir = alt
			}
		}
	}
	loader := persona.NewLoader(personaDir)

	// 初始化 SQLite
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "data/talkaboutit.db"
	}
	if !filepath.IsAbs(dbPath) {
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			alt := filepath.Join(exeDir, dbPath)
			if _, err := os.Stat(filepath.Dir(alt)); err == nil {
				dbPath = alt
			}
		}
	}
	store, err := session.NewStore(dbPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer store.Close()

	// 初始化 LLM Provider（通过工厂）
	var generate engine.GenerateFunc
	provider, err := llm.NewProvider(*cfg)
	if err != nil {
		log.Printf("LLM Provider 初始化失败，回退到 mock: %v", err)
		generate = nil // engine.NewEngine 内部会 fallback 到 DefaultMockGenerate
	} else {
		log.Printf("LLM Provider 初始化成功: %s / %s", provider.Name(), provider.Model())
		generate = engine.LLMGenerate(provider)
	}

	// 初始化 engine
	eng := engine.NewEngine(store, loader, generate)

	// 初始化 API handler
	handler := api.NewHandler(loader, store, eng)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := ":" + port
	log.Printf("TalkAboutIt server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
