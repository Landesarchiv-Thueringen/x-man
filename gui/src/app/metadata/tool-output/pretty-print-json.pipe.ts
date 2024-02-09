import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'prettyPrintJson',
  standalone: true,
})
export class PrettyPrintJsonPipe implements PipeTransform {
  transform(jsonString: string): string {
    return JSON.stringify(JSON.parse(jsonString), null, 2);
  }
}
