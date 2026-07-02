package main

import (
	"encoding/binary"
	"io"

	"github.com/go-json-experiment/json"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/diagnostic"
	"github.com/typescript-eslint/tsgolint/internal/linter"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

type headlessRange struct {
	Pos int `json:"pos"`
	End int `json:"end"`
}

func headlessRangeFromRange(r core.TextRange) *headlessRange {
	if !r.IsValid() {
		return nil
	}
	return &headlessRange{
		Pos: r.Pos(),
		End: r.End(),
	}
}

// headlessRuleMessage mirrors rule.RuleMessage without exposing the internal rule package shape over the wire.
type headlessRuleMessage struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Help        string `json:"help,omitempty"`
}

func headlessRuleMessageFromRuleMessage(msg rule.RuleMessage) headlessRuleMessage {
	return headlessRuleMessage{
		Id:          msg.Id,
		Description: msg.Description,
		Help:        msg.Help,
	}
}

type headlessFix struct {
	Text  string        `json:"text"`
	Range headlessRange `json:"range"`
}

type headlessSuggestion struct {
	Message headlessRuleMessage `json:"message"`
	Fixes   []headlessFix       `json:"fixes"`
}

func headlessFixesFromRuleFixes(fixes []rule.RuleFix) []headlessFix {
	headlessFixes := make([]headlessFix, len(fixes))
	for i, fix := range fixes {
		headlessFixes[i] = headlessFix{
			Text:  fix.Text,
			Range: *headlessRangeFromRange(fix.Range),
		}
	}
	return headlessFixes
}

// headlessDiagnosticKind identifies which diagnostic-specific fields are present in a headlessDiagnostic.
type headlessDiagnosticKind uint8

const (
	headlessDiagnosticKindRule headlessDiagnosticKind = iota
	headlessDiagnosticKindTsconfig
)

// headlessLabeledRange highlights supporting source spans for a diagnostic.
type headlessLabeledRange struct {
	Label string        `json:"label"`
	Range headlessRange `json:"range"`
}

// headlessDiagnostic is the JSON payload for rule and internal diagnostics emitted by headless mode.
type headlessDiagnostic struct {
	Kind          headlessDiagnosticKind `json:"kind"`
	Range         *headlessRange         `json:"range,omitempty"`
	Message       headlessRuleMessage    `json:"message"`
	FilePath      *string                `json:"file_path"`
	LabeledRanges []headlessLabeledRange `json:"labeled_ranges,omitempty"`

	// Only for kind="rule"
	Rule        *string              `json:"rule,omitempty"`
	Fixes       []headlessFix        `json:"fixes,omitempty"`
	Suggestions []headlessSuggestion `json:"suggestions,omitempty"`
}

func headlessDiagnosticFromRuleDiagnostic(
	rd *rule.RuleDiagnostic,
	includeFixes bool,
	includeSuggestions bool,
) headlessDiagnostic {
	filePath := rd.SourceFile.FileName()
	hd := headlessDiagnostic{
		Kind:          headlessDiagnosticKindRule,
		Range:         headlessRangeFromRange(rd.Range),
		Rule:          &rd.RuleName,
		Message:       headlessRuleMessageFromRuleMessage(rd.Message),
		FilePath:      &filePath,
		LabeledRanges: nil,
	}

	if len(rd.LabeledRanges) > 0 {
		hd.LabeledRanges = make([]headlessLabeledRange, len(rd.LabeledRanges))
		for i, labeledRange := range rd.LabeledRanges {
			hd.LabeledRanges[i] = headlessLabeledRange{
				Label: labeledRange.Label,
				Range: *headlessRangeFromRange(labeledRange.Range),
			}
		}
	}

	if includeFixes {
		hd.Fixes = headlessFixesFromRuleFixes(rd.Fixes())
	}
	if includeSuggestions {
		suggestions := rd.GetSuggestions()
		hd.Suggestions = make([]headlessSuggestion, len(suggestions))
		for i, suggestion := range suggestions {
			hd.Suggestions[i] = headlessSuggestion{
				Message: headlessRuleMessageFromRuleMessage(suggestion.Message),
				Fixes:   headlessFixesFromRuleFixes(suggestion.Fixes()),
			}
		}
	}

	return hd
}

func headlessDiagnosticFromInternalDiagnostic(internalDiagnostic *diagnostic.Internal) headlessDiagnostic {
	return headlessDiagnostic{
		Kind:  headlessDiagnosticKindTsconfig,
		Range: headlessRangeFromRange(internalDiagnostic.Range),
		Message: headlessRuleMessage{
			Id:          internalDiagnostic.Id,
			Description: internalDiagnostic.Description,
			Help:        internalDiagnostic.Help,
		},
		FilePath: internalDiagnostic.FilePath,
	}
}

// headlessMessageType is stored in the binary frame header before each JSON payload.
type headlessMessageType uint8

const (
	headlessMessageTypeError headlessMessageType = iota
	headlessMessageTypeDiagnostic
	headlessMessageTypeTiming
)

type headlessMessagePayloadError struct {
	Error string `json:"error"`
}

type headlessTimingPayload struct {
	Rules []headlessRuleTiming `json:"rules"`
}

type headlessRuleTiming struct {
	RuleName string `json:"rule_name"`
	// Duration is serialized as nanoseconds.
	Duration uint64 `json:"duration"`
	Calls    uint64 `json:"calls"`
}

func headlessTimingPayloadFromRecords(records []linter.RuleTimingRecord) headlessTimingPayload {
	rules := make([]headlessRuleTiming, len(records))
	for i, record := range records {
		rules[i] = headlessRuleTiming{
			RuleName: record.RuleName,
			Duration: uint64(record.Duration),
			Calls:    record.Calls,
		}
	}
	return headlessTimingPayload{Rules: rules}
}

func writeMessage(w io.Writer, messageType headlessMessageType, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var header [5]byte
	// The headless protocol prefixes each JSON payload with a 4-byte little-endian length and 1-byte message type.
	binary.LittleEndian.PutUint32(header[:], uint32(len(payloadBytes)))
	header[4] = byte(messageType)
	w.Write(header[:])
	w.Write(payloadBytes)
	return nil
}
