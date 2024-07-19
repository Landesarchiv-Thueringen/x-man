import { Pipe, PipeTransform } from '@angular/core';

const labelMap: { [key in string]?: string } = {
  tool: 'Werkzeug',
  filename: 'Dateiname',
  formatName: 'Formatbezeichnung',
  formatVersion: 'Formatversion',
  mimeType: 'MIME-Type',
  puid: 'PUID',
  valid: 'Valide',
  wellFormed: 'Wohlgeformt',
  encoding: 'Zeichenkodierung',
  error: 'Fehler',
  status: 'Status',
};

@Pipe({
  name: 'fileFeature',
  standalone: true,
})
export class FileFeaturePipe implements PipeTransform {
  transform(value: string): string {
    const label = labelMap[value];
    return label ?? value;
  }
}
