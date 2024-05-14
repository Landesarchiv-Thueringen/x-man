import { Pipe, PipeTransform } from '@angular/core';
import { Medium } from '../../../services/records.service';

export const media = {
  '001': {
    shortDesc: 'Elektronisch',
    desc: 'Elektronisch: Das Schriftgutobjekt liegt ausschließlich in elektronischer Form vor.',
  },
  '002': {
    shortDesc: 'Hybrid',
    desc: 'Hybrid: Das Schriftgutobjekt liegt teilweise in elektronischer Form und teilweise als Papier vor.',
  },
  '003': { shortDesc: 'Papier', desc: 'Papier: Das Schriftgutobjekt liegt ausschließlich als Papier vor.' },
} as const;

@Pipe({
  name: 'medium',
  standalone: true,
})
export class MediumPipe implements PipeTransform {
  transform(value: Medium | undefined, attribute: 'shortDesc' | 'desc'): string | null {
    if (value == null) {
      return null;
    }
    return media[value][attribute];
  }
}
