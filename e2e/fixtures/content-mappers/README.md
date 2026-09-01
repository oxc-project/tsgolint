# Content mapper fixture

An Ember project whose `.gts` files are type-checked through
[`ember-content-mapper`](https://www.npmjs.com/package/ember-content-mapper), a TypeScript 7 content
mapper that wraps Glint.

This fixture is **not** part of the automated e2e suite: it needs a real `npm install` and Node >=
22.21, which the suite does not otherwise require. The automated coverage of the content mapper path
lives in `internal/linter/content_mapper_test.go`, which drives a dependency-free mapper.

## Running it

```sh
npm install                 # in this directory
just build                  # in the repo root
cd e2e/fixtures/content-mappers && ../../../tsgolint
```

Expected: two type-aware findings inside `<template>` in `widget.gts` (`no-unnecessary-condition` and
`strict-boolean-expressions` on `this.always`), one in the script section of `fixable.gts`
(`no-unnecessary-type-assertion` on `n as number`), and nothing anchored on Glint's `__glintDSL__`
scaffolding.

`counter.gts` carries a deliberate type error inside `<template>` (`{{this.cuont}}` against a class
that declares `count`). The CLI does not report type errors, so check it through the headless
interface with `report_semantic`; it should be reported as TS2551 over the `.gts` offsets of `cuont`.

## Known wart: a broken mapper can report nothing useful

When the `contentMappers` entry names a package that does not resolve, TypeScript drops the mapper and
unregisters its extensions. The `.gts` files then match no tsconfig, so tsgolint builds no program for
that config and its "could not be resolved" error never surfaces — all you get is the
`unsupported-file-extension` diagnostic, which does not mention the mapper that was supposed to claim
the file.

Whether the real error appears depends on the **set of files being linted**, not on what the tsconfig
contains: it surfaces only when some natively parseable file from the same tsconfig is also in the
lint set, because that is what causes a program to be built. Linting a whole project usually includes
one; linting a single `.gts` (an editor, or `oxlint path/to/component.gts`) does not, so the same
broken config reports the mapper error in a project-wide run and stays silent in a single-file one.
