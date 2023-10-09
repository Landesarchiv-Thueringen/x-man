import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

import { Message } from '../message/message.service';

export interface Process {
  id: string;
  xdomeaID: string;
  createdAt: string;
  messages: Message[];
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
