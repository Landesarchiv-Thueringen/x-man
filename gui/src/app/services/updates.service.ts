import { Injectable, NgZone } from '@angular/core';
import { Observable, Subject, distinctUntilChanged, filter, map } from 'rxjs';
import { environment } from '../../environments/environment';
import { notNull } from '../utils/predicates';
import { AuthService } from './auth.service';

export interface Update {
  collection: 'submission_processes' | 'processing_errors' | 'tasks';
  processId: string;
  operation: 'insert' | 'update' | 'delete';
}

@Injectable({
  providedIn: 'root',
})
export class UpdatesService {
  private readonly updatesSubject = new Subject<Update>();
  private eventSource?: EventSource;

  constructor(
    private auth: AuthService,
    private ngZone: NgZone,
  ) {
    auth
      .observeLoginInformation()
      .pipe(map(notNull), distinctUntilChanged())
      .subscribe((isLoggedIn) => {
        if (isLoggedIn) {
          this.subscribe();
        } else {
          this.unsubscribe();
        }
      });
  }

  observe(collection: Update['collection']): Observable<Update> {
    return this.updatesSubject.pipe(filter((update) => update.collection === collection));
  }

  private subscribe() {
    if (this.eventSource) {
      return;
    }
    const token = this.auth.getToken();
    // EventSource doesn't support the authorization header, so we append the
    // token as query parameter.
    this.eventSource = new EventSource(environment.endpoint + '/updates?token=' + token);
    this.eventSource.onmessage = (event) => {
      const messageData: Update = JSON.parse(event.data);
      this.ngZone.run(() => this.updatesSubject.next(messageData));
    };
    // Try to reconnect after a connection loss with a small delay. The delay
    // has two reason:
    // 1. In case of a persisting connection problem, we limit the reconnection
    //    attempts.
    // 2. This handler also triggers on a page reload when the client itself
    //    canceled the connection. If we would resubscribe right away, we would
    //    carry two subscriptions into the reloaded page. However, the timeout
    //    will be canceled after the reload, so with this, we keep only the new
    //    subscription.
    this.eventSource.onerror = () => {
      setTimeout(() => {
        this.unsubscribe();
        this.subscribe();
      }, 1000);
    };
  }

  private unsubscribe() {
    this.eventSource?.close();
    this.eventSource = undefined;
  }
}
