export function notNull<T>(value: T | null | undefined | void): value is T {
  return value != null;
}

export function notEmpty<T>(value: T | null | undefined | void | ''): value is T {
  return !!value;
}
