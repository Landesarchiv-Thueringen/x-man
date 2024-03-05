import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Collection } from '../collections/collections.service';
import { User } from '../users/users.service';

export interface Agency {
  id: number;
  name: string;
  abbreviation: string;
  prefix: string;
  code: string;
  transferDir: string;
  collectionId?: number;
  collection?: Collection;
  users: User[];
}

export interface CompleteUser extends User {
  agencies: Agency[];
  emailNotifications: boolean;
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
