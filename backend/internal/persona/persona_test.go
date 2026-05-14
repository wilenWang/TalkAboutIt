package persona

import (
	"testing"

	"github.com/wilenwang/talkaboutit/internal/llm"
)

// TestLoaderLoadAll 验证 Loader 能正确加载所有 persona JSON。
func TestLoaderLoadAll(t *testing.T) {
	loader := NewLoader("../../personas")
	personas, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("加载 persona 失败: %v", err)
	}

	expectedIDs := []string{
		"steve-jobs",
		"elon-musk",
		"naval-ravikant",
		"zhang-yiming",
		"zhang-xiaolong",
	}

	if len(personas) != len(expectedIDs) {
		t.Errorf("期望加载 %d 个 persona，实际加载 %d 个", len(expectedIDs), len(personas))
	}

	for _, id := range expectedIDs {
		p, ok := personas[id]
		if !ok {
			t.Errorf("期望存在 persona %q，但未找到", id)
			continue
		}
		if p.SchemaVersion != "persona.v1" {
			t.Errorf("%s 的 schema_version 期望 persona.v1，得到 %s", id, p.SchemaVersion)
		}
		if p.ID != id {
			t.Errorf("%s 的 id 期望 %s，得到 %s", id, id, p.ID)
		}
		if p.Name == "" {
			t.Errorf("%s 的 name 不应为空", id)
		}
	}
}

// TestLoaderLoadOne 验证 Loader 能正确加载单个 persona。
func TestLoaderLoadOne(t *testing.T) {
	loader := NewLoader("../../personas")
	p, err := loader.LoadOne("steve-jobs")
	if err != nil {
		t.Fatalf("加载 steve-jobs 失败: %v", err)
	}

	if p.ID != "steve-jobs" {
		t.Errorf("id 期望 steve-jobs，得到 %s", p.ID)
	}
	if p.Name != "Steve Jobs" {
		t.Errorf("name 期望 Steve Jobs，得到 %s", p.Name)
	}
}

func TestConversationContext_BuildAndTruncate(t *testing.T) {
	ctx := NewConversationContext("steve-jobs", "static-system", &PerPersonaState{})
	ctx.Append("user", "", "round1")
	ctx.Append("assistant", "steve-jobs", "reply1")
	ctx.Append("assistant", "elon-musk", "peer1")
	ctx.Append("user", "", "round2")
	ctx.Append("assistant", "steve-jobs", "reply2")

	req := ctx.BuildChatRequest(512, 0.8)
	if len(req.Messages) != len(ctx.Messages) {
		t.Fatalf("BuildChatRequest 应复制全部消息，got=%d want=%d", len(req.Messages), len(ctx.Messages))
	}
	if req.Messages[0].Role != "system" || req.Messages[0].Content != "static-system" {
		t.Fatalf("首条消息应为 system，got=%+v", req.Messages[0])
	}

	ctx.Truncate(4)
	if len(ctx.Messages) != 4 {
		t.Fatalf("Truncate 后应保留 4 条消息，got=%d", len(ctx.Messages))
	}
	if ctx.Messages[0].Role != "system" {
		t.Fatalf("Truncate 后首条仍应为 system，got=%+v", ctx.Messages[0])
	}
	if ctx.Messages[1].Role != "system" || ctx.Messages[1].Name != "" {
		t.Fatalf("Truncate 后第二条应为 system 摘要消息（Name 为空），got=%+v", ctx.Messages[1])
	}
	if ctx.Messages[3].Content != "reply2" {
		t.Fatalf("Truncate 后应保留最近消息，got=%+v", ctx.Messages[3])
	}
}

func TestConversationContext_AppendPreservesName(t *testing.T) {
	ctx := NewConversationContext("elon-musk", "system", nil)
	ctx.Append("assistant", "steve-jobs", "Stay hungry.")

	got := ctx.Messages[len(ctx.Messages)-1]
	want := llm.ChatMessage{Role: "assistant", Name: "steve-jobs", Content: "Stay hungry."}
	if got != want {
		t.Fatalf("unexpected message: got=%+v want=%+v", got, want)
	}
}
