// @ts-check
// Generate the Go structs for lint rule options from the JSON schemas.

// 1. Look for all `schema.json` files in `internal/rules`
// 2. For each schema, generate a Go struct using `go-jsonschema` tool, and produce
//    a `.go` file next to the `schema.json` file as `option.go`, under the same package.
//    Example: `internal/rules/no_floating_promises/schema.json
//          => `internal/rules/no_floating_promises/options.go` (with package `no_floating_promises`)

/*

Generates Go code from JSON Schema files.

Usage:
  go-jsonschema FILE ... [flags]

Flags:
      --capitalization strings        Specify a preferred Go capitalization for a string. For example, by default a field
                                      named 'id' becomes 'Id'. With --capitalization ID, it will be generated as 'ID'.
      --disable-readonly-validation   Do not include validation of readonly fields
  -e, --extra-imports                 Allow extra imports (non standard library)
  -h, --help                          help for go-jsonschema
      --min-sized-ints                Uses sized int and uint values based on the min and max values for the field
      --minimal-names                 Uses the shortest possible names
      --only-models                   Generate only models (no unmarshal methods, no validation)
  -o, --output string                 File to write (- for standard output) (default "-")
  -p, --package string                Default name of package to declare Go files under, unless overridden with
                                      --schema-package
      --resolve-extension strings     Add a file extension that is used to resolve schema names, e.g. {"$ref": "./foo"} will
                                      also look for foo.json if --resolve-extension json is provided.
      --schema-output strings         File to write (- for standard output) a specific schema ID to;
                                      must be in the format URI=FILENAME.
      --schema-package strings        Name of package to declare Go files for a specific schema ID under;
                                      must be in the format URI=PACKAGE.
      --schema-root-type strings      Override name to use for the root type of a specific schema ID;
                                      must be in the format URI=TYPE. By default, it is derived from the file name.
  -t, --struct-name-from-title        Use the schema title as the generated struct name
      --tags strings                  Specify which struct tags to generate. Defaults are json, yaml, mapstructure (default [json,yaml,mapstructure])
  -v, --verbose                       Verbose output
      --yaml-extension strings        Add a file extension that should be recognized as YAML. Default are .yml, .yaml. (default [.yml,.yaml])

      */

import fs from "fs";
import path from "path";
import { execSync } from "child_process";

// ensure go-jsonschema is installed
try {
  execSync("go-jsonschema -h", { stdio: "ignore" });
  console.log("go-jsonschema is installed.");
} catch (e) {
  console.log("go-jsonschema is not installed. Please install it first.");
  process.exit(1);
}

console.log("Generating Go structs from JSON schemas...");

// find every directory in internal/rules that contains schema.json and generate Go struct
const rulesDir = path.join(process.cwd(), "internal", "rules");

function findSchemaDirs(dir) {
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  const schemaDirs = [];

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      schemaDirs.push(...findSchemaDirs(fullPath));
    } else if (entry.isFile() && entry.name === "schema.json") {
      schemaDirs.push(dir);
    }
  }

  return schemaDirs;
}

const schemaDirs = findSchemaDirs(rulesDir);

for (const schemaDir of schemaDirs) {
  const schemaPath = path.join(schemaDir, "schema.json");
  const outputPath = path.join(schemaDir, "options.go");

  console.log(
    `Generating Go struct for schema: ${schemaPath} and outputting to: ${outputPath}`
  );
  try {
    execSync(
      `go-jsonschema "${schemaPath}" -o "${outputPath}" -p ${path.basename(
        schemaDir
      )} --tags json`,
      { stdio: "inherit" }
    );
  } catch (e) {
    console.error(`Failed to generate Go struct for schema: ${schemaPath}`, e);
    process.exit(1);
  }
}
