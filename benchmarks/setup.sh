#!/usr/bin/env -S bash -euxo pipefail

ESLINT_VERSION=9.23.0
TYPESCRIPT_ESLINT_VERSION=8.64.0

for proj in {vscode,typescript,typeorm}; do
  pushd $proj
  cp ../eslint.config.mjs ./eslint.config.mjs
  npm install --ignore-scripts -D "eslint@$ESLINT_VERSION" "typescript-eslint@$TYPESCRIPT_ESLINT_VERSION"
  popd
done

pushd vuejs
cp ../eslint.config.mjs ./eslint.config.mjs
pnpm install --ignore-scripts -w -D "eslint@$ESLINT_VERSION" "typescript-eslint@$TYPESCRIPT_ESLINT_VERSION"
popd
