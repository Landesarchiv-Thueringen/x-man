import { HttpClient } from '@angular/common/http';
import { Injectable, inject, isDevMode } from '@angular/core';
import { Observable, map, shareReplay } from 'rxjs';

interface AboutInformation {
  version: string;
}

@Injectable({
  providedIn: 'root',
})
export class AboutService {
  readonly aboutInformation: Observable<AboutInformation>;
  readonly versionSuffix = isDevMode() ? '-dev' : '';

  constructor() {
    const httpClient = inject(HttpClient);

    this.aboutInformation = httpClient.get<AboutInformation>('/api/about').pipe(
      map(({ version, ...aboutInfo }) => ({ ...aboutInfo, version: version + this.versionSuffix })),
      shareReplay(1),
    );
  }
}
