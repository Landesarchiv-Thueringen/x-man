import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Collection } from '../collections/collections.service';
import { TransferDirectory } from './transfer-directory.service';

export interface Institution {
  id: number;
  name: string;
  abbreviation: string;
  transferDirectory: TransferDirectory;
  collectionId?: number;
  collection?: Collection;
  userIds: string[];
}

@Injectable({
  providedIn: 'root',
})
export class InstitutionsService {
  private readonly institutions = new BehaviorSubject<Institution[]>([]);

  constructor(private httpClient: HttpClient) {
    this.fetchInstitutions();
  }

  private fetchInstitutions() {
    this.httpClient
      .get<Institution[]>(environment.endpoint + '/institutions')
      .subscribe((institutions) => this.institutions.next(institutions));
  }

  getInstitutions(): Observable<Institution[]> {
    return this.institutions;
  }

  createInstitution(institution: Omit<Institution, 'id'>) {
    this.httpClient.put<string>(environment.endpoint + '/institution', institution).subscribe((response) => {
      const id = parseInt(response);
      this.institutions.next([...this.institutions.value, { ...institution, id }]);
    });
  }

  updateInstitution(id: number, updatedValues: Institution) {
    const found = this.institutions.value.find((i) => i.id === id);
    if (found) {
      Object.assign(found, updatedValues);
      this.institutions.next(this.institutions.value);
      this.httpClient.post(environment.endpoint + '/institution/' + id, updatedValues).subscribe();
    }
  }

  deleteInstitution(institution: Institution) {
    this.institutions.next(this.institutions.value.filter((i) => i !== institution));
    this.httpClient.delete(environment.endpoint + '/institution/' + institution.id).subscribe();
  }
}
