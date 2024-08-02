import { Pipe, PipeTransform } from '@angular/core';
import { PackagingStats } from '../../services/packaging.service';

@Pipe({
  name: 'packagingStats',
  standalone: true,
})
export class PackagingStatsPipe implements PipeTransform {
  transform(value: PackagingStats): string {
    return printPackagingStats(value);
  }
}

export function printPackagingStats(stats: PackagingStats): string {
  const result: string[] = [];
  if (stats.files === 1) {
    result.push('1 Akte');
  } else if (stats.files > 1) {
    result.push(`${stats.files} Akten`);
  }
  if (stats.subfiles === 1) {
    result.push('1 Teilakte');
  } else if (stats.subfiles > 1) {
    result.push(`${stats.subfiles} Teilakten`);
  }
  if (stats.processes === 1) {
    result.push('1 Vorgang');
  } else if (stats.processes > 1) {
    result.push(`${stats.processes} Vorgänge`);
  }
  if (stats.other === 1) {
    result.push('1 Sammelpaket');
  } else if (stats.other > 1) {
    result.push(`${stats.other} Sammelpakete`);
  }
  return result.join(', ');
}
