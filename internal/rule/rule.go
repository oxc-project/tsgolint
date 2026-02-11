package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

const (
	lastTokenKind                        ast.Kind = 1000
	lastOnExitTokenKind                  ast.Kind = 2000
	lastOnAllowPatternTokenKind          ast.Kind = 3000
	lastOnAllowPatternOnExitTokenKind    ast.Kind = 4000
	lastOnNotAllowPatternTokenKind       ast.Kind = 5000
	lastOnNotAllowPatternOnExitTokenKind ast.Kind = 6000
)

func ListenerOnExit(kind ast.Kind) ast.Kind {
	return kind + 1000
}

// TODO(port): better name
func ListenerOnAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnExitTokenKind
}
func ListenerOnNotAllowPattern(kind ast.Kind) ast.Kind {
	return kind + lastOnAllowPatternOnExitTokenKind
}

type RuleListeners map[ast.Kind](func(node *ast.Node))

type Rule struct {
	Name string
	Run  func(ctx RuleContext, options any) RuleListeners
}

type RuleMessage struct {
	Id          string
	Description string
	Help        string
}

type RuleFix struct {
	Text  string
	Range core.TextRange
}

type RuleLabeledRange struct {
	Label string         `json:"label"`
	Range core.TextRange `json:"range"`
}

func RuleFixInsertBefore(file *ast.SourceFile, node *ast.Node, text string) RuleFix {
	trimmed := utils.TrimNodeTextRange(file, node)
	return RuleFix{
		Text:  text,
		Range: trimmed.WithEnd(trimmed.Pos()),
	}
}
func RuleFixInsertAfter(node *ast.Node, text string) RuleFix {
	return RuleFix{
		Text:  text,
		Range: node.Loc.WithPos(node.End()),
	}
}
func RuleFixReplace(file *ast.SourceFile, node *ast.Node, text string) RuleFix {
	return RuleFixReplaceRange(utils.TrimNodeTextRange(file, node), text)
}
func RuleFixReplaceRange(textRange core.TextRange, text string) RuleFix {
	return RuleFix{
		Text:  text,
		Range: textRange,
	}
}
func RuleFixRemove(file *ast.SourceFile, node *ast.Node) RuleFix {
	return RuleFixReplace(file, node, "")
}
func RuleFixRemoveRange(textRange core.TextRange) RuleFix {
	return RuleFixReplaceRange(textRange, "")
}

type RuleSuggestion struct {
	Message  RuleMessage
	FixesArr []RuleFix
}

func (s RuleSuggestion) Fixes() []RuleFix {
	return s.FixesArr
}

type RuleDiagnostic struct {
	Range    core.TextRange
	RuleName string
	Message  RuleMessage
	// nil if no fixes were provided
	FixesPtr *[]RuleFix
	// nil if no suggestions were provided
	Suggestions   *[]RuleSuggestion
	SourceFile    *ast.SourceFile
	LabeledRanges []RuleLabeledRange
}

func (d RuleDiagnostic) Fixes() []RuleFix {
	if d.FixesPtr == nil {
		return []RuleFix{}
	}
	return *d.FixesPtr
}
func (d RuleDiagnostic) GetSuggestions() []RuleSuggestion {
	if d.Suggestions == nil {
		return []RuleSuggestion{}
	}
	return *d.Suggestions
}

type RuleContext struct {
	SourceFile                 *ast.SourceFile
	Program                    *compiler.Program
	TypeChecker                *checker.Checker
	ReportDiagnostic           func(diagnostic RuleDiagnostic)
	ReportRange                func(textRange core.TextRange, msg RuleMessage)
	ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestionsFn func() []RuleSuggestion)
	ReportNode                 func(node *ast.Node, msg RuleMessage)
	ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixesFn func() []RuleFix)
	ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestionsFn func() []RuleSuggestion)
}

func ReportNodeWithFixesOrSuggestions(ctx RuleContext, node *ast.Node, fix bool, msg RuleMessage, suggestionMsg RuleMessage, fixes ...RuleFix) {
	if fix {
		ctx.ReportNodeWithFixes(node, msg, func() []RuleFix { return fixes })
	} else {
		ctx.ReportNodeWithSuggestions(node, msg, func() []RuleSuggestion {
			return []RuleSuggestion{{
				Message:  suggestionMsg,
				FixesArr: fixes,
			}}
		})
	}
}
