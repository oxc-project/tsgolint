package naming_convention

import (
	"bytes"
	"errors"

	"github.com/go-json-experiment/json"
)

// The upstream schema allows a selector string or array of selector strings.
type NamingSelector []string

func (s *NamingSelector) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty selector")
	}
	if data[0] == '"' {
		var single string
		if err := json.Unmarshal(data, &single); err != nil {
			return err
		}
		*s = NamingSelector{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*s = many
	return nil
}

func (s NamingSelector) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

// The upstream schema allows either a regex string (implicitly match=true)
// or an object with a regex and explicit match boolean.
type NamingFilter struct {
	Regex string
	Match bool
}

func (f *NamingFilter) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return errors.New("empty filter")
	}
	if data[0] == '"' {
		f.Match = true
		return json.Unmarshal(data, &f.Regex)
	}
	var raw struct {
		Regex string `json:"regex"`
		Match bool   `json:"match"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Regex, f.Match = raw.Regex, raw.Match
	return nil
}
func (f NamingFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Regex string `json:"regex"`
		Match bool   `json:"match"`
	}{f.Regex, f.Match})
}
