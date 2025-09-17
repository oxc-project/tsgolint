package tsgolint

import (
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
)

// V1 Headless payload format
type headlessConfigForFileV1 struct {
	FilePath string   `json:"file_path"`
	Rules    []string `json:"rules"`
}
type headlessPayloadV1 struct {
	Files []headlessConfigForFileV1 `json:"files"`
}

// V2 (current) Headless payload format
type headlessPayload struct {
	Version int              `json:"version"` // version must be 2
	Configs []headlessConfig `json:"configs"`
}

type headlessConfig struct {
	FilePaths []string       `json:"file_paths"`
	Rules     []headlessRule `json:"rules"`
}

type headlessRule struct {
	Name string `json:"name"`
}

func deserializePayload(data []byte) (*headlessPayload, error) {
	version, err := getPayloadVersion(data)
	if err != nil {
		return nil, err
	}

	if version == 2 {
		var payload headlessPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, errors.New("failed to deserialize V2 payload: " + err.Error())
		}
		return &payload, nil
	}

	// Version 0 or unset indicates V1 payload
	if version != 0 {
		return nil, fmt.Errorf("unsupported version `%d`: expected `unset` or `2`", version)
	}

	var payloadV1 headlessPayloadV1
	if err := json.Unmarshal(data, &payloadV1); err != nil {
		return nil, errors.New("failed to deserialize V1 payload: " + err.Error())
	}

	// Validate V1 payload
	if len(payloadV1.Files) == 0 {
		return nil, errors.New("V1 payload has no files")
	}

	// Convert V1 to V2
	payloadV2 := &headlessPayload{
		Version: 2,
		Configs: make([]headlessConfig, len(payloadV1.Files)),
	}
	for i, fileV1 := range payloadV1.Files {
		config := headlessConfig{
			FilePaths: []string{fileV1.FilePath}, // V1 has single file, V2 supports multiple
			Rules:     make([]headlessRule, len(fileV1.Rules)),
		}
		for j, rule := range fileV1.Rules {
			config.Rules[j] = headlessRule{Name: rule} // V1 rules are just strings
		}
		payloadV2.Configs[i] = config
	}

	return payloadV2, nil
}

func getPayloadVersion(data []byte) (int, error) {
	var versionCheck struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(data, &versionCheck); err != nil {
		return 0, err
	}
	return versionCheck.Version, nil
}
