// Package persona 提供从文件系统加载 Persona JSON 资产的能力。
package persona

import (
	"fmt"
	"os"
	"path/filepath"
)

// Loader 负责从指定目录加载 Persona JSON 文件。
type Loader struct {
	dir string
}

// NewLoader 创建一个从 dir 目录加载 Persona 的 Loader。
func NewLoader(dir string) *Loader {
	return &Loader{dir: dir}
}

// LoadAll 加载目录下所有 .json 文件并解析为 Persona。
// 返回以 persona ID 为键的映射表。
func (l *Loader) LoadAll() (map[string]Persona, error) {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return nil, fmt.Errorf("读取 persona 目录失败: %w", err)
	}

	result := make(map[string]Persona)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(l.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %w", path, err)
		}

		p, err := ValidateJSON(data)
		if err != nil {
			return nil, fmt.Errorf("校验 %s 失败: %w", path, err)
		}

		result[p.ID] = *p
	}

	return result, nil
}

// LoadOne 加载指定 ID 的 Persona。
func (l *Loader) LoadOne(id string) (Persona, error) {
	path := filepath.Join(l.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Persona{}, fmt.Errorf("读取 %s 失败: %w", path, err)
	}

	p, err := ValidateJSON(data)
	if err != nil {
		return Persona{}, fmt.Errorf("校验 %s 失败: %w", path, err)
	}

	if p.ID != id {
		return Persona{}, fmt.Errorf("文件内 ID %q 与请求 ID %q 不匹配", p.ID, id)
	}

	return *p, nil
}
