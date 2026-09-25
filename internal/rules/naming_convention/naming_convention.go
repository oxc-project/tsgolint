package naming_convention

import "github.com/typescript-eslint/tsgolint/internal/rule"

// NamingConventionRule is intentionally a stub for the upstream test-port commit.
var NamingConventionRule = rule.Rule{
	Name: "naming-convention",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return nil
	},
}
