class ValueObject {
  constructor(private readonly value: number) {}
}

class OtherValueObject {
  constructor(private readonly value: number) {}
}

const valueA = new ValueObject(10);
const valueB = new ValueObject(11);

const otherValueA = new OtherValueObject(20);
const otherValueB = new OtherValueObject(21);

const maybeValue: ValueObject | null | undefined =
  Math.random() > 0.66 ? new ValueObject(1) : Math.random() > 0.5 ? null : undefined;

const valueLt = valueA < valueB;
const valueLte = valueA <= valueB;
const valueGt = valueA > valueB;
const valueGte = valueA >= valueB;
const valueEq = valueA == valueB;
const valueStrictEq = valueA === valueB;
const valueNe = valueA != valueB;
const valueStrictNe = valueA !== valueB;

const otherValueLt = otherValueA < otherValueB;
const otherValueLte = otherValueA <= otherValueB;
const otherValueGt = otherValueA > otherValueB;
const otherValueGte = otherValueA >= otherValueB;
const otherValueEq = otherValueA == otherValueB;
const otherValueStrictEq = otherValueA === otherValueB;
const otherValueNe = otherValueA != otherValueB;
const otherValueStrictNe = otherValueA !== otherValueB;

const maybeUndefined = maybeValue !== undefined;
const maybeNull = maybeValue !== null;
