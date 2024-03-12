import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, map, shareReplay } from 'rxjs';
import { environment } from '../../environments/environment';

interface AboutInformation {
  version: string;
}

@Injectable({
  providedIn: 'root',
})
export class AboutService {
  readonly aboutInformation: Observable<AboutInformation>;
  readonly versionSuffix = environment.production ? '' : '-dev';

  constructor(httpClient: HttpClient) {
    this.aboutInformation = httpClient.get<AboutInformation>(environment.endpoint + '/about').pipe(
      map(({ version, ...aboutInfo }) => ({ ...aboutInfo, version: version + this.versionSuffix })),
      shareReplay(1),
    );
  }
}
