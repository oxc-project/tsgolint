package only_throw_error

import (
	"fmt"
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
	"github.com/typescript-eslint/tsgolint/internal/utils"
)

func TestOnlyThrowErrorPackageSpecifiers(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		imported string
		declared string
		allowed  string
		matches  bool
	}{
		{"typescript", "typescript", "typescript", true},
		{"typescript", "typescript", "script", false},
		{"semver", "semver", "semver", true},
		{"semver", "semver", "semve", false},
		{"semver", "semver", "sem", false},
		{"semver", "semver", "s.mver", false},
		{"s.mver", "s.mver", "s.mver", true},
		{"semver-compare", "semver-compare", "semver", false},
		{"my-semver", "my-semver", "semver", false},
		{"@angular/core", "@angular/core", "@angular", true},
		{"@angular/core", "@angular/core", "@angular/core", true},
		{"@angularlol/core", "@angularlol/core", "@angular", false},
		{"@angular/core-extra", "@angular/core-extra", "@angular/core", false},
		{"semver", "@types/semver", "semver", true},
		{"semver", "@types/semver", "@types/semver", true},
		{"semver", "@types/semver", "semve", false},
		{"semver", "@types/semver", "sem", false},
		{"semver", "@types/semver", "s.mver", false},
		{"@babel/code-frame", "@types/babel__code-frame", "@babel/code-frame", true},
		{"@babel/code-frame", "@types/babel__code-frame", "@types/babel__code-frame", true},
		{"@babel/code-frame-extra", "@types/babel__code-frame-extra", "@babel/code-frame", false},
	} {
		t.Run(tc.declared+"/allow="+tc.allowed, func(t *testing.T) {
			t.Parallel()
			code := fmt.Sprintf("import type { ErrorLike } from %q; declare const error: ErrorLike; throw error;", tc.imported)
			// A non-Error type ensures the allow specifier is the only reason this passes.
			files := map[string]string{
				"node_modules/" + tc.declared + "/package.json":   fmt.Sprintf(`{"name": %q, "version": "1.0.0", "types": "lib/index.d.ts"}`, tc.declared),
				"node_modules/" + tc.declared + "/lib/index.d.ts": `export interface ErrorLike { message: string }`,
			}
			options := OnlyThrowErrorOptions{Allow: []utils.TypeOrValueSpecifier{{From: utils.TypeOrValueSpecifierFromPackage, Name: []string{"ErrorLike"}, Package: tc.allowed}}}
			var valid []rule_tester.ValidTestCase
			var invalid []rule_tester.InvalidTestCase
			if tc.matches {
				valid = []rule_tester.ValidTestCase{{Code: code, Files: files, Options: options}}
			} else {
				invalid = []rule_tester.InvalidTestCase{{Code: code, Files: files, Options: options, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "object"}}}}
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &OnlyThrowErrorRule, valid, invalid)
		})
	}
}
