import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export type AppraisalCode = 'A' | 'V' | 'B' | '';

export interface Appraisal {
  recordId: string;
  decision: AppraisalCode;
  note: string;
}

export interface AppraisalDescription {
  shortDesc: string;
  desc: string;
}

export const appraisalDescriptions = {
  A: { shortDesc: 'Archivieren', desc: 'Das Schriftgutobjekt ist archivwürdig.' },
  B: { shortDesc: 'Durchsicht', desc: 'Das Schriftgutobjekt ist zum Bewerten markiert.' },
  V: { shortDesc: 'Vernichten', desc: 'Das Schriftgutobjekt ist zum Vernichten markiert.' },
} as const;

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

  getAppraisalDescription(code?: AppraisalCode): AppraisalDescription | undefined {
    if (!code) {
      return undefined;
    }
    return appraisalDescriptions[code];
  }

  getAppraisals(processId: string): Observable<Appraisal[]> {
    return this.httpClient.get<Appraisal[]>(environment.endpoint + '/appraisals/' + processId);
  }

  setDecision(processId: string, recordId: string, decision: AppraisalCode): Observable<Appraisal[]> {
    return this.httpClient.post<Appraisal[]>(environment.endpoint + '/appraisal-decision', decision, {
      params: { processId, recordId },
    });
  }

  setInternalNote(processId: string, recordId: string, internalNote: string): Observable<Appraisal[]> {
    return this.httpClient.post<Appraisal[]>(environment.endpoint + '/appraisal-note', internalNote, {
      params: { processId, recordId },
    });
  }

  setAppraisals(
    processId: string,
    recordObjectIds: string[],
    decision: AppraisalCode,
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
