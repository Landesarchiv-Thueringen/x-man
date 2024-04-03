export function notNull<T>(value: T | null | void): value is T {
  return value != null;
}
