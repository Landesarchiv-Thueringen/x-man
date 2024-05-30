import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, interval, map, shareReplay, startWith, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { Agency } from '../services/agencies.service';
import { MessageType } from '../services/message.service';

type ProcessingErrorResolution = 'mark-solved' | 'reimport-message' | 'delete-message' | 'delete-transfer-file';

export interface ProcessingError {
  id: string;
  createdAt: string;
  resolved: boolean;
  resolution: ProcessingErrorResolution;
  title: string;
  info: string;
  stack: string;
  agency?: Agency;
  processId?: string;
  messageType?: MessageType;
  transferPath?: string;
}

@Injectable({
  providedIn: 'root',
})
export class ClearingService {
  apiEndpoint: string;
  seenTime = parseInt(window.localStorage.getItem('processing-errors-seen-time') ?? '0');

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  /** Fetches processing errors every `updateInterval` milliseconds. */
  observeProcessingErrors(): Observable<ProcessingError[]> {
    return interval(environment.updateInterval).pipe(
      startWith(void 0), // initial fetch
      switchMap(() => this.getProcessingErrors()),
      map((errors) => errors ?? []),
      shareReplay({ bufferSize: 1, refCount: true }),
    );
  }

  /**
   * Returns the number of new unresolved processing errors since `markAllSeen`
   * was called.
   */
  observeNumberUnseen(): Observable<number> {
    return this.observeProcessingErrors().pipe(
      map((errors) => errors?.filter((e) => !e.resolved && new Date(e.createdAt).valueOf() > this.seenTime).length),
    );
  }

  /**
   * Resets the number returned by `observeNumberUnseen` to 0.
   */
  markAllSeen(): void {
    const now = Date.now();
    window.localStorage.setItem('processing-errors-seen-time', now.toString());
    this.seenTime = now;
  }

  private getProcessingErrors() {
    return this.httpClient.get<ProcessingError[]>(this.apiEndpoint + '/processing-errors');
  }

  resolveError(errorId: string, resolution: ProcessingErrorResolution): Observable<void> {
    const url = this.apiEndpoint + '/processing-errors/resolve/' + errorId;
    return this.httpClient.post<void>(url, resolution);
  }
}
