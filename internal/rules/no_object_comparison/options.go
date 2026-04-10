package no_object_comparison

import "github.com/go-json-experiment/json"

type NoObjectComparisonClassOption struct {
	Name                    string `json:"name,omitempty"`
	ForbidEqualityOperators bool   `json:"forbidEqualityOperators,omitempty"`
}

type NoObjectComparisonOptions struct {
	Classes []NoObjectComparisonClassOption `json:"classes,omitempty"`
}

func (j *NoObjectComparisonOptions) UnmarshalJSON(value []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(value, &raw); err != nil {
		return err
	}

	type Plain NoObjectComparisonOptions
	var plain Plain
	if err := json.Unmarshal(value, &plain); err != nil {
		return err
	}

	if v, ok := raw["classes"]; !ok || v == nil {
		plain.Classes = []NoObjectComparisonClassOption{}
	}

	*j = NoObjectComparisonOptions(plain)
	return nil
}
