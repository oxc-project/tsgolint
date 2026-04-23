import { styleText } from 'node:util';

export function accent(text: string) {
  return styleText('blue', text);
}
