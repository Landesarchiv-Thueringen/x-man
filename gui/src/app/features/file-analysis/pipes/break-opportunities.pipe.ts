import { Pipe, PipeTransform } from '@angular/core';

/**
 * Inserts line break opportunities after characters that are commonly used for
 * separation in strings like filenames or MIME types.
 */
@Pipe({
  name: 'breakOpportunities',
  standalone: true,
})
export class BreakOpportunitiesPipe implements PipeTransform {
  transform(value: string): string {
    return value.replaceAll('_', '_<wbr>').replaceAll('/', '/<wbr>');
  }
}
