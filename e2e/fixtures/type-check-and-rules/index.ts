// A type error and two type-aware lint rule violations in the same file.
const x: string = 123;

function takesString(_value: string) {}

takesString('' as any);

async function run(): Promise<void> {}

run();
