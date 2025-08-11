#!/bin/bash

 find "$(pwd)/fixtures" -type f -not -name "*.json" | jq -Rn '{files: [inputs | {file_path: ., rules: ["no-floating-promises"]}]}' | ./tsgolint headless
