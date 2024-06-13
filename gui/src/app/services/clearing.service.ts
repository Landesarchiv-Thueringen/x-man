import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, combineLatest, map, shareReplay, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { Agency } from '../services/agencies.service';
import { MessageType } from '../services/message.service';
import { UpdatesService } from './updates.service';

type ProcessingErrorResolution =
  | 'mark-solved'
  | 'mark-done'
  | 'reimport-message'
  | 'delete-message'
  | 'delete-transfer-file'
  | 'obsolete';

export interface ProcessingError {
  id: string;
  createdAt: string;
  resolved: boolean;
  resolvedAt: string;
  resolution: ProcessingErrorResolution;
  title: string;
  info: string;
  stack: string;
  agency?: Agency;
  processId: string;
  messageType: MessageType;
  processStep: string;
  transferPath: string;
  taskId: string;
}

@Injectable({
  providedIn: 'root',
})
export class ClearingService {
  private seenTime = new BehaviorSubject(0);
  private processingErrors = this.getProcessingErrorsObservable();

  constructor(
    private httpClient: HttpClient,
    private updates: UpdatesService,
  ) {
    const seenTime = window.localStorage.getItem('processing-errors-seen-time');
    if (seenTime) {
      this.seenTime.next(parseInt(seenTime));
    }
  }

  /** Fetches processing errors every `updateInterval` milliseconds. */
  observeProcessingErrors(): Observable<ProcessingError[]> {
    return this.processingErrors;
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
        ([errors, seenTime]) => errors?.filter((e) => !e.resolved && new Date(e.createdAt).valueOf() > seenTime).length,
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
    return this.httpClient.get<ProcessingError[]>(environment.endpoint + '/processing-errors');
  }

  resolveError(errorId: string, resolution: ProcessingErrorResolution): Observable<void> {
    const url = environment.endpoint + '/processing-errors/resolve/' + errorId;
    return this.httpClient.post<void>(url, resolution);
  }
}
