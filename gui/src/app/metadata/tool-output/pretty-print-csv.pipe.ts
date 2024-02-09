import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'prettyPrintCsv',
  standalone: true,
})
export class PrettyPrintCsvPipe implements PipeTransform {
  transform(csvString: string): string {
    const rows = csvString.split('\n').map((row) => row.split(',').map((cell) => cell.trim()));
    const nColumns = Math.max(...rows.map((row) => row.length));
    for (let i = 0; i < nColumns; i++) {
      const columnWidth = Math.max(...rows.map((row) => row[i]?.length ?? 0));
      rows.forEach((row) => {
        if (i < row.length - 1) {
          row[i] = padToLength(row[i] ?? '', columnWidth);
        }
      });
    }
    return rows.map((row) => row.join(', ')).join('\n');
  }
}

function padToLength(s: string, length: number): string {
  return s + ' '.repeat(length - s.length);
}
