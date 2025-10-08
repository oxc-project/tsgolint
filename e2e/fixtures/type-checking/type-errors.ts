// File with various TypeScript type errors for testing

// Type mismatch errors
const numberVar: number = "string"; // TS2322: Type 'string' is not assignable to type 'number'
const stringVar: string = 123; // TS2322: Type 'number' is not assignable to type 'string'

// Function parameter type errors
function addNumbers(a: number, b: number): number {
  return a + b;
}
addNumbers("1", "2"); // TS2345: Argument of type 'string' is not assignable to parameter of type 'number'
addNumbers(1); // TS2554: Expected 2 arguments, but got 1

// Property access errors
const obj = { name: "test", age: 25 };
console.log(obj.unknownProperty); // TS2339: Property 'unknownProperty' does not exist

// Interface/type errors
interface Person {
  name: string;
  age: number;
}

const invalidPerson: Person = {
  name: "John",
  // Missing 'age' property - TS2741
};

// Array type errors
const numbers: number[] = [1, 2, "3"]; // TS2322: Type 'string' is not assignable to type 'number'

// Null/undefined errors (with strict null checks)
let nullableString: string = null; // TS2322: Type 'null' is not assignable to type 'string'
let undefinedNumber: number = undefined; // TS2322: Type 'undefined' is not assignable to type 'number'

// Return type errors
function returnString(): string {
  return 42; // TS2322: Type 'number' is not assignable to type 'string'
}

// Generic type errors
function identity<T>(value: T): T {
  return "wrong"; // TS2322: Type 'string' is not assignable to type 'T'
}

// Const assertion and readonly errors
const readonlyArray = [1, 2, 3] as const;
readonlyArray[0] = 4; // TS2540: Cannot assign to '0' because it is a read-only property

// Optional chaining type errors
const maybeObj: { prop?: { nested: string } } = {};
const value: string = maybeObj.prop.nested; // TS2532: Object is possibly 'undefined'