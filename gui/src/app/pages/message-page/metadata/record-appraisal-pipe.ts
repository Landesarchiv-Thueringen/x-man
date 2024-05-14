import { Pipe, PipeTransform } from '@angular/core';

@Pipe({ name: 'recordAppraisal', standalone: true })
export class RecordAppraisalPipe implements PipeTransform {
  transform(value: string): string {
    if (value === 'B') {
      // de: Durchsicht
      return 'D';
    }
    return value;
  }
}
