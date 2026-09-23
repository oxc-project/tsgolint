type Data = { name: string };
type Empty = Omit<Data, 'name'>;

type Keys<T> = T extends infer U ? keyof U : never;
type Mapped<T extends object> = { [Key in Keys<T>]: Key };
type Referenced<T extends object> = Mapped<T>;
