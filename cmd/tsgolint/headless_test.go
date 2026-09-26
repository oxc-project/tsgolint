package main

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf8"

	"github.com/go-json-experiment/json"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

func TestHeadlessRuleMessageEncodesLoneSurrogatesInDiagnosticFrame(t *testing.T) {
	highSurrogate := string([]byte{0xED, 0xA0, 0x80})                      // U+D800 in WTF-8
	lowSurrogate := string([]byte{0xED, 0xB0, 0x80})                       // U+DC00 in WTF-8
	pairedSurrogates := string([]byte{0xED, 0xA0, 0xBD, 0xED, 0xB8, 0x80}) // U+1F600 in WTF-8

	message := headlessRuleMessageFromRuleMessage(rule.RuleMessage{
		Id:          "nameMustMatch",
		Description: "Name café 😀 " + highSurrogate + " / " + lowSurrogate + " / " + pairedSurrogates,
		Help:        "review " + highSurrogate,
	})
	if !utf8.ValidString(message.Description) || !utf8.ValidString(message.Help) {
		t.Fatal("headless rule message must contain valid UTF-8")
	}

	diagnostic := headlessDiagnostic{
		Kind:    headlessDiagnosticKindRule,
		Message: message,
	}
	var frame bytes.Buffer
	if err := writeMessage(&frame, headlessMessageTypeDiagnostic, diagnostic); err != nil {
		t.Fatalf("writeMessage() error = %v", err)
	}

	data := frame.Bytes()
	if len(data) < 5 {
		t.Fatalf("frame has %d bytes, want at least the 5-byte header", len(data))
	}
	payloadLength := binary.LittleEndian.Uint32(data[:4])
	if int(payloadLength) != len(data)-5 {
		t.Fatalf("header payload length = %d, actual payload length = %d", payloadLength, len(data)-5)
	}
	if got := headlessMessageType(data[4]); got != headlessMessageTypeDiagnostic {
		t.Fatalf("message type = %d, want diagnostic (%d)", got, headlessMessageTypeDiagnostic)
	}

	var decoded headlessDiagnostic
	if err := json.Unmarshal(data[5:], &decoded); err != nil {
		t.Fatalf("decode diagnostic payload: %v", err)
	}
	if decoded.Message.Description != "Name café 😀 \\uD800 / \\uDC00 / 😀" {
		t.Errorf("description = %q", decoded.Message.Description)
	}
	if decoded.Message.Help != "review \\uD800" {
		t.Errorf("help = %q", decoded.Message.Help)
	}
}
