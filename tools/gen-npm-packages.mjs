// @ts-check
/// <reference lib="es2023" />

import fs from 'node:fs/promises'
import path from 'node:path'
import assert from 'node:assert/strict'
import process from 'node:process'

const npmPackageVersion = requiredEnvVar('TSGOLINT_VERSION')

const NPM_ORG = `oxlint-tsgolint`

const GOOS2PROCESS_PLATFORM = {
  windows: 'win32',
  linux: 'linux',
  darwin: 'darwin',
}
const GOARCH2PROCESS_ARCH = {
  amd64: 'x64',
  arm64: 'arm64',
}

const binariesMatrix = Object.entries(GOOS2PROCESS_PLATFORM).flatMap(
  ([goos, platform]) =>
    Object.entries(GOARCH2PROCESS_ARCH).map(([goarch, arch]) => ({
      goarch,
      goos,
      arch,
      platform,
      artifactName: `tsgolint-${goos}-${goarch}`,
      npmPackageName: `@${NPM_ORG}/${platform}-${arch}`,
    })),
)

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

const repoRoot = path.join(import.meta.dirname, '..')

const npmDir = path.join(repoRoot, 'npm')
const licensePath = path.join(repoRoot, 'LICENSE')
const buildDir = path.join(repoRoot, 'build')

await Promise.all([
  ...binariesMatrix.map(
    async ({ arch, platform, artifactName, npmPackageName }) => {
      const packageName = `${platform}-${arch}`
      const packageDir = path.join(npmDir, packageName)
      const binaryName = `tsgolint${platform === 'win32' ? '.exe' : ''}`

      await fs.rm(packageDir, { recursive: true, force: true })
      await fs.mkdir(packageDir)
      await Promise.all([
        fs.writeFile(
          path.join(packageDir, 'package.json'),
          JSON.stringify(
            {
              ...commonPackageJson,
              name: npmPackageName,
              preferUnplugged: true,
              files: [binaryName],
              os: [platform],
              arch: [arch],
            },
            null,
            2,
          ),
        ),
        fs.copyFile(licensePath, path.join(packageDir, 'LICENSE')),
        fs.copyFile(
          path.join(buildDir, artifactName, 'tsgolint'),
          path.join(packageDir, binaryName),
        ),
      ])
    },
  ),
  (async () => {
    const packageDir = path.join(npmDir, 'core')
    await Promise.all([
      fs.writeFile(
        path.join(packageDir, 'package.json'),
        JSON.stringify(
          {
            ...commonPackageJson,
            name: 'oxlint-tsgolint',
            bin: {
              tsgolint: './bin/tsgolint.js',
            },
            optionalDependencies: Object.fromEntries(
              binariesMatrix.map(({ npmPackageName }) => [
                npmPackageName,
                npmPackageVersion,
              ]),
            ),
          },
          null,
          2,
        ),
      ),
      fs.copyFile(licensePath, path.join(packageDir, 'LICENSE')),
    ])
  })(),
])

function requiredEnvVar(/** @type {string} */ name) {
  const value = process.env[name]
  assert.ok(value != null, `missing $${name} env variable`)
  return value
}
