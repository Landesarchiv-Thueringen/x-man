import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, first, shareReplay, switchMap } from 'rxjs';
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
  readonly config: Observable<Config>;

  constructor(
    private auth: AuthService,
    private httpClient: HttpClient,
  ) {
    this.config = this.auth.observeLoginInformation().pipe(
      first(notNull),
      switchMap(() => this.httpClient.get<Config>(environment.endpoint + '/config')),
      shareReplay(1),
    );
  }
}
