class Point {
  constructor(
    public x: number,
    public y: number,
  ) {}
}

const point: Point = { x: -1, y: -2 };

declare const plain: { x: number; y: number };
const points: Point[] = [plain];

const validPoint = new Point(1, 2);
const validPoints: Point[] = [new Point(1, 2)];

export {};
