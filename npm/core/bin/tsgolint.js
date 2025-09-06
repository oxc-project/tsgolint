#!/usr/bin/env node

const process = require('node:process');
const child_process = require('node:child_process');

const exePath = require.resolve(
  `@oxlint-tsgolint/${process.platform}-${process.arch}/tsgolint${process.platform === 'win32' ? '.exe' : ''}`,
);

try {
  child_process.execFileSync(exePath, process.argv.slice(2), {
    stdio: 'inherit',
  });
} catch (e) {
  if (e.status) {
    process.exitCode = e.status;
  } else {
    throw e;
  }
}
