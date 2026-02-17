// @ts-check
// Generate the Go structs for lint rule options from the JSON schemas.

// 1. Look for all `schema.json` files in `internal/rules`
// 2. For each schema, generate a Go struct using `go-jsonschema` tool, and produce
//    a `.go` file next to the `schema.json` file as `option.go`, under the same package.
//    Example: `internal/rules/no_floating_promises/schema.json
//          => `internal/rules/no_floating_promises/options.go` (with package `no_floating_promises`)

import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

// ensure go-jsonschema is installed
try {
  execSync('go-jsonschema -h', { stdio: 'ignore' });
  console.log('go-jsonschema is installed.');
} catch (e) {
  console.log('go-jsonschema is not installed. Please install it first.');
  process.exit(1);
}

console.log('Generating Go structs from JSON schemas...');

// find every directory in internal/rules that contains schema.json and generate Go struct
const rulesDir = path.join(process.cwd(), 'internal', 'rules');

function findSchemaDirs(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const schemaDirs = [];

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      schemaDirs.push(...findSchemaDirs(fullPath));
    } else if (entry.isFile() && entry.name === 'schema.json') {
      schemaDirs.push(dir);
    }
  }

  return schemaDirs;
}

const schemaDirs = findSchemaDirs(rulesDir);

/**
 * Rules that should not run through go-jsonschema codegen.
 * Keep these options files manually maintained when the schema shape is unsupported.
 */
const skipSchemaGenerationRules = new Map([
  [
    'prefer_destructuring',
    'schema uses tuple-style array items, which go-jsonschema cannot parse',
  ],
]);

/**
 * Find fields in the schema that use oneOf with boolean + $ref to an object.
 * These need to be converted to utils.BoolOr[T] in the generated Go code.
 * @param {any} schema - The JSON schema
 * @returns {Array<{fieldName: string, optionsType: string, defaultValue: boolean | undefined}>}
 */
function findBoolOrFields(schema) {
  const results = [];
  const definitions = schema.definitions || {};

  // Helper to convert definition name to Go type name (e.g., ignorePrimitivesOptions -> IgnorePrimitivesOptions)
  function toGoTypeName(defName) {
    // Split by underscore and capitalize each part
    return defName
      .split('_')
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join('');
  }

  // Check each definition for properties with oneOf boolean + $ref pattern
  for (const [defName, def] of Object.entries(definitions)) {
    if (def.type !== 'object' || !def.properties) continue;

    for (const [propName, prop] of Object.entries(def.properties)) {
      if (!prop.oneOf) continue;

      // Check if oneOf contains boolean and a $ref
      const hasBool = prop.oneOf.some((o) => o.type === 'boolean');
      const refItem = prop.oneOf.find((o) => o.$ref);

      if (hasBool && refItem) {
        // Extract the referenced type name from $ref (e.g., "#/definitions/ignorePrimitivesOptions" -> "ignorePrimitivesOptions")
        const refMatch = refItem.$ref.match(/#\/definitions\/(\w+)/);
        if (refMatch) {
          const referencedDefName = refMatch[1];
          const goTypeName = toGoTypeName(referencedDefName);
          // Get the default value if it's a boolean
          const defaultValue = typeof prop.default === 'boolean' ? prop.default : undefined;
          results.push({ fieldName: propName, optionsType: goTypeName, defaultValue });
        }
      }
    }
  }

  return results;
}

for (const schemaDir of schemaDirs) {
  const schemaPath = path.join(schemaDir, 'schema.json');
  const outputPath = path.join(schemaDir, 'options.go');
  const ruleName = path.basename(schemaDir);
  const skipReason = skipSchemaGenerationRules.get(ruleName);

  if (skipReason) {
    console.log(`Skipping schema codegen for ${ruleName}: ${skipReason}`);
    continue;
  }

  console.log(`Generating Go struct for schema: ${schemaPath} and outputting to: ${outputPath}`);
  try {
    execSync(
      `go-jsonschema  "${schemaPath}" -o "${outputPath}" -p ${path.basename(schemaDir)
      } --tags json --resolve-extension json`,
      {
        stdio: 'inherit',
      },
    );

    // Post-process specific rules that need custom modifications
    let content = fs.readFileSync(outputPath, 'utf8');
    let modified = false;

    // General post-processing for ALL schemas:
    // Replace encoding/json with go-json-experiment/json for compatibility with TypeOrValueSpecifier
    // Only if the file actually uses json (has UnmarshalJSON methods)
    if (content.includes('import "encoding/json"')) {
      if (content.includes('json.Unmarshal') || content.includes('json.Marshal')) {
        content = content.replace(/^import "encoding\/json"/m, 'import "github.com/go-json-experiment/json"');
        modified = true;
      } else {
        // Remove unused json import
        content = content.replace(/^import "encoding\/json"\n/m, '');
        modified = true;
      }
    }

    // If the schema uses shared_schemas.json (TypeOrValueSpecifier), replace generated
    // interface{} types with proper utils.TypeOrValueSpecifier imports.
    const hasTypeOrValueSpecifier = /\bTypeOrValueSpecifier\b/.test(content);
    if (hasTypeOrValueSpecifier) {
      // 1. Replace element type aliases to use utils.TypeOrValueSpecifier
      const elemTypePattern = /^type (\w+Elem) interface\{\}$/gm;
      content = content.replace(elemTypePattern, 'type $1 = utils.TypeOrValueSpecifier');

      // 2. Remove TypeOrValueSpecifier interface definition (if generated)
      content = content.replace(/^type TypeOrValueSpecifier interface\{\}\s*\n/gm, '');

      // 3. Remove FileSpecifier, LibSpecifier, and PackageSpecifier types (including comments and UnmarshalJSON)
      // These are duplicates - we'll use the ones from utils instead
      content = content.replace(
        /\/\/ Describes specific types.*?[\s\S]*?type FileSpecifier struct \{[\s\S]*?\n\}\s*\n\s*\n\/\/ UnmarshalJSON[\s\S]*?func \(j \*FileSpecifier\) UnmarshalJSON[\s\S]*?\n\}\s*\n/,
        '',
      );
      content = content.replace(
        /\/\/ Describes specific types.*?lib\.\*\.d\.ts[\s\S]*?type LibSpecifier struct \{[\s\S]*?\n\}\s*\n\s*\n\/\/ UnmarshalJSON[\s\S]*?func \(j \*LibSpecifier\) UnmarshalJSON[\s\S]*?\n\}\s*\n/,
        '',
      );
      content = content.replace(
        /\/\/ Describes specific types.*?packages\.[\s\S]*?type PackageSpecifier struct \{[\s\S]*?\n\}\s*\n\s*\n\/\/ UnmarshalJSON[\s\S]*?func \(j \*PackageSpecifier\) UnmarshalJSON[\s\S]*?\n\}\s*\n/,
        '',
      );

      // 4. Replace remaining unqualified references to TypeOrValueSpecifier
      content = content.replace(/\bTypeOrValueSpecifier\b/g, (match, offset) => {
        const prev = offset > 0 ? content[offset - 1] : '';
        return prev === '.' ? match : 'utils.TypeOrValueSpecifier';
      });

      // Clean up multiple consecutive empty lines
      content = content.replace(/\n{3,}/g, '\n\n');

      // 5. Add utils import if not present
      if (!content.includes('"github.com/typescript-eslint/tsgolint/internal/utils"')) {
        if (content.includes('import "github.com/go-json-experiment/json"')) {
          content = content.replace(
            /^import "github\.com\/go-json-experiment\/json"/m,
            'import "github.com/go-json-experiment/json"\nimport "github.com/typescript-eslint/tsgolint/internal/utils"',
          );
        } else if (content.includes('import (')) {
          content = content.replace(
            /^import \(/m,
            'import (\n\t"github.com/typescript-eslint/tsgolint/internal/utils"',
          );
        } else {
          content = content.replace(
            /^(package \w+)\n/m,
            '$1\n\nimport "github.com/typescript-eslint/tsgolint/internal/utils"\n',
          );
        }
      }

      if (!content.includes('json.')) {
        content = content.replace(/^import "github\.com\/go-json-experiment\/json"\n/gm, '');
      }

      // 6. Remove unused imports if no longer needed (since we removed UnmarshalJSON methods)
      if (!content.includes('fmt.')) {
        content = content.replace(/^import "fmt"\n/gm, '');
      }

      content = content.replace(/\n{2,}$/, '\n');

      modified = true;
      console.log(`  Post-processed ${ruleName} to use utils.TypeOrValueSpecifier instead of generated types`);
    }

    if (ruleName === 'restrict_template_expressions') {
      // Remove omitempty from the Allow field to allow distinguishing between
      // "not provided" (use default) and "explicitly empty" (use empty array).
      // Without this, empty slices get omitted during JSON marshaling, causing
      // the default to be applied when it shouldn't be.
      content = content.replace(
        /Allow (\[\][^ ]+) `json:"allow,omitempty"`/,
        'Allow $1 `json:"allow"`',
      );

      // Fix default value initialization - need to properly construct utils.TypeOrValueSpecifier
      content = content.replace(
        /plain\.Allow = \[\](?:RestrictTemplateExpressionsOptionsAllowElem|TypeOrValueSpecifier|utils\.TypeOrValueSpecifier)\{\s*map\[string\]interface\{\}\{[\s\S]*?\},\s*\}\s*\}\s*$/m,
        `plain.Allow = []utils.TypeOrValueSpecifier{
			{
				From: utils.TypeOrValueSpecifierFromLib,
				Name: []string{"Error", "URL", "URLSearchParams"},
			},
		}
	}`,
      );

      modified = true;
      console.log(`  Post-processed ${ruleName} to remove omitempty from Allow field`);
    }

    if (ruleName === 'return_await') {
      // go-jsonschema doesn't handle null values for string enums, but test cases
      // that omit the Options field pass nil, which gets marshaled to null.
      // Add null handling to use the schema's default value.
      const originalUnmarshal =
        /\/\/ UnmarshalJSON implements json\.Unmarshaler\.\nfunc \(j \*ReturnAwaitOptions\) UnmarshalJSON\(value \[\]byte\) error \{/;
      const newUnmarshal = `// UnmarshalJSON implements json.Unmarshaler.
func (j *ReturnAwaitOptions) UnmarshalJSON(value []byte) error {
	// Handle null value by setting default (schema specifies "in-try-catch" as default)
	if string(value) == "null" {
		*j = ReturnAwaitOptionsInTryCatch
		return nil
	}
`;
      content = content.replace(originalUnmarshal, newUnmarshal);
      modified = true;
      console.log(`  Post-processed ${ruleName} to add null handling for default value`);
    }

    // Handle oneOf patterns with boolean + object (e.g., ignorePrimitives)
    // These generate `interface{}` which requires manual type switching. Replace with utils.BoolOr[T].
    // Skip rules that already have manual handling for these patterns.
    const skipBoolOrRules = ['no_misused_promises'];
    const schema = JSON.parse(fs.readFileSync(schemaPath, 'utf8'));
    const boolOrFields = skipBoolOrRules.includes(ruleName) ? [] : findBoolOrFields(schema);

    if (boolOrFields.length > 0) {
      for (const { fieldName, optionsType, defaultValue } of boolOrFields) {
        // Convert field name to PascalCase for Go (e.g., ignorePrimitives -> IgnorePrimitives)
        const goFieldName = fieldName.charAt(0).toUpperCase() + fieldName.slice(1);

        // Find the interface{} field and replace with utils.BoolOr[T]
        const interfacePattern = new RegExp(`(${goFieldName}\\s+)interface\\{\\}(\\s+\`json:"${fieldName}[^"]*"\`)`);
        const newContent = content.replace(interfacePattern, `$1utils.BoolOr[${optionsType}]$2`);

        if (newContent !== content) {
          content = newContent;

          // Also fix the default value assignment in UnmarshalJSON if it assigns a raw boolean
          // e.g., plain.ChecksVoidReturn = true -> plain.ChecksVoidReturn = utils.BoolOrTrue[ChecksVoidReturnOptions]()
          if (defaultValue !== undefined) {
            const defaultPattern = new RegExp(`plain\\.${goFieldName} = (true|false)`, 'g');
            const boolDefault = defaultValue ? 'true' : 'false';
            content = content.replace(
              defaultPattern,
              `plain.${goFieldName} = utils.BoolOrValue[${optionsType}](${boolDefault})`,
            );
          }

          // Add utils import if not present
          if (!content.includes('"github.com/typescript-eslint/tsgolint/internal/utils"')) {
            if (content.includes('import "github.com/go-json-experiment/json"')) {
              content = content.replace(
                /^import "github\.com\/go-json-experiment\/json"/m,
                'import "github.com/go-json-experiment/json"\nimport "github.com/typescript-eslint/tsgolint/internal/utils"',
              );
            } else if (content.includes('import (')) {
              content = content.replace(
                /^import \(/m,
                'import (\n\t"github.com/typescript-eslint/tsgolint/internal/utils"',
              );
            } else {
              // Add new import block after package declaration
              content = content.replace(
                /^(package \w+)\n/m,
                '$1\n\nimport "github.com/typescript-eslint/tsgolint/internal/utils"\n',
              );
            }
          }

          modified = true;
          console.log(
            `  Post-processed ${ruleName}: replaced ${goFieldName} interface{} with utils.BoolOr[${optionsType}]`,
          );
        }
      }
    }

    if (modified) {
      fs.writeFileSync(outputPath, content, 'utf8');
    }
  } catch (e) {
    console.error(`Failed to generate Go struct for schema: ${schemaPath}`, e);
    process.exit(1);
  }
}
