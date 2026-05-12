package persona

import (
	"strings"
	"testing"
)

func TestValidateValidPersona(t *testing.T) {
	p := Persona{
		SchemaVersion: "persona.v1",
		ID:            "test",
		Name:          "Test",
		Description:   "A test persona.",
		Stance:        Stance{Intensity: 3},
		SpeakingStyle: SpeakingStyle{Verbosity: 3},
	}
	if err := p.Validate(); err != nil {
		t.Errorf("合法 persona 不应校验失败: %v", err)
	}
}

func TestValidateMissingRequiredFields(t *testing.T) {
	p := Persona{}
	err := p.Validate()
	if err == nil {
		t.Fatal("空 persona 应校验失败")
	}
	msg := err.Error()
	for _, want := range []string{"schema_version", "id", "name", "description"} {
		if !contains(msg, want) {
			t.Errorf("错误信息应包含 %q", want)
		}
	}
}

func TestValidateIntensityOutOfRange(t *testing.T) {
	p := Persona{
		SchemaVersion: "persona.v1",
		ID:            "test",
		Name:          "Test",
		Description:   "desc",
		Stance:        Stance{Intensity: 6},
		SpeakingStyle: SpeakingStyle{Verbosity: 3},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("intensity 超出范围应校验失败")
	}
	if !contains(err.Error(), "intensity") {
		t.Error("错误信息应包含 intensity")
	}
}

func TestValidateVerbosityOutOfRange(t *testing.T) {
	p := Persona{
		SchemaVersion: "persona.v1",
		ID:            "test",
		Name:          "Test",
		Description:   "desc",
		Stance:        Stance{Intensity: 3},
		SpeakingStyle: SpeakingStyle{Verbosity: 0},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("verbosity 超出范围应校验失败")
	}
	if !contains(err.Error(), "verbosity") {
		t.Error("错误信息应包含 verbosity")
	}
}

func TestValidateCoreBeliefPriorityOutOfRange(t *testing.T) {
	p := Persona{
		SchemaVersion: "persona.v1",
		ID:            "test",
		Name:          "Test",
		Description:   "desc",
		Stance:        Stance{Intensity: 3},
		CoreBeliefs:   []CoreBelief{{Belief: "b", Priority: 6}},
		SpeakingStyle: SpeakingStyle{Verbosity: 3},
	}
	err := p.Validate()
	if err == nil {
		t.Fatal("core belief priority 超出范围应校验失败")
	}
	if !contains(err.Error(), "priority") {
		t.Error("错误信息应包含 priority")
	}
}

func TestValidateJSONUnknownField(t *testing.T) {
	data := []byte(`{
		"schema_version": "persona.v1",
		"id": "test",
		"name": "Test",
		"description": "desc",
		"stance": {"intensity": 3},
		"speaking_style": {"verbosity": 3},
		"unknown_field": "should fail"
	}`)
	_, err := ValidateJSON(data)
	if err == nil {
		t.Fatal("未知字段应校验失败")
	}
	if !contains(err.Error(), "unknown_field") {
		t.Errorf("错误信息应提示未知字段: %v", err)
	}
}

func TestValidateJSONInvalidSchemaVersion(t *testing.T) {
	data := []byte(`{
		"schema_version": "persona.v2",
		"id": "test",
		"name": "Test",
		"description": "desc",
		"stance": {"intensity": 3},
		"speaking_style": {"verbosity": 3}
	}`)
	_, err := ValidateJSON(data)
	if err == nil {
		t.Fatal("错误 schema_version 应校验失败")
	}
	if !contains(err.Error(), "schema_version") {
		t.Error("错误信息应包含 schema_version")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
