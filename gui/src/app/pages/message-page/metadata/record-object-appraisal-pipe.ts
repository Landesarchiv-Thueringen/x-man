import { Pipe, PipeTransform } from '@angular/core';

@Pipe({ name: 'recordObjectAppraisal', standalone: true })
export class RecordObjectAppraisalPipe implements PipeTransform {
  transform(value: string): string {
    if (value === 'B') {
      // de: Durchsicht
      return 'D';
    }
    return value;
  }
}
