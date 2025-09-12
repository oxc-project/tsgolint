// File with both type errors and lint errors for testing

// Type error: wrong type assignment
const x: number = "wrong type"; // TS2322

// Lint error: async function without await
async function noAwait() { // require-await lint error
  return 42;
}

// Lint error: floating promise
noAwait(); // no-floating-promises lint error

// Type error: missing property
interface User {
  name: string;
  email: string;
}

const user: User = { // TS2741: Property 'email' is missing
  name: "John"
};

// Both type and lint errors in same function
async function problematicFunction(): Promise<number> { // require-await lint error
  const result: string = 123; // TS2322: Type error
  return "not a number"; // TS2322: Type error
}

// Lint error: unsafe operations with any
const anyValue: any = "test";
const unsafeAssignment: number = anyValue; // no-unsafe-assignment lint error
anyValue.someMethod(); // no-unsafe-call, no-unsafe-member-access lint errors

// Type error: incorrect function arguments
function strictFunction(a: string, b: number): void {
  console.log(a, b);
}
strictFunction(123, "wrong"); // TS2345: Argument type errors