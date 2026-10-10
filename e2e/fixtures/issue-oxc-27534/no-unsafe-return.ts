declare function parseUntrustedPayload(text: string): any;

export const parseDisplayName = (rawText: string): string =>
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  parseUntrustedPayload(rawText).displayName;
