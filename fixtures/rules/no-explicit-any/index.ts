// Examples of incorrect code for no-explicit-any rule

// Variable declarations with explicit any
const value: any = "hello";
let data: any;
var info: any = {};

// Function parameters with explicit any
function processData(data: any) {
  return data;
}

// Function return types with explicit any
function getData(): any {
  return "hello";
}

// Rest parameters with explicit any
function processArgs(...args: any[]) {
  return args;
}

// Method declarations with explicit any
class Example {
  method(param: any): any {
    return param;
  }
  
  getData(): any {
    return "data";
  }
}

// Property declarations with explicit any
interface Config {
  data: any;
  options: any;
}

// Type aliases with explicit any
type DataType = any;
type ConfigType = any;

// Type annotations with explicit any
const typedValue: any = "value";
const array: any[] = [];
const object: { [key: string]: any } = {};

// Generic type parameters with explicit any
function genericFunction<T = any>(param: T): T {
  return param;
}

// Interface properties with explicit any
interface TestInterface {
  prop1: any;
  prop2: any[];
  prop3: { [key: string]: any };
}
