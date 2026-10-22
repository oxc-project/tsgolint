enum Fruit {
  Apple,
  Banana,
}

const fruit: Fruit = 1;
const fruits: Fruit[] = [0, Fruit.Apple];

function takesFruit(value: Fruit): void {}
takesFruit(1);

function getFruit(): Fruit {
  return 0;
}

const asserted = 1 as Fruit;
let mutable = Fruit.Apple;
mutable++;

declare const names: { [key in Fruit]: string };
names[0];
