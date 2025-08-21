package main

type headlessPayload struct {
	Version string           `json:"version,omitempty"`
	Configs []headlessConfig `json:"configs"`
}

type headlessConfig struct {
	FilePaths []string       `json:"file_paths"`
	Rules     []headlessRule `json:"rules"`
}

type headlessRule struct {
	Name string `json:"name"`
}
