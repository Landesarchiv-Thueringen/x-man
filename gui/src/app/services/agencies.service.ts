import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

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
  private httpClient = inject(HttpClient);

  private readonly agencies = new BehaviorSubject<Agency[]>([]);

  constructor() {
    this.fetchAgencies();
  }

  private fetchAgencies() {
    this.httpClient
      .get<Agency[]>('/api/agencies')
      .subscribe((agencies) => this.agencies.next(agencies));
  }

  observeAgencies(): Observable<Agency[]> {
    return this.agencies;
  }

  createAgency(agency: Omit<Agency, 'id'>) {
    this.httpClient.put<{ id: string }>('/api/agency', agency).subscribe(({ id }) => {
      this.agencies.next([...this.agencies.value, { ...agency, id }]);
    });
  }

  updateAgency(id: string, updatedValues: Omit<Agency, 'id'>) {
    const found = this.agencies.value.find((i) => i.id === id);
    if (found) {
      Object.assign(found, updatedValues);
      this.agencies.next(this.agencies.value);
      this.httpClient.post('/api/agency', found).subscribe();
    }
  }

  deleteAgency(agency: Agency) {
    this.agencies.next(this.agencies.value.filter((i) => i !== agency));
    this.httpClient.delete('/api/agency/' + agency.id).subscribe();
  }
}
