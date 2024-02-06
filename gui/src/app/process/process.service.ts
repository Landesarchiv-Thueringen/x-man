import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { ProcessingError } from '../clearing/clearing.service';
import { Message } from '../message/message.service';

export interface Agency {
  name: string;
  abbreviation: string;
}

export interface Process {
  id: string;
  agency: Agency;
  xdomeaID: string;
  receivedAt: string;
  institution: string;
  note: string;
  message0501: Message;
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
  started: boolean;
  complete: boolean;
  startTime?: string;
  completionTime?: string;
  itemCount: number;
  itemCompletedCount: number;
}

@Injectable({
  providedIn: 'root',
})
export class ProcessService {
  apiEndpoint: string;

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

  getProcessByXdomeaID(id: string): Observable<Process> {
    return this.httpClient.get<Process>(this.apiEndpoint + '/process-by-xdomea-id/' + id);
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
