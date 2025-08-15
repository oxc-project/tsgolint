package rule

import (
	"testing"
)

func TestRuleMessageHelp(t *testing.T) {
	// Test RuleMessage with Help field
	messageWithHelp := RuleMessage{
		Id:          "testId",
		Description: "Test description",
		Help:        "This is help text",
	}

	if messageWithHelp.Id != "testId" {
		t.Errorf("Expected Id to be 'testId', got %s", messageWithHelp.Id)
	}
	if messageWithHelp.Description != "Test description" {
		t.Errorf("Expected Description to be 'Test description', got %s", messageWithHelp.Description)
	}
	if messageWithHelp.Help != "This is help text" {
		t.Errorf("Expected Help to be 'This is help text', got %s", messageWithHelp.Help)
	}

	// Test RuleMessage without Help field (should default to empty string)
	messageWithoutHelp := RuleMessage{
		Id:          "testId2",
		Description: "Test description 2",
	}

	if messageWithoutHelp.Id != "testId2" {
		t.Errorf("Expected Id to be 'testId2', got %s", messageWithoutHelp.Id)
	}
	if messageWithoutHelp.Description != "Test description 2" {
		t.Errorf("Expected Description to be 'Test description 2', got %s", messageWithoutHelp.Description)
	}
	if messageWithoutHelp.Help != "" {
		t.Errorf("Expected Help to be empty string when not set, got %s", messageWithoutHelp.Help)
	}

	// Test that existing code usage patterns continue to work (simulating existing rules)
	buildTestMessage := func() RuleMessage {
		return RuleMessage{
			Id:          "existingRule",
			Description: "Existing rule description",
			// Help field is intentionally omitted to test backward compatibility
		}
	}

	existingMessage := buildTestMessage()
	if existingMessage.Help != "" {
		t.Errorf("Expected Help to be empty for existing style usage, got %s", existingMessage.Help)
	}

	// Test that new style usage works
	buildNewMessage := func() RuleMessage {
		return RuleMessage{
			Id:          "newRule",
			Description: "New rule description",
			Help:        "New rule help text",
		}
	}

	newMessage := buildNewMessage()
	if newMessage.Help != "New rule help text" {
		t.Errorf("Expected Help to be 'New rule help text', got %s", newMessage.Help)
	}
}