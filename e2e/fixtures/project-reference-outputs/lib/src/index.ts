import type { Value } from '../../core/src/index';
import { createValue } from '../../core/src/index';

export async function getValue(): Promise<Value> {
  return createValue();
}
