// Package persona 提供 Persona Schema v1 的校验能力。
package persona

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Validate 对 Persona 进行完整校验，返回所有错误。
func (p Persona) Validate() error {
	var errs []string

	if p.SchemaVersion != "persona.v1" {
		errs = append(errs, fmt.Sprintf("schema_version 必须是 'persona.v1'，得到 %q", p.SchemaVersion))
	}
	if p.ID == "" {
		errs = append(errs, "id 不能为空")
	}
	if p.Name == "" {
		errs = append(errs, "name 不能为空")
	}
	if p.Description == "" {
		errs = append(errs, "description 不能为空")
	}
	if p.Stance.Intensity < 1 || p.Stance.Intensity > 5 {
		errs = append(errs, fmt.Sprintf("stance.intensity 必须在 1-5 之间，得到 %d", p.Stance.Intensity))
	}
	for i, cb := range p.CoreBeliefs {
		if cb.Priority < 1 || cb.Priority > 5 {
			errs = append(errs, fmt.Sprintf("core_beliefs[%d].priority 必须在 1-5 之间，得到 %d", i, cb.Priority))
		}
	}
	if p.SpeakingStyle.Verbosity < 1 || p.SpeakingStyle.Verbosity > 5 {
		errs = append(errs, fmt.Sprintf("speaking_style.verbosity 必须在 1-5 之间，得到 %d", p.SpeakingStyle.Verbosity))
	}

	if len(errs) > 0 {
		return fmt.Errorf("persona %q 校验失败:\n- %s", p.ID, strings.Join(errs, "\n- "))
	}
	return nil
}

// ValidateJSON 解析并校验 JSON 数据，拒绝未知字段。
func ValidateJSON(data []byte) (*Persona, error) {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	var p Persona
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return &p, nil
}
