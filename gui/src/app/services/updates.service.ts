import { Injectable, signal } from '@angular/core';
import {
  Observable,
  Subject,
  distinctUntilChanged,
  filter,
  map,
  startWith,
  throttleTime,
} from 'rxjs';
import { environment } from '../../environments/environment';
import { NIL_UUID } from '../utils/constants';
import { notNull } from '../utils/predicates';
import { AuthService } from './auth.service';

export interface Update {
  collection: 'submission_processes' | 'processing_errors' | 'warnings' | 'tasks';
  processId: string;
  operation: 'insert' | 'update' | 'delete';
}

/**
 * Provides methods to receive updates from the server via server-sent events.
 */
@Injectable({
  providedIn: 'root',
})
export class UpdatesService {
  private readonly updatesSubject = new Subject<Update | 'reconnect'>();
  private eventSource?: EventSource;
  private keepAliveTimer?: number;
  private reconnectTimer?: number;
  private connectTimeout?: number;
  /**
   * The connection state of the event source.
   *
   * Will be updated to 'failed' or 'connected' when either state is reached.
   * While reconnecting, the previous state is kept until the connection is
   * reestablished or a timeout is reached.
   */
  state = signal<null | 'failed' | 'connected'>(null);

  constructor(private auth: AuthService) {
    auth
      .observeLoginInformation()
      .pipe(map(notNull), distinctUntilChanged())
      .subscribe((isLoggedIn) => {
        if (isLoggedIn) {
          this.subscribe();
        } else {
          this.state.set(null);
          this.unsubscribe();
          if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = undefined;
          }
          if (this.connectTimeout) {
            window.clearTimeout(this.connectTimeout);
            this.connectTimeout = undefined;
          }
        }
      });
    // Try to reconnect on page focus, i.e., when the user switches to the page
    // tab or clicks the browser window. This is mainly for easier development
    // after restarting the server.
    window.addEventListener('focus', () => {
      if (this.state() === 'failed' && auth.isLoggedIn()) {
        this.subscribe();
      }
    });
    // Prevent error when leaving or reloading the page.
    //
    // Note that this inhibits the browser's back/forward cache, but this is not
    // particularly useful for a single-page application anyway.
    addEventListener('beforeunload', () => this.unsubscribe());
  }

  // Emits each time the given database collection could have changed.
  //
  // Also emits once on subscription and throttles to at most one emission per
  // 200ms.
  observeCollection(collection: Update['collection']): Observable<void> {
    return this.updatesSubject.pipe(
      filter((update) => update === 'reconnect' || update.collection === collection),
      map(() => void 0),
      startWith(void 0),
      throttleTime(200, undefined, { leading: true, trailing: true }),
    );
  }

  // Emits each time a database change could have caused the given submission
  // process to have changed.
  //
  // Also emits once on subscription and throttles to at most one emission per
  // 200ms.
  observeSubmissionProcess(processId: string): Observable<void> {
    return this.updatesSubject.pipe(
      filter(
        (update) =>
          update === 'reconnect' ||
          update.processId === processId ||
          (update.collection === 'submission_processes' && update.processId === NIL_UUID),
      ),
      map(() => void 0),
      startWith(void 0),
      throttleTime(200, undefined, { leading: true, trailing: true }),
    );
  }

  private subscribe() {
    if (this.eventSource) {
      return;
    }
    if (this.reconnectTimer) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = undefined;
    }
    if (this.connectTimeout) {
      window.clearTimeout(this.connectTimeout);
      this.connectTimeout = undefined;
    }
    const token = this.auth.getToken();
    // EventSource doesn't support the authorization header, so we append the
    // token as query parameter.
    this.eventSource = new EventSource(environment.endpoint + '/updates?token=' + token, {});
    this.eventSource.addEventListener('message', (event) => {
      const messageData: Update = JSON.parse(event.data);
      this.updatesSubject.next(messageData);
    });
    this.connectTimeout = window.setTimeout(() => {
      this.connectTimeout = undefined;
      this.state.set('failed');
    }, 1000);
    this.eventSource.addEventListener('error', () => {
      // When the connection is closed without an error, the browser tries to
      // reconnect automatically. However, it also invokes this error hook in
      // this case. If the readyState is CLOSED, there was an actual error and
      // the browser didn't try to reconnect.
      if (this.eventSource?.readyState === this.eventSource?.CLOSED) {
        // Try to reconnect after a connection loss with a small delay. The delay
        // has two reason:
        // 1. In case of a persisting connection problem, we limit the reconnection
        //    attempts.
        // 2. This handler also triggers on a page reload when the client itself
        //    canceled the connection. If we would resubscribe right away, we would
        //    carry two subscriptions into the reloaded page. However, the timeout
        //    will be canceled after the reload, so with this, we keep only the new
        //    subscription.
        this.unsubscribe();
        this.reconnectTimer = window.setTimeout(() => {
          this.reconnectTimer = undefined;
          this.subscribe();
        }, 10_000);
      }
    });
    this.eventSource.addEventListener('open', () => {
      clearTimeout(this.connectTimeout);
      this.connectTimeout = undefined;
      this.state.set('connected');
      // Emit a `reconnect` event on `open`, except for the first time.
      if (this.keepAliveTimer) {
        this.updatesSubject.next('reconnect');
      }
    });
    this.eventSource.addEventListener('heartbeat', () => this.renewKeepAliveTimer());
  }

  private unsubscribe() {
    this.eventSource?.close();
    this.eventSource = undefined;
  }

  private renewKeepAliveTimer() {
    if (this.keepAliveTimer) {
      window.clearTimeout(this.keepAliveTimer);
    }
    this.keepAliveTimer = window.setTimeout(() => {
      this.unsubscribe();
      this.subscribe();
    }, 45_000);
  }
}
