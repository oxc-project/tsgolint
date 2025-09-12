// File with no type errors or lint errors for testing

// Correct type assignments
const correctNumber: number = 42;
const correctString: string = "hello";
const correctBoolean: boolean = true;

// Correct function with proper types
function multiply(a: number, b: number): number {
  return a * b;
}

const result = multiply(5, 10);

// Correct interface usage
interface Product {
  id: number;
  name: string;
  price: number;
}

const validProduct: Product = {
  id: 1,
  name: "Laptop",
  price: 999.99
};

// Correct array types
const numberArray: number[] = [1, 2, 3, 4, 5];
const stringArray: string[] = ["a", "b", "c"];

// Correct async function with await
async function fetchData(): Promise<string> {
  const data = await Promise.resolve("data");
  return data;
}

// Properly handled promise
fetchData().then(data => console.log(data));

// Correct generic usage
function genericIdentity<T>(value: T): T {
  return value;
}

const numIdentity = genericIdentity(100);
const strIdentity = genericIdentity("test");

// Correct optional chaining
const safeObj: { prop?: { nested?: string } } = {
  prop: { nested: "value" }
};
const safeValue = safeObj.prop?.nested ?? "default";

// Export to prevent unused variable warnings
export { correctNumber, validProduct, numberArray, safeValue };