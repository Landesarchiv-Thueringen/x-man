import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, interval } from 'rxjs';
import { first, map, shareReplay, startWith, switchMap } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { Agency } from './agencies.service';
import { ProcessingError } from './clearing.service';

export interface ProcessData {
  process: SubmissionProcess;
  processingErrors: ProcessingError[];
}

export interface SubmissionProcess {
  processId: string;
  createdAt: string;
  agency: Agency;
  note: string;
  processState: ProcessState;
}

export interface ProcessState {
  receive0501: ProcessStep;
  appraisal: ProcessStep;
  receive0505: ProcessStep;
  receive0503: ProcessStep;
  formatVerification: ProcessStep;
  archiving: ProcessStep;
}

export interface ProcessStep {
  updatedAt: string;
  complete: boolean;
  completedAt: string;
  progress: string;
  running: boolean;
  unresolvedErrors: number;
}

@Injectable({
  providedIn: 'root',
})
export class ProcessService {
  private apiEndpoint: string;
  private cachedProcessId?: string;
  private cachedProcessData?: Observable<ProcessData>;

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  getProcesses(allUsers: boolean) {
    if (allUsers) {
      return this.httpClient.get<SubmissionProcess[]>(this.apiEndpoint + '/processes');
    } else {
      return this.httpClient.get<SubmissionProcess[]>(this.apiEndpoint + '/processes/my');
    }
  }

  observeProcessData(id: string): Observable<ProcessData> {
    if (id !== this.cachedProcessId) {
      this.cachedProcessId = id;
      this.cachedProcessData = interval(environment.updateInterval).pipe(
        startWith(void 0),
        switchMap(() => this.httpClient.get<ProcessData>(this.apiEndpoint + '/process/' + id)),
        shareReplay({ bufferSize: 1, refCount: true }),
      );
    }
    return this.cachedProcessData!;
  }

  getProcessData(id: string): Observable<ProcessData> {
    return this.observeProcessData(id).pipe(first());
  }

  setNote(processId: string, note: string): Observable<void> {
    return this.httpClient.patch(this.apiEndpoint + '/process-note/' + processId, note).pipe(map(() => void 0));
  }

  getReport(processId: string): Observable<Blob> {
    return this.httpClient.get(this.apiEndpoint + '/report/' + processId, { responseType: 'blob' });
  }

  deleteProcess(processId: string): Observable<void> {
    return this.httpClient.delete(this.apiEndpoint + '/process/' + processId).pipe(map(() => void 0));
  }
}
