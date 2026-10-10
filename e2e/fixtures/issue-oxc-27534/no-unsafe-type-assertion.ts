declare const input: { id: number | string; name: string };

// eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
export const user = {
  id: input.id,
  name: input.name,
} as { id: number; name: string };
