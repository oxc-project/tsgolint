#!/bin/bash

# Snapshot test script that runs all tsgolint rules and creates a deterministic snapshot

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_FILE="$SCRIPT_DIR/e2e_diagnostics.snap"

# All available rules
RULES=$(cat << 'EOF'
["await-thenable", "no-array-delete", "no-base-to-string", "no-confusing-void-expression", "no-duplicate-type-constituents", "no-floating-promises", "no-for-in-array", "no-implied-eval", "no-meaningless-void-operator", "no-misused-promises", "no-misused-spread", "no-mixed-enums", "no-redundant-type-constituents", "no-unnecessary-boolean-literal-compare", "no-unnecessary-template-expression", "no-unnecessary-type-arguments", "no-unnecessary-type-assertion", "no-unsafe-argument", "no-unsafe-assignment", "no-unsafe-call", "no-unsafe-enum-comparison", "no-unsafe-member-access", "no-unsafe-return", "no-unsafe-type-assertion", "no-unsafe-unary-minus", "non-nullable-type-assertion-style", "only-throw-error", "prefer-promise-reject-errors", "prefer-reduce-type-parameter", "prefer-return-this-type", "promise-function-async", "related-getter-setter-pairs", "require-array-sort-compare", "require-await", "restrict-plus-operands", "restrict-template-expressions", "return-await", "switch-exhaustiveness-check", "unbound-method", "use-unknown-in-catch-callback-variable"]
EOF
)

# Remove .DS_Store if it exists (macOS artifact)
[ -f ./fixtures/.DS_Store ] && rm ./fixtures/.DS_Store

# Create temp files for stderr and exit code
STDERR_FILE=$(mktemp)
EXIT_CODE_FILE=$(mktemp)

# Generate the config and run tsgolint in single-threaded mode for deterministic results
DIAGNOSTICS=$(find "$SCRIPT_DIR/fixtures" -type f -not -name "*.json" | xargs realpath | jq -Rn --argjson rules "$RULES" '{files: [inputs | {file_path: ., rules: $rules}]}' | { GOMAXPROCS=1 ./tsgolint headless 2>"$STDERR_FILE"; echo $? > "$EXIT_CODE_FILE"; } | node -e "
const fs = require('fs');

function parseHeadlessOutput(data) {
    // Parse binary headless output to extract JSON diagnostics
    let offset = 0;
    const diagnostics = [];

    while (offset < data.length) {
        if (offset + 5 > data.length) {
            break;
        }

        // Read header: 4 bytes length + 1 byte message type
        const length = data.readUInt32LE(offset);
        const msgType = data[offset + 4];
        offset += 5;

        if (offset + length > data.length) {
            break;
        }

        // Read payload
        const payload = data.slice(offset, offset + length);
        offset += length;

        // Only process diagnostic messages (type 1)
        if (msgType === 1) {
            try {
                const diagnostic = JSON.parse(payload.toString('utf-8'));
                // Make file paths relative to fixtures/ for deterministic snapshots
                const filePath = diagnostic.file_path || '';
                if (filePath.includes('fixtures/')) {
                    diagnostic.file_path = 'fixtures/' + filePath.split('fixtures/', 2)[1];
                }
                diagnostics.push(diagnostic);
            } catch (e) {
                continue;
            }
        }
    }

    return diagnostics;
}

function sortDiagnostics(diagnostics) {
    // Sort diagnostics deterministically
    diagnostics.sort((a, b) => {
        const aFilePath = a.file_path || '';
        const bFilePath = b.file_path || '';
        if (aFilePath !== bFilePath) return aFilePath.localeCompare(bFilePath);

        const aRule = a.rule || '';
        const bRule = b.rule || '';
        if (aRule !== bRule) return aRule.localeCompare(bRule);

        const aPos = (a.range && a.range.pos) || 0;
        const bPos = (b.range && b.range.pos) || 0;
        if (aPos !== bPos) return aPos - bPos;

        const aEnd = (a.range && a.range.end) || 0;
        const bEnd = (b.range && b.range.end) || 0;
        return aEnd - bEnd;
    });

    return diagnostics;
}

// Read binary data from stdin
const chunks = [];
process.stdin.on('data', chunk => chunks.push(chunk));
process.stdin.on('end', () => {
    const data = Buffer.concat(chunks);

    // Parse and sort diagnostics
    let diagnostics = parseHeadlessOutput(data);
    diagnostics = sortDiagnostics(diagnostics);

    // Output as JSON
    console.log(JSON.stringify(diagnostics, null, 2));
});
")

# Read the exit code
EXIT_CODE=$(cat "$EXIT_CODE_FILE")
rm "$EXIT_CODE_FILE"

# Check for panics in stderr
STDERR_OUTPUT=$(cat "$STDERR_FILE")
if echo "$STDERR_OUTPUT" | grep -q "panic:"; then
    echo "ERROR: Program panicked!"
    echo "Panic output:"
    echo "$STDERR_OUTPUT"
    rm "$STDERR_FILE"
    exit 1
fi

# Report exit code if non-zero
if [ "$EXIT_CODE" -ne 0 ]; then
    echo "ERROR: tsgolint exited with code $EXIT_CODE"
    if [ -n "$STDERR_OUTPUT" ]; then
        echo "stderr output:"
        echo "$STDERR_OUTPUT"
    fi
    rm "$STDERR_FILE"
    exit 1
fi

# Check if we got any diagnostics
if [ -z "$DIAGNOSTICS" ] || [ "$DIAGNOSTICS" = "[]" ]; then
    echo "No diagnostics generated. This might indicate an issue with the test."
    if [ -n "$STDERR_OUTPUT" ]; then
        echo "stderr output:"
        echo "$STDERR_OUTPUT"
    fi
    rm "$STDERR_FILE"
    exit 1
fi

# Clean up stderr file
rm "$STDERR_FILE"

# Count diagnostics
DIAGNOSTIC_COUNT=$(echo "$DIAGNOSTICS" | jq '. | length')
echo "Generated $DIAGNOSTIC_COUNT diagnostics"

echo "$DIAGNOSTICS" > "$SNAPSHOT_FILE"
echo "Updated snapshot file: $SNAPSHOT_FILE"

# Only check if the snapshot file has changed
git diff --exit-code "$SNAPSHOT_FILE"
