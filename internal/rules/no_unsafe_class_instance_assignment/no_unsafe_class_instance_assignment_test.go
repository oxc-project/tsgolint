package no_unsafe_class_instance_assignment

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

var reactTypesFiles = map[string]string{
	"react-jsx.d.ts": `declare namespace JSX {
  interface ElementChildrenAttribute {
    children: {};
  }
}
`,
}

func TestNoUnsafeClassInstanceAssignmentRuleRecursiveTypes(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeClassInstanceAssignmentRule, []rule_tester.ValidTestCase{
		{Code: `
class Point { x = 0; }
type Source = { [key: string]: Source };
type Target = { [key: string]: Target | Point };
declare const source: Source;
const target: Target = source;
`},
		{Code: `
class Point { x = 0; }
type Source = { [key: number]: Source };
type Target = { [key: number]: Target | Point };
declare const source: Source;
const target: Target = source;
`},
		{Code: `
class Point { x = 0; }
type Source = { [key: string]: SourceChild };
type SourceChild = { [key: number]: Source };
type Target = { [key: string]: TargetChild | Point };
type TargetChild = { [key: number]: Target | Point };
declare const source: Source;
const target: Target = source;
`},
		{Code: `
class Point { x = 0; }
type Source = Source[];
type Target = (Target | Point)[];
declare const source: Source;
const target: Target = source;
`},
	}, []rule_tester.InvalidTestCase{
		{Code: `
class Point { x = 0; }
type Source = { [key: number]: Source; [key: string]: Source | { x: number } };
type Target = { [key: number]: Target; [key: string]: Target | Point };
declare const source: Source;
const target: Target = source;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonInstance", Line: 6, Column: 24, EndLine: 6, EndColumn: 30}}},
		{Code: `
class Point { x = 0; }
type Source = [Source, { x: number }];
type Target = [Target, Point];
declare const source: Source;
const target: Target = source;
const another: Target = source;
`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonInstance", Line: 6, Column: 24, EndLine: 6, EndColumn: 30},
			{MessageId: "nonInstance", Line: 7, Column: 25, EndLine: 7, EndColumn: 31},
		}},
		{Code: `
class Point { x = 0; }
class OtherPoint { x = 0; }
type Source = [Source, OtherPoint, OtherPoint];
type Target = [Target, Point | OtherPoint, Point];
declare const source: Source;
const target: Target = source;
`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nonInstance", Line: 7, Column: 24, EndLine: 7, EndColumn: 30}}},
	})
}

func TestNoUnsafeClassInstanceAssignmentRule(t *testing.T) {
	t.Parallel()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.minimal.json", t, &NoUnsafeClassInstanceAssignmentRule, []rule_tester.ValidTestCase{
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point = new Point(1, 2);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = new Point(1, 2);
    `},
		{Code: `
interface Point {
  x: number;
  y: number;
}
const point: Point = { x: 1, y: 2 };
    `},
		{Code: `
type Point = { x: number; y: number };
const point: Point = { x: 1, y: 2 };
    `},
		{Code: `
const point = { x: 1, y: 2 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Partial<Point> = { x: 1 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Pick<Point, 'x'> = { x: 1 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point | { x: number; y: number } = { x: 1, y: 2 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point = { x: 1, y: 2 } as unknown as Point;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: typeof Point = class {
  constructor(
    public x: number,
    public y: number,
  ) {}
};
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const points: Point[] = [new Point(1, 2)];
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function draw(point: Point): void;
draw(new Point(1, 2));
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function make(): Point {
  return new Point(1, 2);
}
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const shape: { origin: Point } = { origin: new Point(1, 2) };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = { ...new Point(1, 2) } as any;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point | object = { x: 1, y: 2 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point | Record<string, number> = { x: 1, y: 2 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = { x: 1, y: 2 } as any;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Point3 extends Point {
  z = 0;
}
const point: Point = new Point3(1, 2);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
interface Fancy extends Point {
  z: number;
}
declare const fancy: Fancy;
const point: Point = fancy;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function get<T extends Point>(value: T) {
  const point: Point = value;
}
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = Object.assign(new Point(1, 2), { x: 3 });
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const draw: (point: Point) => void = (point: { x: number; y: number }) => {};
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Promise<Point> = Promise.resolve(new Point(1, 2));
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Promise<Point | null> = Promise.resolve(null);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const points: Point[] = [new Point(1, 2)].map(point => point);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const points: Set<Point>;
const array: Point[] = [...points];
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const maybe: Point | undefined;
const point: Point = maybe!;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const PointClass: typeof Point = Point;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
type Constructor<T = {}> = new (...args: any[]) => T;
function Tagged<TBase extends Constructor>(Base: TBase) {
  return class extends Base {
    tag = '';
  };
}
const TaggedPoint = Tagged(Point);
const point: Point = new TaggedPoint(1, 2);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const value: any;
const point: Point = value;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const readonlyPoint: Readonly<Point>;
const point: Point = readonlyPoint;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Required<Point> = new Point(1, 2);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
type Shape = Point | { kind: 'plain'; x: number };
const shape: Shape = { kind: 'plain', x: 1 };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const points: Point[];
const readonlyPoints: readonly Point[] = points;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const points: Record<string, Point>;
const record: Record<string, Point> = { origin: new Point(0, 0), ...points };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const points: { [key: string]: Point };
const record: Record<string, Point> = points;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const record: Record<number, Point> = { 0: new Point(0, 0) };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const holder: { point: Point } = {
  get point() {
    return new Point(1, 2);
  },
};
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Readonly<Point> = Object.freeze(new Point(1, 2));
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const pair: [Point, number] = [new Point(1, 2), 1];
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Point3 extends Point {}
declare const union: Point | Point3;
const point: Point = union;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function unwrap<T>(value: Readonly<T>): T {
  return value as T;
}
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function Marker(props: { children: Point }): null;
const marker = <Marker>{new Point(1, 2)}</Marker>;
      `, Tsx: true, Files: reactTypesFiles},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point: Readonly<Point> = plain;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point: Readonly<Point> = Object.freeze(plain);
    `},
		{Code: `
class Alpha {
  type = 'alpha' as const;
}
class Beta {
  type = 'beta' as const;
}
declare function useBeta(beta: Beta): void;
declare const definition: (Alpha | Beta) & { type: 'beta' };
useBeta(definition);
    `},
		{Code: `
class Alpha {
  alpha = 1;
}
class Beta {
  beta = 1;
}
declare const scope: Alpha | Beta;
const beta = scope as Beta;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function wrap<T>(value: T) {
  const promise: Promise<T> = Promise.resolve(value);
}
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Point3 extends Point {}
const point: Point | Point3 = new Point3(1, 2);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const shorthand = { point = new Point(1, 2) };
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function draw(...points: Point[]): void;
declare const points: Point[];
draw(...points);
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
let count = 0;
count += 1;
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function noop(): void {
  return;
}
    `},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function* points(): Generator<Point | undefined> {
  yield;
}
    `},
		{Code: `
declare function Marker(props: { children?: Point }): null;
const marker = <Marker>{/* empty */}</Marker>;
      `, Tsx: true, Files: reactTypesFiles},
	}, []rule_tester.InvalidTestCase{
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 36, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  x: number = Math.random();
  y: number = Math.random();
}
const o = { x: 1, y: 1 };
const p: Point = o;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 18, EndColumn: 19, EndLine: 7, Line: 7, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  x: number = Math.random();
  y: number = Math.random();
}
function foo(o: { x: number; y: number }): Point {
  return o;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 11, EndLine: 7, Line: 7, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  x: number = Math.random();
  y: number = Math.random();
}
class Point2 {
  x: number = Math.random();
  y: number = Math.random();
}
function bar(p2: Point2): Point {
  return p2;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 12, EndLine: 11, Line: 11, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  public x: number;
  public y: number;

  constructor(x: number, y: number) {
    if (x < 0 || y < 0) {
      throw new Error('Point must have positive values');
    }

    this.x = x;
    this.y = y;
  }
}
const point: Point = { x: -1, y: -2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 38, EndLine: 15, Line: 15, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point = { x: 1, y: 2 } as Point;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 29, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point = <Point>{ x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 36, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point = { x: 1, y: 2 } satisfies Point;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 29, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function draw(point: Point): void;
draw({ x: 1, y: 2 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 6, EndColumn: 20, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function make(): Point {
  return { x: 1, y: 2 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 24, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const make = (): Point => ({ x: 1, y: 2 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 28, EndColumn: 42, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const points: Point[] = [{ x: 1, y: 2 }];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 40, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const shape: { origin: Point } = { origin: { x: 1, y: 2 } };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 44, EndColumn: 58, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
let point: Point;
point = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 23, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point | undefined = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 34, EndColumn: 48, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point & { z: number } = { x: 1, y: 2, z: 3 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 38, EndColumn: 58, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Label {
  constructor(public text: string) {}
}
const value: Label | Point = { text: 'origin' };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 30, EndColumn: 48, EndLine: 11, Line: 11, MessageId: "objectLiteral"}}},
		{Code: `
class Box<T> {
  constructor(public value: T) {}
}
const box: Box<number> = { value: 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 38, EndLine: 5, Line: 5, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function identity<T>(value: T): T;
identity<Point>({ x: 1, y: 2 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 17, EndColumn: 31, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
abstract class Shape {
  abstract area(): number;
}
const shape: Shape = { area: () => 1 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 39, EndLine: 5, Line: 5, MessageId: "objectLiteral"}}},
		{Code: `
declare class Point {
  x: number;
  y: number;
}
const point: Point = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 36, EndLine: 6, Line: 6, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = { ...new Point(1, 2) };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 44, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const condition: boolean;
const point: Point = condition ? { x: 1, y: 2 } : new Point(3, 4);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 34, EndColumn: 48, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Line {
  start: Point = { x: 0, y: 0 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 18, EndColumn: 32, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function draw(point: Point = { x: 0, y: 0 }) {}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 30, EndColumn: 44, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function Marker(props: { point: Point }): null;
const marker = <Marker point={{ x: 1, y: 2 }} />;
      `, Tsx: true, Files: reactTypesFiles, Errors: []rule_tester.InvalidTestCaseError{{Column: 31, EndColumn: 45, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: InstanceType<typeof Point> = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 43, EndColumn: 57, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const origin = new Point(0, 0);
const point: typeof origin = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 30, EndColumn: 44, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
async function make(): Promise<Point> {
  return { x: 1, y: 2 };
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 24, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const value: { point?: Point };
const { point = { x: 0, y: 0 } } = value;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 17, EndColumn: 31, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Holder {
  get point(): Point {
    return { x: 0, y: 0 };
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 26, EndLine: 10, Line: 10, MessageId: "objectLiteral"}}},
		{Code: `
const Point = class {
  constructor(
    public x: number,
    public y: number,
  ) {}
};
const point: InstanceType<typeof Point> = { x: 1, y: 2 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 43, EndColumn: 57, EndLine: 8, Line: 8, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point: Point = plain;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 27, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Promise<Point> = Promise.resolve({ x: 1, y: 2 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 31, EndColumn: 62, EndLine: 8, Line: 8, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const points: Point[] = [1].map(x => ({ x, y: 1 }));
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 25, EndColumn: 52, EndLine: 8, Line: 8, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function identity<T>(value: T): T;
const point: Point = identity({ x: 1, y: 2 });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 46, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const points: Map<string, Point> = new Map([['origin', plain]]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 36, EndColumn: 64, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const points: [Point, number] = [plain, 1];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 34, EndColumn: 39, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point: Point = { ...plain };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 34, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const shape: { origin: Point } = { origin: plain };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 44, EndColumn: 49, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare function getPlain(): Promise<{ x: number; y: number }>;
async function get() {
  const point: Point = await getPlain();
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 24, EndColumn: 40, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const fallback: Point;
const point: Point = plain ?? fallback;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 27, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point = Promise.resolve(plain) satisfies Promise<Point>;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 15, EndColumn: 37, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare function draw(...points: Point[]): void;
draw(new Point(1, 2), plain);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 23, EndColumn: 28, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
class Vector {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
const point: Point = new Vector(1, 2);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 38, EndLine: 14, Line: 14, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Box<T> {
  constructor(public value: T) {}
}
const box: Box<Point> = new Box(plain);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 25, EndColumn: 39, EndLine: 12, Line: 12, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
interface Fancy extends Point {
  z: number;
}
const fancy: Fancy = { x: 1, y: 2, z: 3 };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 42, EndLine: 11, Line: 11, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
function getPoints(): Point[] {
  return [plain].filter(Boolean);
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 33, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Line {
  constructor(
    public start: Point,
    public end: Point,
  ) {}
}
new Line(plain, new Point(1, 2));
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 15, EndLine: 15, Line: 15, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Base {
  constructor(public point: Point) {}
}
class Derived extends Base {
  constructor() {
    super(plain);
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 16, EndLine: 14, Line: 14, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const draw: ((point: Point) => void) | undefined;
draw?.(plain);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 8, EndColumn: 13, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
interface Repository {
  get(): Point;
}
const repository: Repository = {
  get() {
    return { x: 1, y: 2 };
  },
};
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 26, EndLine: 13, Line: 13, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Repository {
  get(): Point {
    return plain;
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 12, EndColumn: 17, EndLine: 11, Line: 11, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
function* points(): Generator<Point> {
  yield plain;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 9, EndColumn: 14, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const make: () => Point = () => {
  return plain;
};
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 15, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const draw: (point?: Point) => void = (point = plain) => {};
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 48, EndColumn: 53, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Holder {
  constructor(public point: Point = plain) {}
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 37, EndColumn: 42, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Holder {
  accessor point: Point = plain;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 27, EndColumn: 32, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
let point: Point;
({ point } = { point: plain });
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 23, EndColumn: 28, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
class Holder {
  set origin(value: Point) {}
}
new Holder().origin = plain;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 23, EndColumn: 28, EndLine: 12, Line: 12, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const points: Point[];
points[0] = plain;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 13, EndColumn: 18, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
let point: Point | undefined;
point ??= plain;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 11, EndColumn: 16, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point: Point = (0, plain);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 26, EndColumn: 31, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const condition: boolean;
const point: Point = condition ? plain || new Point(1, 2) : new Point(3, 4);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 34, EndColumn: 39, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const key: string;
const record: Record<string, Point> = { [key]: plain };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 48, EndColumn: 53, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const grid: Point[][] = [[plain]];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 27, EndColumn: 32, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
function make<T extends Point>(): T {
  return { x: 1, y: 2 } as T;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 10, EndColumn: 24, EndLine: 9, Line: 9, MessageId: "objectLiteral"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const maybe: { x: number; y: number } | undefined;
const point: Point = maybe!;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 28, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const point = new Promise<Point>(resolve => resolve(plain));
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 53, EndColumn: 58, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
interface Shape {
  origin: Point;
}
const point: Shape['origin'] = plain;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 32, EndColumn: 37, EndLine: 12, Line: 12, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plains: { x: number; y: number }[];
const points: readonly Point[] = plains;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 34, EndColumn: 40, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plains: Record<string, { x: number; y: number }>;
const points: Record<string, Point> = plains;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 39, EndColumn: 45, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const points: Map<string, Point[]> = new Map([['origin', [plain]]]);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 38, EndColumn: 68, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const union: Point | { x: number; y: number };
const point: Point = union;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 27, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const intersection: { x: number } & { y: number };
const point: Point = intersection;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 22, EndColumn: 34, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const points: (Point | null)[] = [plain, null];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 35, EndColumn: 40, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare const condition: boolean;
const points: Point[] = [condition ? plain : new Point(1, 2)];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 38, EndColumn: 43, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare function draw(point: Point): void;
draw(plain as Point);
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 6, EndColumn: 11, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const points: [string, ...Point[]] = ['origin', plain];
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 49, EndColumn: 54, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
const shape: { origin?: Point } = { origin: plain };
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 45, EndColumn: 50, EndLine: 9, Line: 9, MessageId: "nonInstance"}}},
		{Code: `
class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}
declare const plain: { x: number; y: number };
declare function Marker(props: { children: Point }): null;
const marker = <Marker>{plain}</Marker>;
      `, Tsx: true, Files: reactTypesFiles, Errors: []rule_tester.InvalidTestCaseError{{Column: 25, EndColumn: 30, EndLine: 10, Line: 10, MessageId: "nonInstance"}}},
		{Code: `
class Beta {
  beta = 1;
}
declare const value: string | { beta: number };
const beta = value as Beta;
      `, Errors: []rule_tester.InvalidTestCaseError{{Column: 14, EndColumn: 19, EndLine: 6, Line: 6, MessageId: "nonInstance"}}},
	})
}
