declare const names: string[];

// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
export const firstLongName = names.filter(
  (name) => name.length > "a typical display name".length,
)?.[0];
