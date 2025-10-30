# no-deprecated Rule Implementation

## Status

The `no-deprecated` rule has been partially implemented. The core structure and logic are in place, but the implementation is currently incomplete due to missing functionality in the typescript-go shim.

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
6. ✅ Test cases covering various scenarios
7. ✅ Rule registration in main.go

## What's Missing

The implementation requires the `GetJsDocTags()` method to be exposed in the typescript-go shim. This method:

- Exists in the TypeScript compiler API as `symbol.getJsDocTags(checker)`
- Returns an array of `JSDocTagInfo` objects
- Is used to access JSDoc comments like `@deprecated` from symbols

### Required typescript-go Changes

To complete this implementation, the following needs to be added to the typescript-go shim:

1. **Expose `Symbol.GetJsDocTags()` method**:
   ```go
   // In shim/ast or shim/checker
   func Symbol_GetJsDocTags(symbol *ast.Symbol, checker *checker.Checker) []JSDocTagInfo
   ```

2. **Define `JSDocTagInfo` type**:
   ```go
   type JSDocTagInfo struct {
       Name string
       Text []SymbolDisplayPart  // or similar
   }
   ```

3. **Alternatively**, implement JSDoc parsing using existing utilities:
   - Use `parser.GetJSDocCommentRanges()` to get comment ranges
   - Parse the comment text to extract JSDoc tags
   - This would be more complex but doesn't require typescript-go changes

## Alternative Implementation

If modifying typescript-go is not feasible, an alternative approach would be:

1. Get the symbol's declarations
2. For each declaration, use `parser.GetJSDocCommentRanges()` to get JSDoc comments
3. Parse the comment text to extract `@deprecated` tags
4. Extract the deprecation reason from the tag text

This approach would be more complex and potentially less reliable than using the TypeScript compiler API directly.

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
