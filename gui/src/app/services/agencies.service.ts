import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { Collection } from '../pages/admin-page/collections/collections.service';
import { User } from './users.service';

export interface Agency {
  id: number;
  name: string;
  abbreviation: string;
  prefix: string;
  code: string;
  contactEmail: string;
  transferDirURL: string;
  collectionId?: number;
  collection?: Collection;
  users: User[];
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

  getAgencies(): Observable<Agency[]> {
    return this.agencies;
  }

  createAgency(agency: Omit<Agency, 'id'>) {
    this.httpClient.put<string>(environment.endpoint + '/agency', agency).subscribe((response) => {
      const id = parseInt(response);
      this.agencies.next([...this.agencies.value, { ...agency, id }]);
    });
  }

  updateAgency(id: number, updatedValues: Omit<Agency, 'id'>) {
    const found = this.agencies.value.find((i) => i.id === id);
    if (found) {
      Object.assign(found, updatedValues);
      this.agencies.next(this.agencies.value);
      this.httpClient.post(environment.endpoint + '/agency/' + id, updatedValues).subscribe();
    }
  }

  deleteAgency(agency: Agency) {
    this.agencies.next(this.agencies.value.filter((i) => i !== agency));
    this.httpClient.delete(environment.endpoint + '/agency/' + agency.id).subscribe();
  }
}
