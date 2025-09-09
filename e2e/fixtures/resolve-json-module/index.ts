import { default as foo } from './json-module.json';

function foobar(_a: string, _b: string) {}

// if `./json-module.json` was treated as `any`, this would be a linter error (no-unsafe-member-access)
foo.bar;

// if `./json-module.json` was treated as `any`, this would be a linter error (no-unsafe-argument)
foobar(foo);
