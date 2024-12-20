import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { first, map, shareReplay, switchMap } from 'rxjs/operators';
import { Agency } from './agencies.service';
import { ProcessingError } from './clearing.service';
import { ItemProgress, TaskState } from './tasks.service';
import { UpdatesService } from './updates.service';

export interface ProcessData {
  process: SubmissionProcess;
  warnings: Warning[];
  processingErrors: ProcessingError[];
}

export interface SubmissionProcess {
  processId: string;
  createdAt: string;
  agency: Agency;
  note: string;
  processState: ProcessState;
  unresolvedErrors: number;
}

export interface Warning {
  createdAt: string;
  title: string;
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
  progress?: ItemProgress;
  taskId: string;
  taskState: TaskState;
  hasError: boolean;
}

@Injectable({
  providedIn: 'root',
})
export class ProcessService {
  private httpClient = inject(HttpClient);
  private updates = inject(UpdatesService);

  private cachedProcessId?: string;
  private cachedProcessData?: Observable<ProcessData>;

  getProcesses(allUsers: boolean) {
    if (allUsers) {
      return this.httpClient.get<SubmissionProcess[]>('/api/processes');
    } else {
      return this.httpClient.get<SubmissionProcess[]>('/api/processes/my');
    }
  }

  observeProcessData(id: string): Observable<ProcessData> {
    if (id !== this.cachedProcessId) {
      this.cachedProcessId = id;
      this.cachedProcessData = this.updates.observeSubmissionProcess(id).pipe(
        switchMap(() => this.httpClient.get<ProcessData>('/api/process/' + id)),
        shareReplay({ bufferSize: 1, refCount: true }),
      );
    }
    return this.cachedProcessData!;
  }

  getProcessData(id: string): Observable<ProcessData> {
    return this.observeProcessData(id).pipe(first());
  }

  setNote(processId: string, note: string): Observable<void> {
    return this.httpClient.patch('/api/process-note/' + processId, note).pipe(map(() => void 0));
  }

  getAppraisalReport(processId: string): Observable<Blob> {
    return this.httpClient.get('/api/report/appraisal/' + processId, {
      responseType: 'blob',
    });
  }

  getSubmissionReport(processId: string): Observable<Blob> {
    return this.httpClient.get('/api/report/submission/' + processId, {
      responseType: 'blob',
    });
  }

  deleteProcess(processId: string): Observable<void> {
    return this.httpClient.delete('/api/process/' + processId).pipe(map(() => void 0));
  }
}
