package persona

import (
	"testing"
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


