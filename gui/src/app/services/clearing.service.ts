import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { BehaviorSubject, Observable, combineLatest, map, shareReplay, switchMap } from 'rxjs';
import { Agency } from '../services/agencies.service';
import { MessageType } from '../services/message.service';
import { UpdatesService } from './updates.service';

type ProcessingErrorResolution =
  | 'ignore-problem'
  | 'skip-task'
  | 'retry-task'
  | 'reimport-message'
  | 'delete-message'
  | 'delete-transfer-file'
  | 'ignore-transfer-files'
  | 'delete-transfer-files'
  | 'obsolete';

export interface ProcessingError {
  id: string;
  createdAt: string;
  resolved: boolean;
  resolvedAt: string;
  resolution: ProcessingErrorResolution;
  title: string;
  info: string;
  data: any;
  errorType: string;
  stack: string;
  agency?: Agency;
  processId: string | null;
  messageType: MessageType;
  processStep: string;
  transferPath: string;
  taskId: string;
}

@Injectable({
  providedIn: 'root',
})
export class ClearingService {
  private httpClient = inject(HttpClient);
  private updates = inject(UpdatesService);

  private seenTime = new BehaviorSubject(0);
  private processingErrors = this.getProcessingErrorsObservable();

  constructor() {
    const seenTime = window.localStorage.getItem('processing-errors-seen-time');
    if (seenTime) {
      this.seenTime.next(parseInt(seenTime));
    }
  }

  observeProcessingErrors(): Observable<ProcessingError[]> {
    return this.processingErrors;
  }

  observeProcessingError(id: string): Observable<ProcessingError | undefined> {
    return this.processingErrors.pipe(
      map((processingErrors) => processingErrors.find((e) => e.id === id)),
    );
  }

  private getProcessingErrorsObservable(): Observable<ProcessingError[]> {
    return this.updates.observeCollection('processing_errors').pipe(
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
    return combineLatest([this.observeProcessingErrors(), this.seenTime]).pipe(
      map(
        ([errors, seenTime]) =>
          errors?.filter((e) => !e.resolved && new Date(e.createdAt).valueOf() > seenTime).length,
      ),
    );
  }

  getLastSeenTime(): number {
    return this.seenTime.value;
  }

  /**
   * Resets the number returned by `observeNumberUnseen` to 0.
   */
  markAllSeen(time = Date.now()): void {
    window.localStorage.setItem('processing-errors-seen-time', time.toString());
    this.seenTime.next(time);
  }

  private getProcessingErrors() {
    return this.httpClient.get<ProcessingError[]>('/api/processing-errors');
  }

  resolveError(errorId: string, resolution: ProcessingErrorResolution): Observable<void> {
    const url = '/api/processing-errors/resolve/' + errorId;
    return this.httpClient.post<void>(url, resolution);
  }
}
