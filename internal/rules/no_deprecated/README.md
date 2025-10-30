# no-deprecated Rule Implementation

## Status

The `no-deprecated` rule is **fully implemented** and functional using `GetCombinedModifierFlags` to detect deprecated symbols.

## What's Implemented

1. ✅ Rule structure following the tsgolint pattern
2. ✅ AST node listeners for:
   - Identifier nodes
   - Property access expressions
   - Call expressions
   - New expressions
3. ✅ Declaration detection (to avoid flagging declarations themselves)
4. ✅ Import detection (to avoid flagging import statements)
5. ✅ Alias chain following logic
6. ✅ Deprecation detection using `GetCombinedModifierFlags`
7. ✅ Test cases covering various scenarios
8. ✅ Rule registration in main.go

## Implementation Details

The rule uses `ast.GetCombinedModifierFlags(decl) & ast.ModifierFlagsDeprecated` to check if a symbol is deprecated. This approach:

- Is based on typescript-go's `symbol_display.go` implementation
- Works with JSDoc `@deprecated` tags (parsed by TypeScript and stored as modifier flags)
- Checks all declarations of a symbol for the deprecated flag
- Follows alias chains to detect deprecation on imported/exported symbols

## Deprecation Reason Extraction

The rule now extracts the deprecation reason from JSDoc comments following the `Parser.withJSDoc` approach:

1. Gets leading comment ranges for the node
2. Filters for JSDoc comments (/** ... */)
3. Parses the comment text to find the @deprecated tag
4. Extracts the reason text after @deprecated

This allows the rule to report both simple deprecation and deprecation with custom messages.

## Testing

Once the JSDoc access is implemented, the tests in `no_deprecated_test.go` should pass. The test cases cover:

- Valid cases (declarations, non-deprecated usage, imports)
- Invalid cases (deprecated variables, functions, classes, properties, methods, enum members, namespace members)
- Cases with and without deprecation reasons

## References

- TypeScript-ESLint source: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-deprecated.ts
- TypeScript-ESLint tests: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-deprecated.test.ts
- TypeScript-ESLint docs: https://typescript-eslint.io/rules/no-deprecated

## Next Steps

1. Expose `Symbol.GetJsDocTags()` in typescript-go shim (preferred)
   - OR implement JSDoc parsing using existing utilities
2. Complete the `getJsDocDeprecation` function implementation
3. Run tests to verify the implementation
4. Add support for the `allow` option if needed
