import { Pipe, PipeTransform } from '@angular/core';

@Pipe({name: 'featureBreak'})
export class FeatureBreakPipe implements PipeTransform {
  transform(value: string): string {
    let valueWithBreakOppertunities = value.replaceAll('_', '_<wbr>');
    valueWithBreakOppertunities = valueWithBreakOppertunities.replaceAll('/', '/<wbr>');
    valueWithBreakOppertunities = valueWithBreakOppertunities.replaceAll('.', '.<wbr>');
    return valueWithBreakOppertunities;
  }
}