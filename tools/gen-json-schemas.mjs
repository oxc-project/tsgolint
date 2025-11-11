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

for (const schemaDir of schemaDirs) {
  const schemaPath = path.join(schemaDir, 'schema.json');
  const outputPath = path.join(schemaDir, 'options.go');

  console.log(
    `Generating Go struct for schema: ${schemaPath} and outputting to: ${outputPath}`,
  );
  try {
    execSync(
      `go-jsonschema "${schemaPath}" -o "${outputPath}" -p ${
        path.basename(
          schemaDir,
        )
      } --tags json`,
      { stdio: 'inherit' },
    );
  } catch (e) {
    console.error(`Failed to generate Go struct for schema: ${schemaPath}`, e);
    process.exit(1);
  }
}
