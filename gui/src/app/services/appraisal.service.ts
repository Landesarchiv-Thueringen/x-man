import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { environment } from '../../environments/environment';

export type AppraisalDecision = 'A' | 'V' | 'B' | '';

export interface Appraisal {
  recordObjectID: string;
  decision: AppraisalDecision;
  internalNote: string;
}

@Injectable({
  providedIn: 'root',
})
export class AppraisalService {
  private currentProcessId: string | null = null;
  private currentAppraisals = new BehaviorSubject<Appraisal[]>([]);

  constructor(private httpClient: HttpClient) {}

  observeAppraisals(processId: string): Observable<Appraisal[]> {
    if (processId != this.currentProcessId) {
      const appraisalsSubject = this.setProcessId(processId);
      this.httpClient
        .get<Appraisal[]>(environment.endpoint + '/appraisals/' + processId)
        .subscribe((appraisals) => appraisalsSubject.next(appraisals));
    }
    return this.currentAppraisals;
  }

  setDecision(processId: string, recordObjectId: string, decision: AppraisalDecision) {
    const appraisalsSubject = this.setProcessId(processId);
    this.httpClient
      .post<
        Appraisal[]
      >(environment.endpoint + '/appraisal-decision', decision, { params: { processId, recordObjectId } })
      .subscribe((appraisals) => appraisalsSubject.next(appraisals));
  }

  setInternalNote(processId: string, recordObjectId: string, internalNote: string) {
    const appraisalsSubject = this.setProcessId(processId);
    this.httpClient
      .post<
        Appraisal[]
      >(environment.endpoint + '/appraisal-note', internalNote, { params: { processId, recordObjectId } })
      .subscribe((appraisals) => appraisalsSubject.next(appraisals));
  }

  setAppraisals(processId: string, recordObjectIds: string[], decision: AppraisalDecision, internalNote: string) {
    const appraisalsSubject = this.setProcessId(processId);
    return this.httpClient
      .post<Appraisal[]>(environment.endpoint + '/appraisals', {
        processId,
        recordObjectIds,
        decision,
        internalNote,
      })
      .pipe(tap((appraisals) => appraisalsSubject.next(appraisals)));
  }

  private setProcessId(processId: string): BehaviorSubject<Appraisal[]> {
    if (processId != this.currentProcessId) {
      this.currentAppraisals.complete();
      this.currentAppraisals = new BehaviorSubject<Appraisal[]>([]);
      this.currentProcessId = processId;
    }
    return this.currentAppraisals;
  }
}
