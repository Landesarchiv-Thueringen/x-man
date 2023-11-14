import { Pipe, PipeTransform } from '@angular/core';

@Pipe({name: 'recordObjectAppraisal'})
export class RecordObjectAppraisalPipe implements PipeTransform {
  transform(value: string): string {
    if (value === 'B') {
      // de: Durchsicht
      return 'D';
    }
    return value;
  }
}