import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export const packagingOptions: { value: PackagingOption; label: string }[] = [
  { value: '', label: 'keine Unterpaketierung' },
  { value: 'sub-file', label: 'Teilakte' },
  { value: 'process', label: 'Vorgang' },
];

export type PackagingOption = '' | 'sub-file' | 'process';
export type PackagingDecision = '' | 'single' | 'sub';
export type PackagingDecisions = { [recordId: string]: PackagingDecision };

export interface RecordOption {
  recordId: string;
  packaging: PackagingOption;
}

@Injectable({
  providedIn: 'root',
})
export class RecordOptionsService {
  constructor(private httpClient: HttpClient) {}

  getRecordOptions(processId: string): Observable<RecordOption[]> {
    return this.httpClient.get<RecordOption[]>(
      environment.endpoint + '/record-options/' + processId,
    );
  }

  getPackaging(processId: string): Observable<PackagingDecisions> {
    return this.httpClient.get<PackagingDecisions>(
      environment.endpoint + '/packaging/' + processId,
    );
  }

  setPackaging(
    processId: string,
    recordIds: string[],
    packaging: PackagingOption,
  ): Observable<{
    recordOptions: RecordOption[];
    packaging: PackagingDecisions;
  }> {
    return this.httpClient.post<{
      recordOptions: RecordOption[];
      packaging: PackagingDecisions;
    }>(environment.endpoint + '/packaging', {
      processId,
      recordIds,
      packaging,
    });
  }
}
