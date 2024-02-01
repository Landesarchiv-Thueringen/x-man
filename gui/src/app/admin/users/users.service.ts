import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, filter, map } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Permissions } from '../../utility/authorization/auth.service';
import { Institution } from '../institutions/institutions.service';

export interface User {
  id: string;
  displayName: string;
  permissions: Permissions;
}

@Injectable({
  providedIn: 'root',
})
export class UsersService {
  private readonly users = new BehaviorSubject<User[]>([]);

  constructor(private httpClient: HttpClient) {
    httpClient.get<User[]>(environment.endpoint + '/users').subscribe((users) => this.users.next(users));
  }

  getUsers(): Observable<User[]> {
    return this.users.pipe(filter((a) => a.length > 0));
  }

  getUserById(id: string): Observable<User> {
    return this.users.pipe(map((users) => this.findById(users, id)));
  }

  getUsersByIds(ids: string[]): Observable<User[]> {
    return this.users.pipe(map((users) => ids.map((id) => this.findById(users, id))));
  }

  getInstitutionsForUser(userId: string): Observable<Institution[]> {
    return this.httpClient.get<Institution[]>(environment.endpoint + '/institutions', { params: { userId } });
  }

  private findById(user: User[], id: string): User {
    return (
      user.find((a) => a.id === id) ?? {
        displayName: '<Unbekannter Mitarbeiter>',
        id: id,
        permissions: {} as Permissions,
      }
    );
  }
}
