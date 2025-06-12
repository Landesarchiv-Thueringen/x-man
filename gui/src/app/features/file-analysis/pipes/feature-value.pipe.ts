import { Pipe, PipeTransform } from '@angular/core';
import { FeatureValue } from '../results';

@Pipe({
  name: 'featureValue',
  standalone: true,
})
export class FeatureValuePipe implements PipeTransform {
  transform(value: FeatureValue | undefined): string | number | boolean | undefined {
    return value ? value.value : undefined;
  }
}
