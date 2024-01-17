import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
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
  startTime?: string;
  complete: boolean;
  completionTime: string;
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

  getProcesses() {
    return this.httpClient.get<Process[]>(this.apiEndpoint + '/processes');
  }

  getProcessByXdomeaID(id: string): Observable<Process> {
    return this.httpClient.get<Process>(this.apiEndpoint + '/process-by-xdomea-id/' + id);
  }

  getReport(processId: string): Observable<Blob> {
    return this.httpClient.get(this.apiEndpoint + '/report/' + processId, { responseType: 'blob' });
  }
}
