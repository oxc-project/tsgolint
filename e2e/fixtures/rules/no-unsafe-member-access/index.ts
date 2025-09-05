// Examples of incorrect code for no-unsafe-member-access rule

declare const anyValue: any;

anyValue.foo; // unsafe member access

anyValue.bar.baz; // unsafe nested member access

anyValue['key']; // unsafe computed member access

const result = anyValue.method(); // unsafe method access

// Issue #13550: assignment to member access
const a: any = {};
a.foo = 45; // unsafe member access in assignment