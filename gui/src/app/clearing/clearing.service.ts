// angular
import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment';

// project
import { Agency } from '../process/process.service';

export interface ProcessingError {
  id: number;
  detectedAt: string;
  agency: Agency;
  resolved: boolean;
  description: string;
  messageStorePath?: string;
  transferDirPath?: string;
}

@Injectable({
  providedIn: 'root',
})
export class ClearingService {
  apiEndpoint: string;

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  getProcessingErrors() {
    return this.httpClient.get<ProcessingError[]>(this.apiEndpoint + '/processing-errors');
  }
}
