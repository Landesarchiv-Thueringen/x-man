import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

import { Message } from '../message/message.service';
import { ProcessingError } from '../clearing/clearing.service';

export interface Process {
  id: string;
  xdomeaID: string;
  receivedAt: string;
  institution: string;
  archivingComplete: boolean;
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
  complete: boolean;
  completionTime: string;
  itemCount: number;
  itemCompletetCount: number;
}

@Injectable({
  providedIn: 'root'
})
export class ProcessService {
  apiEndpoint: string;

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  getProcesses() {
    return this.httpClient.get<Process[]>(this.apiEndpoint + '/processes');
  }
}
