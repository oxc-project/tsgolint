package gorule

import (
	"go/ast"
	"go/token"
)

// GoRuleListeners maps AST node types to listener functions for Go files
type GoRuleListeners map[string]func(node ast.Node)

// GoRule defines the interface for Go linting rules
type GoRule struct {
	Name string
	Run  func(ctx GoRuleContext, options any) GoRuleListeners
}

// GoRuleMessage represents a linting message for Go files
type GoRuleMessage struct {
	Id          string
	Description string
}

// GoRuleFix represents a code fix for Go files
type GoRuleFix struct {
	Text  string
	Range token.Pos
	End   token.Pos
}

// GoRuleSuggestion represents a suggestion for Go files
type GoRuleSuggestion struct {
	Message  GoRuleMessage
	FixesArr []GoRuleFix
}

func (s GoRuleSuggestion) Fixes() []GoRuleFix {
	return s.FixesArr
}

// GoRuleDiagnostic represents a diagnostic result for Go files
type GoRuleDiagnostic struct {
	Range       token.Pos
	End         token.Pos
	RuleName    string
	Message     GoRuleMessage
	FixesPtr    *[]GoRuleFix
	Suggestions *[]GoRuleSuggestion
	FileName    string
}

func (d GoRuleDiagnostic) Fixes() []GoRuleFix {
	if d.FixesPtr == nil {
		return []GoRuleFix{}
	}
	return *d.FixesPtr
}

func (d GoRuleDiagnostic) GetSuggestions() []GoRuleSuggestion {
	if d.Suggestions == nil {
		return []GoRuleSuggestion{}
	}
	return *d.Suggestions
}

// GoRuleContext provides context for Go rule execution
type GoRuleContext struct {
	File                       *ast.File
	FileSet                    *token.FileSet
	ReportRange                func(start, end token.Pos, msg GoRuleMessage)
	ReportRangeWithSuggestions func(start, end token.Pos, msg GoRuleMessage, suggestions ...GoRuleSuggestion)
	ReportNode                 func(node ast.Node, msg GoRuleMessage)
	ReportNodeWithFixes        func(node ast.Node, msg GoRuleMessage, fixes ...GoRuleFix)
	ReportNodeWithSuggestions  func(node ast.Node, msg GoRuleMessage, suggestions ...GoRuleSuggestion)
}