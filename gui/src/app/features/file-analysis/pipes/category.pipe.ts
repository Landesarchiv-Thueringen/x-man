import { Pipe, PipeTransform } from '@angular/core';

const labelMap: { [key in string]?: string } = {
  audio: 'Audio',
  av_container: 'Containerformat',
  format: 'Dateiformat',
  general: 'Allgemein',
  text: 'Text',
  video: 'Video',
};

@Pipe({
  name: 'category',
  standalone: true,
})
export class CategoryPipe implements PipeTransform {
  transform(value: string): string {
    const label = labelMap[value];
    return label ?? value;
  }
}
