import { Pipe, PipeTransform } from '@angular/core';
import { ConfidentialityLevel } from '../../../services/records.service';

export const confidentialityLevels = {
  '001': { shortDesc: 'Geheim', desc: 'Geheim: Das Schriftgutobjekt ist als geheim eingestuft.' },
  '002': { shortDesc: 'NfD', desc: 'NfD: Das Schriftgutobjekt ist als "nur für den Dienstgebrauch (nfD)" eingestuft.' },
  '003': { shortDesc: 'Offen', desc: 'Offen: Das Schriftgutobjekt ist nicht eingestuft.' },
  '004': { shortDesc: 'Streng geheim', desc: 'Streng geheim: Das Schriftgutobjekt ist als streng geheim eingestuft.' },
  '005': { shortDesc: 'Vertraulich', desc: 'Vertraulich: Das Schriftgutobjekt ist als vertraulich eingestuft.' },
} as const;

@Pipe({
  name: 'confidentialityLevel',
  standalone: true,
})
export class ConfidentialityLevelPipe implements PipeTransform {
  transform(value: ConfidentialityLevel | undefined, attribute: 'shortDesc' | 'desc'): string | null {
    if (value == null) {
      return null;
    }
    return confidentialityLevels[value][attribute];
  }
}
