# no-deprecated Rule Status

## Current Implementation

The rule is structurally complete with:
- AST listeners for Identifier, PropertyAccessExpression, CallExpression, and NewExpression
- Logic to skip declarations and imports
- Symbol deprecation checking using `GetCombinedModifierFlags` 
- JSDoc comment parsing to extract deprecation reasons
- Fallback JSDoc parsing when modifier flags don't work
- Comprehensive test suite ported from typescript-eslint (100+ test cases)

## Known Issue

**All invalid tests are currently failing** - the rule is not detecting deprecated usage.

### Root Cause

The issue appears to be that `ast.GetCombinedModifierFlags(decl)` is not returning the `ModifierFlagsDeprecated` flag even when a declaration has a `@deprecated` JSDoc tag.

This could be because:
1. TypeScript's parser isn't setting the deprecated flag from JSDoc
2. The flag isn't being propagated correctly
3. There's a missing compiler option or configuration
4. The shim layer isn't exposing the flags correctly

### What's Been Tried

1. ✅ Using `GetCombinedModifierFlags` (per typescript-go's symbol_display.go)
2. ✅ Parsing JSDoc comments directly as a fallback
3. ✅ Checking all symbol declarations
4. ✅ Following alias chains
5. ❌ Using `Checker.IsDeprecatedDeclaration()` - not exposed in shim

### Potential Solutions

1. **Expose `Checker.IsDeprecatedDeclaration()` in the shim** - This is the most reliable method used by typescript-go internally
   - Location: `typescript-go/internal/checker/checker.go`
   - Would need to add to `shim/checker/extra-shim.json` and run `just shim`

2. **Use JSDoc tags directly** - Symbol.GetJsDocTags() would work but isn't exposed
   - Would need shim updates

3. **Parse JSDoc more aggressively** - Current fallback might have bugs

4. **Check if compiler options are needed** - Maybe JSDoc parsing needs to be enabled

### Next Steps

The most promising solution is to expose `Checker.IsDeprecatedDeclaration()` in the shim, as this is what typescript-go uses internally and should be reliable.
