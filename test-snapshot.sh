#!/bin/bash

# Snapshot test script that runs all tsgolint rules and creates a deterministic snapshot

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNAPSHOT_FILE="$SCRIPT_DIR/diagnostics.snapshot"

# All available rules
RULES=$(cat << 'EOF'
["await-thenable", "no-array-delete", "no-base-to-string", "no-confusing-void-expression", "no-duplicate-type-constituents", "no-floating-promises", "no-for-in-array", "no-implied-eval", "no-meaningless-void-operator", "no-misused-promises", "no-misused-spread", "no-mixed-enums", "no-redundant-type-constituents", "no-unnecessary-boolean-literal-compare", "no-unnecessary-template-expression", "no-unnecessary-type-arguments", "no-unnecessary-type-assertion", "no-unsafe-argument", "no-unsafe-assignment", "no-unsafe-call", "no-unsafe-enum-comparison", "no-unsafe-member-access", "no-unsafe-return", "no-unsafe-type-assertion", "no-unsafe-unary-minus", "non-nullable-type-assertion-style", "only-throw-error", "prefer-promise-reject-errors", "prefer-reduce-type-parameter", "prefer-return-this-type", "promise-function-async", "related-getter-setter-pairs", "require-array-sort-compare", "require-await", "restrict-plus-operands", "restrict-template-expressions", "return-await", "switch-exhaustiveness-check", "unbound-method", "use-unknown-in-catch-callback-variable"]
EOF
)

echo "Running tsgolint with all rules..."

# Generate the config and run tsgolint in single-threaded mode for deterministic results
DIAGNOSTICS=$(find "$SCRIPT_DIR/fixtures" -type f -not -name "*.json" | xargs realpath | jq -Rn --argjson rules "$RULES" '{files: [inputs | {file_path: ., rules: $rules}]}' | GOMAXPROCS=1 ./tsgolint headless | python3 -c "
import sys
import json
import struct
import os

def parse_headless_output(data):
    '''Parse binary headless output to extract JSON diagnostics.'''
    offset = 0
    diagnostics = []
    
    while offset < len(data):
        if offset + 5 > len(data):
            break
        
        # Read header: 4 bytes length + 1 byte message type
        length = struct.unpack('<I', data[offset:offset+4])[0]
        msg_type = data[offset+4]
        offset += 5
        
        if offset + length > len(data):
            break
        
        # Read payload
        payload = data[offset:offset+length]
        offset += length
        
        # Only process diagnostic messages (type 1)
        if msg_type == 1:
            try:
                diagnostic = json.loads(payload.decode('utf-8'))
                # Make file paths relative to fixtures/ for deterministic snapshots
                file_path = diagnostic.get('file_path', '')
                if 'fixtures/' in file_path:
                    diagnostic['file_path'] = 'fixtures/' + file_path.split('fixtures/', 1)[1]
                diagnostics.append(diagnostic)
            except Exception as e:
                continue
    
    return diagnostics

def sort_diagnostics(diagnostics):
    '''Sort diagnostics deterministically.'''
    def sort_key(diag):
        return (
            diag.get('file_path', ''),
            diag.get('rule', ''),
            diag.get('range', {}).get('pos', 0),
            diag.get('range', {}).get('end', 0)
        )
    
    diagnostics.sort(key=sort_key)
    return diagnostics

# Read binary data from stdin
data = sys.stdin.buffer.read()

# Parse and sort diagnostics
diagnostics = parse_headless_output(data)
diagnostics = sort_diagnostics(diagnostics)

# Output as JSON
print(json.dumps(diagnostics, indent=2, sort_keys=True))
")

echo "Processing diagnostic output..."

# Check if we got any diagnostics
if [ -z "$DIAGNOSTICS" ] || [ "$DIAGNOSTICS" = "[]" ]; then
    echo "No diagnostics generated. This might indicate an issue with the test."
    exit 1
fi

# Count diagnostics
DIAGNOSTIC_COUNT=$(echo "$DIAGNOSTICS" | jq '. | length')
echo "Generated $DIAGNOSTIC_COUNT diagnostics"

# Save snapshot if update mode is requested
if [[ "${1:-}" == "--update" ]]; then
    echo "$DIAGNOSTICS" > "$SNAPSHOT_FILE"
    echo "Updated snapshot file: $SNAPSHOT_FILE"
    exit 0
fi

# Check if snapshot exists
if [[ ! -f "$SNAPSHOT_FILE" ]]; then
    echo "Snapshot file does not exist: $SNAPSHOT_FILE"
    echo "Run with --update to create it:"
    echo "  $0 --update"
    exit 1
fi

# Compare with existing snapshot
EXISTING_SNAPSHOT=$(cat "$SNAPSHOT_FILE")
if [[ "$DIAGNOSTICS" = "$EXISTING_SNAPSHOT" ]]; then
    echo "✅ Snapshot test passed - diagnostics match expected output"
    exit 0
else
    echo "❌ Snapshot test failed - diagnostics do not match expected output"
    echo ""
    echo "To see the diff:"
    echo "  diff <(cat '$SNAPSHOT_FILE') <(echo '$DIAGNOSTICS')"
    echo ""
    echo "To update the snapshot:"
    echo "  $0 --update"
    exit 1
fi