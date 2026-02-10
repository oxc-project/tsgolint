// Examples of incorrect code for no-invalid-void-type rule

type AliasVoid = void;

type InvalidUnion = string | void;

function takesVoidParam(arg: void) {}

declare function generic<T>(): T;

function callGeneric() {
  generic<void>();
}
