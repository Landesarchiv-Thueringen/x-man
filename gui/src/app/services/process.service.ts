import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, interval } from 'rxjs';
import { first, map, shareReplay, startWith, switchMap } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { Message } from '../services/message.service';
import { Agency } from './agencies.service';
import { ProcessingError } from './clearing.service';
import { Task } from './tasks.service';

export interface Process {
  agency: Agency;
  id: string;
  receivedAt: string;
  institution: string;
  note: string;
  message0501Id: string;
  message0501: Message;
  message0503Id: string;
  message0503: Message;
  processingErrors: ProcessingError[];
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
  complete: boolean;
  completionTime?: string;
  updateTime: string;
  message?: string;
  tasks: Task[];
  processingErrors: ProcessingError[];
}

@Injectable({
  providedIn: 'root',
})
export class ProcessService {
  private apiEndpoint: string;
  private cachedProcessId?: string;
  private cachedProcess?: Observable<Process>;

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  getProcesses(allUsers: boolean) {
    if (allUsers) {
      return this.httpClient.get<Process[]>(this.apiEndpoint + '/processes');
    } else {
      return this.httpClient.get<Process[]>(this.apiEndpoint + '/processes/my');
    }
  }

  observeProcess(id: string): Observable<Process> {
    if (id !== this.cachedProcessId) {
      this.cachedProcessId = id;
      this.cachedProcess = interval(environment.updateInterval).pipe(
        startWith(void 0),
        switchMap(() => this.httpClient.get<Process>(this.apiEndpoint + '/process/' + id)),
        shareReplay({ bufferSize: 1, refCount: true }),
      );
    }
    return this.cachedProcess!;
  }

  getProcess(id: string): Observable<Process> {
    return this.observeProcess(id).pipe(first());
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
