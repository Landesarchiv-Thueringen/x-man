import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { firstValueFrom, Observable } from 'rxjs';

export const packagingChoices: { value: PackagingChoice; label: string; disabled?: boolean }[] = [
  { value: 'root', label: 'Wurzelebene' },
  { value: 'level-1', label: '1. Unterebene' },
  { value: 'level-2', label: '2. Unterebene' },
];

export type PackagingChoice = 'root' | 'level-1' | 'level-2';
export type PackagingDecision = '' | 'single' | 'sub';
export interface PackagingStats {
  files: number;
  subfiles: number;
  processes: number;
  other: number;
  deepestLevelHasItems: boolean;
}

export interface PackagingData {
  choices: { [recordId in string]?: PackagingChoice };
  decisions: { [recordId in string]?: PackagingDecision };
  stats: { [recordId in string]?: PackagingStats };
}

export type PackagingStatsMap = { [option in PackagingChoice]: PackagingStats };

@Injectable({
  providedIn: 'root',
})
export class PackagingService {
  private httpClient = inject(HttpClient);

  getPackaging(processId: string): Observable<PackagingData> {
    return this.httpClient.get<PackagingData>('/api/packaging/' + processId);
  }

  setPackagingChoice(
    processId: string,
    recordIds: string[],
    packagingChoice: PackagingChoice,
  ): Observable<PackagingData> {
    return this.httpClient.post<PackagingData>('/api/packaging', {
      processId,
      recordIds,
      packagingChoice,
    });
  }

  getPackagingStats(processId: string, rootRecords: string[]): Promise<PackagingStatsMap> {
    return firstValueFrom(
      this.httpClient.post<PackagingStatsMap>('/api/packaging-stats/' + processId, rootRecords),
    );
  }
}
