import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Agency {
  id: string;
  name: string;
  abbreviation: string;
  prefix: string;
  code: string;
  contactEmail: string;
  transferDirURL: string;
  collectionId?: string;
  users: string[];
}

@Injectable({
  providedIn: 'root',
})
export class AgenciesService {
  private readonly agencies = new BehaviorSubject<Agency[]>([]);

  constructor(private httpClient: HttpClient) {
    this.fetchAgencies();
  }

  private fetchAgencies() {
    this.httpClient
      .get<Agency[]>(environment.endpoint + '/agencies')
      .subscribe((agencies) => this.agencies.next(agencies));
  }

  observeAgencies(): Observable<Agency[]> {
    return this.agencies;
  }

  createAgency(agency: Omit<Agency, 'id'>) {
    this.httpClient.put<{ id: string }>(environment.endpoint + '/agency', agency).subscribe(({ id }) => {
      this.agencies.next([...this.agencies.value, { ...agency, id }]);
    });
  }

  updateAgency(id: string, updatedValues: Omit<Agency, 'id'>) {
    const found = this.agencies.value.find((i) => i.id === id);
    if (found) {
      Object.assign(found, updatedValues);
      this.agencies.next(this.agencies.value);
      this.httpClient.post(environment.endpoint + '/agency', found).subscribe();
    }
  }

  deleteAgency(agency: Agency) {
    this.agencies.next(this.agencies.value.filter((i) => i !== agency));
    this.httpClient.delete(environment.endpoint + '/agency/' + agency.id).subscribe();
  }
}
