#!/usr/bin/env node
import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import { Expression, ParseResult, parseSync, Program } from 'oxc-parser';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

interface TestCase {
  code: string;
  errors?: {
    messageId?: string;
    message?: string;
    line?: number;
    column?: number;
    endLine?: number;
    endColumn?: number;
    suggestions?: {
      messageId?: string;
      message?: string;
      output: string;
    }[];
  }[];
  output?: string;
  options?: any[];
}

interface RuleTester {
  valid: TestCase[];
  invalid: TestCase[];
}

function kebabToPascal(str: string): string {
  return str
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join('');
}

function camelToPascal(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function extractOptionsFromAST(node: any): any {
  if (!node) return null;

  if (node.type === 'ObjectExpression') {
    const result: any = {};
    for (const prop of node.properties || []) {
      if (prop.type === 'ObjectProperty' || prop.type === 'Property') {
        const key = prop.key?.name || prop.key?.value;
        if (key) {
          result[key] = extractOptionsFromAST(prop.value);
        }
      }
    }
    return result;
  } else if (node.type === 'ArrayExpression') {
    return (node.elements || []).map((el: any) => extractOptionsFromAST(el));
  } else if (node.type === 'StringLiteral' || node.type === 'Literal') {
    if (typeof node.value === 'string') {
      return node.value;
    }
    return node.value;
  } else if (node.type === 'NumericLiteral') {
    return node.value;
  } else if (node.type === 'BooleanLiteral') {
    return node.value;
  } else if (node.type === 'NullLiteral') {
    return null;
  } else if (node.type === 'Identifier') {
    // Handle special identifiers like true, false, null, undefined
    if (node.name === 'true') return true;
    if (node.name === 'false') return false;
    if (node.name === 'null') return null;
    if (node.name === 'undefined') return undefined;
    return node.name;
  }

  return null;
}

function convertOptionsToGoCode(options: any[], ruleName: string): string {
  if (!options || options.length === 0) {
    return '';
  }

  // For most rules, options is an array with a single object
  if (options.length === 1 && typeof options[0] === 'object' && !Array.isArray(options[0])) {
    const ruleNamePascal = kebabToPascal(ruleName);
    const optionsTypeName = `${ruleNamePascal}Options`;
    const opt = options[0];

    let goCode = `Options: ${optionsTypeName}{`;
    const fields: string[] = [];

    for (const [key, value] of Object.entries(opt)) {
      const fieldName = camelToPascal(key);

      if (Array.isArray(value)) {
        // Handle arrays (e.g., IgnoredTypeNames: []string{"RegExp"})
        if (value.every(v => typeof v === 'string')) {
          const values = value.map(v => `"${v}"`).join(', ');
          fields.push(`${fieldName}: []string{${values}}`);
        } else {
          // Handle other array types
          fields.push(`${fieldName}: /* TODO: handle array type */`);
        }
      } else if (typeof value === 'boolean') {
        fields.push(`${fieldName}: ${value}`);
      } else if (typeof value === 'string') {
        fields.push(`${fieldName}: "${value}"`);
      } else if (typeof value === 'number') {
        fields.push(`${fieldName}: ${value}`);
      } else {
        fields.push(`${fieldName}: /* TODO: handle complex type */`);
      }
    }

    goCode += fields.join(', ') + '}';
    return goCode;
  }

  // For other cases, we'll need manual handling
  return `Options: /* TODO: handle options: ${JSON.stringify(options)} */`;
}

function extractTestCases(parseResult: ParseResult): RuleTester {
  const valid: TestCase[] = [];
  const invalid: TestCase[] = [];

  // Find the RuleTester.run call
  function traverse(node: any): void {
    if (!node || typeof node !== 'object') return;

    if (
      node.type === 'CallExpression' &&
      (node.callee?.type === 'MemberExpression' || node.callee?.type === 'StaticMemberExpression') &&
      node.callee?.object?.name === 'ruleTester' &&
      node.callee?.property?.name === 'run'
    ) {
      // The third argument should be the test cases object
      const testCasesArg = node.arguments?.[2];
      if (testCasesArg?.type === 'ObjectExpression') {
        for (const prop of testCasesArg.properties || []) {
          if (prop.type === 'ObjectProperty' || prop.type === 'Property') {
            const key = prop.key?.name || prop.key?.value;
            if (key === 'valid' && prop.value?.type === 'ArrayExpression') {
              for (let i = 0; i < (prop.value.elements?.length || 0); i++) {
                const element = prop.value.elements[i];
                if (element) {
                  // Skip spread elements or other non-parseable nodes
                  if (element.type === 'SpreadElement') {
                    console.warn(`Skipping spread element in valid test case at index ${i}`);
                    continue;
                  }
                  try {
                    const testCase = parseTestCase(element);
                    if (!testCase.code) {
                      console.warn(`Skipping valid test case at index ${i}: missing code property`);
                      continue;
                    }
                    valid.push(testCase);
                  } catch (e) {
                    console.warn(
                      `Skipping valid test case at index ${i}: ${e instanceof Error ? e.message : String(e)}`,
                    );
                  }
                }
              }
            } else if (key === 'invalid' && prop.value?.type === 'ArrayExpression') {
              for (let i = 0; i < (prop.value.elements?.length || 0); i++) {
                const element = prop.value.elements[i];
                if (element) {
                  // Skip spread elements or other non-parseable nodes
                  if (element.type === 'SpreadElement') {
                    console.warn(`Skipping spread element in invalid test case at index ${i}`);
                    continue;
                  }
                  try {
                    const testCase = parseTestCase(element);
                    if (!testCase.code) {
                      console.warn(`Skipping invalid test case at index ${i}: missing code property`);
                      continue;
                    }
                    if (!testCase.errors || testCase.errors.length === 0) {
                      console.warn(`Warning: Invalid test case at index ${i} has no errors specified`);
                    }
                    invalid.push(testCase);
                  } catch (e) {
                    console.warn(
                      `Skipping invalid test case at index ${i}: ${e instanceof Error ? e.message : String(e)}`,
                    );
                  }
                }
              }
            }
          }
        }
      }
    }

    // Recursively traverse the AST
    for (const key in node) {
      if (key === 'parent') continue; // Skip parent references to avoid cycles
      const value = node[key];
      if (value && typeof value === 'object') {
        if (Array.isArray(value)) {
          for (const child of value) {
            if (child && typeof child === 'object') {
              traverse(child);
            }
          }
        } else {
          traverse(value);
        }
      }
    }
  }

  function parseTestCase(node: any): TestCase {
    const testCase: TestCase = { code: '' };

    if (!node) {
      throw new Error('Test case node is null or undefined');
    }

    // Handle string literals (different AST parsers use different names)
    if (node.type === 'StringLiteral' || node.type === 'Literal' || node.type === 'TemplateLiteral') {
      testCase.code = getStringValue(node);
      if (!testCase.code && testCase.code !== '') {
        throw new Error('Failed to extract code from string literal');
      }
    } else if (node.type === 'TaggedTemplateExpression') {
      // Handle tagged template expressions (e.g., noFormat`...`)
      // For now, try to extract the template literal part
      if (node.quasi?.type === 'TemplateLiteral') {
        testCase.code = getStringValue(node.quasi);
      } else {
        // Fall back to empty string or throw error
        console.warn(`Warning: Could not extract code from TaggedTemplateExpression`);
        testCase.code = '// Tagged template expression - could not extract';
      }
    } else if (node.type === 'ObjectExpression') {
      for (const prop of node.properties || []) {
        if (prop.type === 'ObjectProperty' || prop.type === 'Property') {
          const key = prop.key?.name || prop.key?.value;

          switch (key) {
            case 'code':
              testCase.code = getStringValue(prop.value);
              // Special handling for TaggedTemplateExpression that might have whitespace-only content
              if (prop.value?.type === 'TaggedTemplateExpression' && !testCase.code?.trim()) {
                // For tagged templates like noFormat`...`, even empty/whitespace code might be intentional
                // Don't throw error, but log a warning
                console.warn('Warning: Tagged template expression resulted in empty/whitespace-only code');
              } else if (!testCase.code) {
                throw new Error('Test case code property is empty');
              }
              break;
            case 'errors':
              if (prop.value?.type === 'ArrayExpression') {
                testCase.errors = [];
                for (const errorNode of prop.value.elements || []) {
                  if (errorNode) {
                    const error = parseError(errorNode);
                    testCase.errors.push(error);
                  }
                }
              }
              break;
            case 'output':
              testCase.output = getStringValue(prop.value);
              break;
            case 'options':
              if (prop.value?.type === 'ArrayExpression') {
                testCase.options = [];
                for (const optNode of prop.value.elements || []) {
                  if (optNode) {
                    const extractedOption = extractOptionsFromAST(optNode);
                    if (extractedOption !== null) {
                      testCase.options.push(extractedOption);
                    }
                  }
                }
              }
              break;
          }
        }
      }
    } else {
      throw new Error(`Unexpected test case node type: ${node.type}`);
    }

    return testCase;
  }

  function parseError(node: any): any {
    const error: any = {};

    if (!node) {
      throw new Error('Error node is null or undefined');
    }

    if (node.type === 'ObjectExpression') {
      for (const prop of node.properties || []) {
        if (prop.type === 'ObjectProperty' || prop.type === 'Property') {
          const key = prop.key?.name || prop.key?.value;

          switch (key) {
            case 'messageId':
            case 'message':
              error[key] = getStringValue(prop.value);
              break;
            case 'line':
            case 'column':
            case 'endLine':
            case 'endColumn':
              const numValue = getNumberValue(prop.value);
              if (numValue !== undefined) {
                error[key] = numValue;
              }
              break;
            case 'suggestions':
              if (prop.value?.type === 'ArrayExpression') {
                error.suggestions = [];
                for (const suggestionNode of prop.value.elements || []) {
                  if (suggestionNode) {
                    const suggestion = parseSuggestion(suggestionNode);
                    error.suggestions.push(suggestion);
                  }
                }
              }
              break;
          }
        }
      }
    } else {
      throw new Error(`Unexpected error node type: ${node.type}`);
    }

    return error;
  }

  function parseSuggestion(node: any): any {
    const suggestion: any = {};

    if (!node) {
      throw new Error('Suggestion node is null or undefined');
    }

    if (node.type === 'ObjectExpression') {
      for (const prop of node.properties || []) {
        if (prop.type === 'ObjectProperty' || prop.type === 'Property') {
          const key = prop.key?.name || prop.key?.value;

          switch (key) {
            case 'messageId':
            case 'message':
            case 'output':
              suggestion[key] = getStringValue(prop.value);
              break;
          }
        }
      }
    } else {
      throw new Error(`Unexpected suggestion node type: ${node.type}`);
    }

    return suggestion;
  }

  function getStringValue(node: any): string {
    if (!node) {
      return '';
    }

    if (node.type === 'StringLiteral' || node.type === 'Literal') {
      // Handle both StringLiteral and Literal (some parsers use different names)
      if (typeof node.value === 'string') {
        return node.value;
      }
      return node.value?.toString() || '';
    } else if (node.type === 'TemplateLiteral') {
      // For template literals, concatenate the raw text parts
      let result = '';
      for (let i = 0; i < (node.quasis?.length || 0); i++) {
        result += node.quasis[i]?.value?.raw || '';
        if (i < (node.expressions?.length || 0)) {
          // For now, we don't evaluate expressions in templates
          // This could be improved to handle simple expressions
          result += '${...}';
        }
      }
      return result;
    } else if (node.type === 'TaggedTemplateExpression') {
      // Handle tagged template expressions like noFormat`...`
      // Extract the template literal part
      if (node.quasi?.type === 'TemplateLiteral') {
        return getStringValue(node.quasi);
      } else if (node.quasi) {
        // Try to extract from quasi if it's a different structure
        let result = '';
        const quasi = node.quasi;
        if (quasi.quasis && Array.isArray(quasi.quasis)) {
          for (let i = 0; i < quasi.quasis.length; i++) {
            result += quasi.quasis[i]?.value?.raw || '';
            if (i < (quasi.expressions?.length || 0)) {
              result += '${...}';
            }
          }
        }
        return result;
      }
    }
    return '';
  }

  function getNumberValue(node: any): number | undefined {
    if (!node) {
      return undefined;
    }

    if (node.type === 'NumericLiteral') {
      return node.value;
    }
    return undefined;
  }

  traverse(parseResult.program);

  return { valid, invalid };
}

function escapeGoString(str: string): string {
  // Escape backticks in Go raw string literals
  // If the string contains backticks, we need to use a different approach
  if (str.includes('`')) {
    // For strings with backticks, use regular string literals with escaping
    return '"' + str
      .replace(/\\/g, '\\\\')
      .replace(/"/g, '\\"')
      .replace(/\n/g, '\\n')
      .replace(/\r/g, '\\r')
      .replace(/\t/g, '\\t') +
      '"';
  }
  // For strings without backticks, use raw string literals
  return '`' + str + '`';
}

function generateGoRuleImplementation(ruleName: string, testCases: RuleTester): string {
  const ruleNamePascal = kebabToPascal(ruleName);
  const packageName = ruleName.replace(/-/g, '_');

  // Extract unique messageIds from test cases
  const messageIds = new Set<string>();
  for (const testCase of testCases.invalid) {
    if (testCase.errors) {
      for (const error of testCase.errors) {
        if (error.messageId) {
          messageIds.add(error.messageId);
        }
        if (error.suggestions) {
          for (const suggestion of error.suggestions) {
            if (suggestion.messageId) {
              messageIds.add(suggestion.messageId);
            }
          }
        }
      }
    }
  }

  // Extract options structure from test cases
  const optionsFields = new Map<string, Set<string>>();
  for (const testCase of [...testCases.valid, ...testCases.invalid]) {
    if (testCase.options && testCase.options.length > 0) {
      for (const option of testCase.options) {
        if (typeof option === 'object' && !Array.isArray(option)) {
          for (const [key, value] of Object.entries(option)) {
            if (!optionsFields.has(key)) {
              optionsFields.set(key, new Set());
            }
            // Track the type of this field
            if (Array.isArray(value)) {
              optionsFields.get(key)!.add('array');
            } else if (typeof value === 'boolean') {
              optionsFields.get(key)!.add('bool');
            } else if (typeof value === 'string') {
              optionsFields.get(key)!.add('string');
            } else if (typeof value === 'number') {
              optionsFields.get(key)!.add('number');
            }
          }
        }
      }
    }
  }

  let output = `package ${packageName}

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/typescript-eslint/tsgolint/internal/rule"
)

`;

  // Generate Options struct if there are options
  if (optionsFields.size > 0) {
    output += `type ${ruleNamePascal}Options struct {
`;
    for (const [key, types] of optionsFields.entries()) {
      const fieldName = camelToPascal(key);
      let goType = '';

      // Determine Go type based on observed types
      if (types.has('array')) {
        // For now, assume string array (most common case)
        goType = '[]string';
      } else if (types.has('bool')) {
        goType = 'bool';
      } else if (types.has('string')) {
        goType = 'string';
      } else if (types.has('number')) {
        goType = 'int';
      } else {
        goType = 'interface{}';
      }

      output += `\t${fieldName} ${goType} \`json:"${key}"\`
`;
    }
    output += `}

`;
  }

  // Generate message functions for each unique messageId
  for (const messageId of messageIds) {
    const funcName = `build${messageId.charAt(0).toUpperCase() + messageId.slice(1)}Message`;
    output += `func ${funcName}() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "${messageId}",
		Description: "TODO: Add description for ${messageId}",
	}
}

`;
  }

  // Generate the main rule structure
  output += `var ${ruleNamePascal}Rule = rule.Rule{
	Name: "${ruleName}",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			// TODO: Implement the rule logic here
			// This is a stub implementation that needs to be filled in
			// based on the TypeScript ESLint rule implementation
			
			// Example listener for expressions:
			// ast.KindCallExpression: func(node *ast.Node) {
			//     // Check the node and report if necessary
			//     ctx.ReportNode(node, buildMessageFunction())
			// },
		}
	},
}
`;

  return output;
}

function generateGoTest(ruleName: string, testCases: RuleTester): string {
  const ruleNamePascal = kebabToPascal(ruleName);
  const packageName = ruleName.replace(/-/g, '_');

  let output = `package ${packageName}

import (
\t"testing"

\t"github.com/typescript-eslint/tsgolint/internal/rule_tester"
\t"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func Test${ruleNamePascal}Rule(t *testing.T) {
\trule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &${ruleNamePascal}Rule, []rule_tester.ValidTestCase{`;

  // Add valid test cases
  for (const testCase of testCases.valid) {
    const escapedCode = escapeGoString(testCase.code);
    output += `
\t\t{`;

    // Add Code field
    output += `Code: ${escapedCode}`;

    // Add Options field if present
    if (testCase.options && testCase.options.length > 0) {
      const optionsCode = convertOptionsToGoCode(testCase.options, ruleName);
      if (optionsCode) {
        output += `, ${optionsCode}`;
      }
    }

    output += `},`;
  }

  output += `
\t}, []rule_tester.InvalidTestCase{`;

  // Add invalid test cases
  for (const testCase of testCases.invalid) {
    const escapedCode = escapeGoString(testCase.code);
    output += `
\t\t{
\t\t\tCode: ${escapedCode},`;

    // Add Options field if present
    if (testCase.options && testCase.options.length > 0) {
      const optionsCode = convertOptionsToGoCode(testCase.options, ruleName);
      if (optionsCode) {
        output += `
\t\t\t${optionsCode},`;
      }
    }

    if (testCase.errors && testCase.errors.length > 0) {
      output += `
\t\t\tErrors: []rule_tester.InvalidTestCaseError{`;

      for (const error of testCase.errors) {
        output += `
\t\t\t\t{`;

        if (error.messageId) {
          output += `
\t\t\t\t\tMessageId: "${error.messageId}",`;
        }
        if (error.message) {
          output += `
\t\t\t\t\tMessage: "${error.message}",`;
        }
        if (error.line) {
          output += `
\t\t\t\t\tLine: ${error.line},`;
        }
        if (error.column) {
          output += `
\t\t\t\t\tColumn: ${error.column},`;
        }
        if (error.endLine) {
          output += `
\t\t\t\t\tEndLine: ${error.endLine},`;
        }
        if (error.endColumn) {
          output += `
\t\t\t\t\tEndColumn: ${error.endColumn},`;
        }

        if (error.suggestions && error.suggestions.length > 0) {
          output += `
\t\t\t\t\tSuggestions: []rule_tester.InvalidTestCaseSuggestion{`;

          for (const suggestion of error.suggestions) {
            output += `
\t\t\t\t\t\t{`;

            if (suggestion.messageId) {
              output += `
\t\t\t\t\t\t\tMessageId: "${suggestion.messageId}",`;
            }
            if (suggestion.message) {
              output += `
\t\t\t\t\t\t\tMessage: "${suggestion.message}",`;
            }
            if (suggestion.output) {
              const escapedOutput = escapeGoString(suggestion.output);
              output += `
\t\t\t\t\t\t\tOutput: ${escapedOutput},`;
            }

            output += `
\t\t\t\t\t\t},`;
          }

          output += `
\t\t\t\t\t},`;
        }

        output += `
\t\t\t\t},`;
      }

      output += `
\t\t\t},`;
    }

    output += `
\t\t},`;
  }

  output += `
\t})
}
`;

  return output;
}

async function downloadTestFile(ruleName: string): Promise<string> {
  const url =
    `https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main/packages/eslint-plugin/tests/rules/${ruleName}.test.ts`;

  console.log(`Downloading test file from: ${url}`);

  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`Failed to download test file: ${response.statusText}`);
  }

  return await response.text();
}

async function downloadRuleSource(ruleName: string): Promise<string> {
  const url =
    `https://raw.githubusercontent.com/typescript-eslint/typescript-eslint/main/packages/eslint-plugin/src/rules/${ruleName}.ts`;

  console.log(`Downloading rule source from: ${url}`);

  const response = await fetch(url);

  if (!response.ok) {
    throw new Error(`Failed to download rule source: ${response.statusText}`);
  }

  return await response.text();
}

function checkRequiresTypeChecking(ruleSource: string): boolean {
  // Check if the rule requires type checking by looking for requiresTypeChecking: true in the meta
  // This regex looks for the meta object and checks if requiresTypeChecking is set to true
  const metaRegex = /meta:\s*{[^}]*requiresTypeChecking:\s*true/s;

  // Also check for common type-aware patterns as a fallback
  const typeAwarePatterns = [
    /getTypeChecker\(\)/,
    /services\.program/,
    /context\.sourceCode\.parserServices/,
    /getParserServices\(/,
    /requiresTypeChecking:\s*true/,
  ];

  // First check the explicit requiresTypeChecking flag
  if (metaRegex.test(ruleSource)) {
    return true;
  }

  // Then check for type-aware patterns
  return typeAwarePatterns.some(pattern => pattern.test(ruleSource));
}

async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    console.error('Usage: pnpm run rulegen <rule-name>');
    console.error('Example: pnpm run rulegen await-thenable');
    process.exit(1);
  }

  const ruleName = args[0];

  try {
    // Download and check the rule source first
    console.log('Checking if rule requires type information...');
    let ruleSource: string;
    try {
      ruleSource = await downloadRuleSource(ruleName);
    } catch (e) {
      throw new Error(`Failed to download rule source: ${e instanceof Error ? e.message : String(e)}`);
    }

    const requiresTypeChecking = checkRequiresTypeChecking(ruleSource);
    if (!requiresTypeChecking) {
      throw new Error(
        `Rule "${ruleName}" does not require type checking. ` +
          `TSGolint is specifically for type-aware rules. ` +
          `Non-type-aware rules should be implemented in oxlint directly.`,
      );
    }

    console.log('✅ Rule requires type checking, proceeding...');

    // Download the test file
    const testContent = await downloadTestFile(ruleName);

    // Parse the TypeScript test file
    console.log('Parsing test file...');
    let result: ParseResult;
    try {
      result = parseSync(`${ruleName}.test.ts`, testContent);
    } catch (e) {
      throw new Error(`Failed to parse TypeScript test file: ${e instanceof Error ? e.message : String(e)}`);
    }

    if (!result.program) {
      throw new Error('Parser returned no program AST');
    }

    if (result.errors && result.errors.length > 0) {
      console.warn(`Warning: Parser reported ${result.errors.length} error(s) while parsing the test file`);
      // Still continue - the test file might have intentional syntax errors
    }

    // Extract test cases
    let testCases: RuleTester;
    try {
      testCases = extractTestCases(result);
    } catch (e) {
      throw new Error(`Failed to extract test cases: ${e instanceof Error ? e.message : String(e)}`);
    }

    if (testCases.valid.length === 0 && testCases.invalid.length === 0) {
      throw new Error('No test cases found. The test file might have an unexpected structure.');
    }

    console.log(`Found ${testCases.valid.length} valid test cases and ${testCases.invalid.length} invalid test cases`);

    // Generate Go files
    const goTestContent = generateGoTest(ruleName, testCases);
    const goRuleContent = generateGoRuleImplementation(ruleName, testCases);

    // Create output directory
    const outputDir = path.join(__dirname, '../../../internal/rules', ruleName.replace(/-/g, '_'));

    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    // Write the Go test file
    const testOutputPath = path.join(outputDir, `${ruleName.replace(/-/g, '_')}_test.go`);
    fs.writeFileSync(testOutputPath, goTestContent);
    console.log(`✅ Generated test file: ${testOutputPath}`);

    // Write the Go rule implementation file (only if it doesn't exist)
    const ruleOutputPath = path.join(outputDir, `${ruleName.replace(/-/g, '_')}.go`);
    if (fs.existsSync(ruleOutputPath)) {
      console.log(`⚠️  Rule implementation already exists: ${ruleOutputPath} (skipping)`);
    } else {
      fs.writeFileSync(ruleOutputPath, goRuleContent);
      console.log(`✅ Generated rule implementation: ${ruleOutputPath}`);
    }
  } catch (error) {
    console.error(`Error: ${error instanceof Error ? error.message : error}`);
    process.exit(1);
  }
}

main().catch(console.error);
