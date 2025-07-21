// @ts-check
/// <reference lib="es2023" />

import fs from 'node:fs/promises'
import path from 'node:path'
import assert from 'node:assert/strict'
import process from 'node:process'

const NPM_ORG = `oxlint-tsgolint`

const GOOS2PROCESS_PLATFORM = {
  'windows': 'win32',
  'linux': 'linux',
  'darwin': 'darwin',
}
const GOARCH2PROCESS_ARCH = {
  'amd64': 'x64',
  'arm64': 'arm64',
}

const binariesMatrix = Object.values(GOOS2PROCESS_PLATFORM)
  .flatMap(platform => Object.values(GOARCH2PROCESS_ARCH).map(arch => ({ 
    arch,
    platform,
    npmPackageName: `@${NPM_ORG}/${platform}-${arch}`,
  })))

const BUILD_NUMBER = requiredEnvVar('TSGOLINT_BUILD_NUMBER')

const BUILD_DATE = new Date().toISOString().slice(0, 10).replaceAll('-', '')

const npmPackageVersion = `0.0.0-${BUILD_NUMBER}+${BUILD_DATE}`

const commonPackageJson = {
  version: npmPackageVersion,
  license: 'MIT',
  author: 'auvred <aauvred@gmail.com>',
  repository: 'github:oxc-project/tsgolint',
  bugs: 'https://github.com/oxc-project/tsgolint/issues',
  homepage: 'https://github.com/oxc-project/tsgolint#readme',
  publishConfig: {
    access: 'public',
  },
}

const npmDir = path.join(import.meta.dirname, '..', 'npm')
const licensePath = path.join(import.meta.dirname, '..', 'LICENSE')

await Promise.all([
  ...binariesMatrix
    .map(async ({ arch, platform, npmPackageName }) => {
      const packageName = `${platform}-${arch}`

      const packageDir = path.join(npmDir, packageName)

      await fs.rm(packageDir, { recursive: true, force: true })
      await fs.mkdir(packageDir)
      await Promise.all([
        fs.writeFile(
          path.join(packageDir, 'package.json'),
          JSON.stringify({
            ...commonPackageJson,
            name: npmPackageName,
            preferUnplugged: true,
            os: [platform],
            arch: [arch],
          }, null, 2)
        ),
        fs.copyFile(licensePath, path.join(packageDir, 'LICENSE')),
      ])
    }),
  (async () => {
    const packageDir = path.join(npmDir, 'core')
    await Promise.all([
      fs.writeFile(
        path.join(packageDir, 'package.json'),
        JSON.stringify({
          ...commonPackageJson,
          name: `${NPM_ORG}/core`,
          bin: {
            tsgolint: './bin/tsgolint.js',
          },
          optionalDependencies: Object.fromEntries(
            binariesMatrix
              .map(({ npmPackageName }) => [
                npmPackageName,
                npmPackageVersion,
              ])
          )
        }, null, 2)
      ),
      fs.copyFile(licensePath, path.join(packageDir, 'LICENSE')),
    ])
  })()
])

function requiredEnvVar(/** @type {string} */ name) {
  const value = process.env[name]
  assert.ok(value != null, `missing $${name} env variable`)
  return value
}
