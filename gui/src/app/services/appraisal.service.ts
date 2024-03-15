import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export type AppraisalDecision = 'A' | 'V' | 'B' | '';

export interface Appraisal {
  recordObjectID: string;
  decision: AppraisalDecision;
  internalNote: string;
}

/**
 * Provides API functions for fetching and updating appraisals.
 *
 * Components should not use this service directly, but use the wrapped
 * functions provided by `MessagePageService`, which provides an updated
 * observable for the page's process.
 */
@Injectable({
  providedIn: 'root',
})
export class AppraisalService {
  constructor(private httpClient: HttpClient) {}

  getAppraisals(processId: string): Observable<Appraisal[]> {
    return this.httpClient.get<Appraisal[]>(environment.endpoint + '/appraisals/' + processId);
  }

  setDecision(processId: string, recordObjectId: string, decision: AppraisalDecision): Observable<Appraisal[]> {
    return this.httpClient.post<Appraisal[]>(environment.endpoint + '/appraisal-decision', decision, {
      params: { processId, recordObjectId },
    });
  }

  setInternalNote(processId: string, recordObjectId: string, internalNote: string): Observable<Appraisal[]> {
    return this.httpClient.post<Appraisal[]>(environment.endpoint + '/appraisal-note', internalNote, {
      params: { processId, recordObjectId },
    });
  }

  setAppraisals(
    processId: string,
    recordObjectIds: string[],
    decision: AppraisalDecision,
    internalNote: string,
  ): Observable<Appraisal[]> {
    return this.httpClient.post<Appraisal[]>(environment.endpoint + '/appraisals', {
      processId,
      recordObjectIds,
      decision,
      internalNote,
    });
  }
}
