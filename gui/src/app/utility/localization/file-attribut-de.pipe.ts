import { Pipe, PipeTransform } from '@angular/core';

@Pipe({name: 'fileFeature'})
export class FileFeaturePipe implements PipeTransform {
  transform(value: string): string {
    switch (value) {
      case 'fileName': {
        return 'Dateiname';
      }
      case 'relativePath': {
        return 'Pfad';
      }
      case 'fileSize': {
        return 'Dateigröße';
      }
      case 'formatVersion': {
        return 'Formatversion';
      }
      case 'mimeType': {
        return 'MIME-Type';
      }
      case 'puid': {
        return 'PUID';
      }
      case 'valid': {
        return 'Valide';
      }
      case 'wellFormed': {
        return 'Wohlgeformt';
      }
      case 'encoding': {
        return 'Zeichenkodierung';
      }
    }
    return value;
  }
}