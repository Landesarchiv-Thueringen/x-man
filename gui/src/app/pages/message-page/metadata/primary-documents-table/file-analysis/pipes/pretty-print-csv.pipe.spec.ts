import { PrettyPrintCsvPipe } from './pretty-print-csv.pipe';

describe('PrettyPrintCsvPipe', () => {
  const pipe = new PrettyPrintCsvPipe();

  it('pretty prints a complete CSV', () => {
    expect(
      pipe.transform(
        [
          'foo,bar,baz', //
          'foooooooooo,b,bazbazbaz',
          'f,b,b',
        ].join('\n'),
      ),
    ).toBe(
      [
        'foo        , bar, baz', //
        'foooooooooo, b  , bazbazbaz',
        'f          , b  , b',
      ].join('\n'),
    );
  });

  it('pretty prints an incomplete CSV', () => {
    expect(
      pipe.transform(
        [
          'foo,bar,baz', //
          'foooooooooo',
          'f,b,b',
        ].join('\n'),
      ),
    ).toBe(
      [
        'foo        , bar, baz', //
        'foooooooooo',
        'f          , b  , b',
      ].join('\n'),
    );
  });
});
