package main

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

// V2 (current) Headless payload format:
//
//	{
//	  "version": 2,
//	  "configs": [
//	    {
//	      "file_paths": ["/abs/path/a.ts"],
//	      "rules": [{ "name": "no-floating-promises", "options": {} }],
//	      "type_check": true
//	    }
//	  ],
//	  "source_overrides": { "/abs/path/a.ts": "source text" },
//	  "report_syntactic": true,
//	  "report_semantic": true
//	}
//
// `report_syntactic` and `report_semantic` select which kinds of TypeScript
// diagnostics are reported, and the config groups' `type_check` which files
// report them.
//
// `file_paths` entries and `source_overrides` keys should be absolute paths; a
// relative one is resolved against the working directory tsgolint runs in.
type headlessPayload struct {
	Version         int               `json:"version"` // version must be 2
	Configs         []headlessConfig  `json:"configs"`
	SourceOverrides map[string]string `json:"source_overrides,omitempty"`
	ReportSyntactic bool              `json:"report_syntactic,omitempty"`
	ReportSemantic  bool              `json:"report_semantic,omitempty"`
}

type headlessConfig struct {
	FilePaths []string       `json:"file_paths"`
	Rules     []headlessRule `json:"rules"`
	// TypeCheck selects the files reporting the TypeScript diagnostics.
	//
	// Semantics:
	//
	//  1. A payload where no config group carries the field keeps the previous
	//     behavior: every linted file reports the diagnostics.
	//  2. The field is resolved per file: the last config group listing a file
	//     defines how it's linted, its rules and its `type_check` together. A
	//     config group listing a file without the field therefore reports the
	//     diagnostics for it, regardless of the `type_check` set by an earlier
	//     config group.
	//  3. `false` excludes every TypeScript diagnostic attached to the file:
	//     the syntactic, semantic and include processor ones alike, matching
	//     whether tsc's own check covers the file. The program-level tsconfig
	//     diagnostics, which belong to no file in particular, are reported
	//     either way, and the type-aware lint rules still run on the file.
	//  4. `report_syntactic` and `report_semantic` remain the run-wide
	//     selectors of which kinds of diagnostics are reported. `type_check`
	//     never selects a kind, so `type_check: true` without them is a no-op.
	//  5. A config group listing no file has no effect.
	TypeCheck *bool `json:"type_check,omitempty"`
}

type headlessRule struct {
	Name    string `json:"name"`
	Options any    `json:"options,omitempty"`
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
