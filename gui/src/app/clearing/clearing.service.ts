import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { Message } from '../message/message.service';
import { Agency, Process } from '../process/process.service';

type ProcessingErrorType = 'agency-mismatch';
type ProcessingErrorResolution = 'reimport-message' | 'delete-message';

export interface ProcessingError {
  id: number;
  detectedAt: string;
  type: ProcessingErrorType;
  agency: Agency;
  resolved: boolean;
  resolution: ProcessingErrorResolution;
  description: string;
  additionalInfo: string;
  process?: Process;
  message?: Message;
  transferPath?: string;
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

  resolveError(errorId: number, resolution: ProcessingErrorResolution): Observable<void> {
    const url = this.apiEndpoint + '/processing-errors/resolve/' + errorId;
    return this.httpClient.post<void>(url, resolution);
  }
}
