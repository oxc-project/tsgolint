#!/bin/bash

 find "$(pwd)/fixtures/src" -type f | jq -Rn '{files: [inputs | {file_path: ., rules: ["no-floating-promises"]}]}' | ./tsgolint headless
