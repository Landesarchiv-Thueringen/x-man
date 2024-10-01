import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

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
  constructor(private httpClient: HttpClient) {}

  getPackaging(processId: string): Observable<PackagingData> {
    return this.httpClient.get<PackagingData>(environment.endpoint + '/packaging/' + processId);
  }

  setPackagingChoice(
    processId: string,
    recordIds: string[],
    packagingChoice: PackagingChoice,
  ): Observable<PackagingData> {
    return this.httpClient.post<PackagingData>(environment.endpoint + '/packaging', {
      processId,
      recordIds,
      packagingChoice,
    });
  }

  getPackagingStats(processId: string, rootRecords: string[]): Observable<PackagingStatsMap> {
    return this.httpClient.post<PackagingStatsMap>(
      environment.endpoint + '/packaging-stats/' + processId,
      rootRecords,
    );
  }
}
