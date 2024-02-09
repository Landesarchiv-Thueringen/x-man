import { Pipe, PipeTransform } from '@angular/core';

const labelMap: { [key: string]: string | undefined } = {
  tool: 'Werkzeug',
  fileName: 'Dateiname',
  relativePath: 'Pfad',
  fileSize: 'Dateigröße',
  formatVersion: 'Formatversion',
  mimeType: 'MIME-Type',
  puid: 'PUID',
  valid: 'Valide',
  wellFormed: 'Wohlgeformt',
  encoding: 'Zeichenkodierung',
  error: 'Fehler',
};

@Pipe({ name: 'fileFeature', standalone: true })
export class FileFeaturePipe implements PipeTransform {
  transform(value: string): string {
    const label = labelMap[value];
    return label ?? value;
  }
}
