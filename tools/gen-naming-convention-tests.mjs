// Usage: node --experimental-vm-modules tools/gen-naming-convention-tests.mjs <upstream checkout> [--check]
// The checkout/archive must be typescript-eslint at UPSTREAM. No upstream tests are rewritten.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import { stripTypeScriptTypes } from 'node:module';
import path from 'node:path';
import vm from 'node:vm';

const UPSTREAM = 'afb56de7d0123c79d2a2ffbb28ff7a8f8d951eab';
const root = path.resolve(process.argv[2]);
const check = process.argv.includes('--check');
const output = 'internal/rules/naming_convention/testdata';
const base = 'packages/eslint-plugin/tests/rules/naming-convention';
const sources = {};
const read = relative => {
  const source = fs.readFileSync(path.join(root, relative), 'utf8');
  sources[relative] = createHash('sha256').update(source).digest('hex');
  return source;
};
const ruleSource = read('packages/eslint-plugin/src/rules/naming-convention.ts');
const messages = vm.runInNewContext('(' + ruleSource.match(/messages: (\{[\s\S]*?\n    \}),\n    schema:/)[1] + ')');
const suites = [];
let current;
class RuleTester {
  run(name, rule, cases) {
    assert.equal(name, 'naming-convention');
    suites.push({ file: current, cases });
  }
}
const context = vm.createContext({});
const modules = new Map();
function synthetic(key, exports) {
  if (!modules.has(key)) modules.set(key, new vm.SyntheticModule(Object.keys(exports), function() {
    for (const [name, value] of Object.entries(exports)) this.setExport(name, value);
  }, { context, identifier: key }));
  return modules.get(key);
}
function moduleFor(file) {
  if (!modules.has(file)) {
    const source = stripTypeScriptTypes(read(file), { mode: 'strip' });
    modules.set(file, new vm.SourceTextModule(source, { context, identifier: file }));
  }
  return modules.get(file);
}
async function linker(id, parent) {
  if (id === '@typescript-eslint/rule-tester') return synthetic(id, {
    RuleTester, noFormat: (raw, ...keys) => String.raw({ raw }, ...keys),
  });
  if (id.endsWith('/src/rules/naming-convention')) return synthetic('rule', { default: {} });
  if (id.endsWith('/src/rules/naming-convention-utils')) return moduleFor('packages/eslint-plugin/src/rules/naming-convention-utils/shared.ts');
  if (id === './enums' && parent.identifier.endsWith('/src/rules/naming-convention-utils/shared.ts')) {
    // The imported enum is used only by helper functions that case expansion never calls.
    return synthetic('enums', { MetaSelectors: {} });
  }
  if (id.endsWith('/RuleTester')) return synthetic('fixtures', { getFixturesRootDir: () => '<fixtures>' });
  assert.ok(id.startsWith('.'), `Unaccounted import: ${id}`);
  return moduleFor(path.posix.normalize(path.posix.join(path.posix.dirname(parent.identifier), id + '.ts')));
}
const files = ['naming-convention.test.ts', ...fs.readdirSync(path.join(root, base, 'cases')).filter(f => f.endsWith('.test.ts')).sort().map(f => 'cases/' + f)];
for (current of files) {
  const module = moduleFor(base + '/' + current);
  await module.link(linker);
  await module.evaluate();
}
function stable(value) {
  if (Array.isArray(value)) return value.map(stable);
  if (value !== null && typeof value === 'object') return Object.fromEntries(Object.keys(value).sort().map(k => [k, stable(value[k])]));
  return value;
}
const fingerprint = value => createHash('sha256').update(JSON.stringify(stable(value))).digest('hex');
function keys(value, allowed) {
  for (const key of Object.keys(value)) assert.ok(allowed.includes(key), `Unhandled field: ${key}`);
}
const manifest = { upstream: UPSTREAM, sources, suites: [], valid: 0, invalid: 0, diagnostics: 0 };
function write(file, value) {
  const text = JSON.stringify(value, null, 2) + '\n';
  if (check) assert.equal(fs.readFileSync(file, 'utf8'), text, `Generated file differs: ${file}`);
  else fs.writeFileSync(file, text);
}
for (const { file, cases } of suites) {
  keys(cases, ['assertionOptions', 'valid', 'invalid']);
  if (cases.assertionOptions) assert.equal(JSON.stringify(cases.assertionOptions), '{"requireData":true}');
  const result = { valid: [], invalid: [] };
  for (const kind of ['valid', 'invalid']) for (const c of cases[kind]) {
    keys(c, ['code', 'options', 'languageOptions', 'errors']);
    if (c.languageOptions) assert.equal(JSON.stringify(c.languageOptions), '{"parserOptions":{"project":"./tsconfig.json","projectService":false,"tsconfigRootDir":"<fixtures>"}}');
    for (const e of c.errors ?? []) {
      keys(e, ['messageId', 'data', 'line', 'column', 'endLine', 'endColumn']);
      assert.ok(e.messageId in messages);
      if (e.data) {
        const required = [...messages[e.messageId].matchAll(/\{\{(\w+)\}\}/g)].map(m => m[1]);
        assert.deepEqual(Object.keys(e.data).sort(), required.sort());
      }
    }
    // Preserve every original field, including absent options vs [], null, and diagnostic data.
    result[kind].push(stable(c));
  }
  const name = file.replace('cases/', '').replace('.test.ts', '');
  const hashes = {};
  for (const kind of ['valid', 'invalid']) {
    hashes[kind] = result[kind].map(fingerprint);
    manifest[kind] += result[kind].length;
  }
  manifest.diagnostics += result.invalid.reduce((n, c) => n + c.errors.length, 0);
  manifest.suites.push({ file, fixture: name + '.json', ...hashes });
  write(output + '/' + name + '.json', result);
}
assert.equal(suites.length, 17);
assert.equal(manifest.valid, 8966);
assert.equal(manifest.invalid, 7146);
assert.equal(manifest.diagnostics, 44065);
write(output + '/messages.json', stable(messages));
write(output + '/manifest.json', manifest);
console.log(`${check ? 'Verified' : 'Generated'} ${manifest.valid} valid + ${manifest.invalid} invalid cases; ${manifest.diagnostics} diagnostics across ${suites.length} suites (${UPSTREAM}).`);
