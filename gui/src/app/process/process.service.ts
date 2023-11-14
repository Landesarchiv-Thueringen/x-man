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
