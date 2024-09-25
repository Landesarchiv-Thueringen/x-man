import { HttpClient } from '@angular/common/http';
import { Injectable, Signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { distinctUntilChanged, map, of, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { notNull } from '../utils/predicates';
import { AuthService } from './auth.service';

export interface Config {
  deleteArchivedProcessesAfterDays: number;
  appraisalLevel: 'root' | 'all';
  supportsEmailNotifications: boolean;
  archiveTarget: 'dimag' | 'filesystem';
  borgSupport: boolean;
}

/**
 * Provides information about the server configuration.
 */
@Injectable({
  providedIn: 'root',
})
export class ConfigService {
  readonly config: Signal<Config | undefined>;

  constructor(
    private auth: AuthService,
    private httpClient: HttpClient,
  ) {
    const config = this.auth.observeLoginInformation().pipe(
      map(notNull),
      distinctUntilChanged(),
      switchMap((isLoggedIn) => {
        if (isLoggedIn) {
          return this.httpClient.get<Config>(environment.endpoint + '/config');
        } else {
          return of(undefined);
        }
      }),
    );
    this.config = toSignal(config);
  }
}
