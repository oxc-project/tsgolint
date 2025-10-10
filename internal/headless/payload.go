package headless

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
)

// V2 (current) Headless payload format
// (We only support V2 here; V1 upgrade logic retained for parity.)

type headlessPayload struct {
	Version int              `json:"version"`
	Configs []headlessConfig `json:"configs"`
}

type headlessConfig struct {
	FilePaths []string       `json:"file_paths"`
	Rules     []headlessRule `json:"rules"`
}

type headlessRule struct { Name string `json:"name"` }

type headlessConfigForFileV1 struct {
	FilePath string   `json:"file_path"`
	Rules    []string `json:"rules"`
}

type headlessPayloadV1 struct { Files []headlessConfigForFileV1 `json:"files"` }

func deserializePayload(data []byte) (*headlessPayload, error) {
	version, err := getPayloadVersion(data)
	if err != nil { return nil, err }
	if version == 2 {
		var payload headlessPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, errors.New("failed to deserialize V2 payload: " + err.Error())
		}
		return &payload, nil
	}
	if version != 0 { return nil, fmt.Errorf("unsupported version `%d`: expected `unset` or `2`", version) }
	var v1 headlessPayloadV1
	if err := json.Unmarshal(data, &v1); err != nil { return nil, errors.New("failed to deserialize V1 payload: " + err.Error()) }
	if len(v1.Files) == 0 { return nil, errors.New("V1 payload has no files") }
	converted := &headlessPayload{Version: 2, Configs: make([]headlessConfig, len(v1.Files))}
	for i, f := range v1.Files {
		cfg := headlessConfig{FilePaths: []string{f.FilePath}, Rules: make([]headlessRule, len(f.Rules))}
		for j, r := range f.Rules { cfg.Rules[j] = headlessRule{Name: r} }
		converted.Configs[i] = cfg
	}
	return converted, nil
}

func getPayloadVersion(data []byte) (int, error) {
	var v struct{ Version int `json:"version"` }
	if err := json.Unmarshal(data, &v); err != nil { return 0, err }
	return v.Version, nil
}
