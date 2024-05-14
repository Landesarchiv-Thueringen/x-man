import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, filter, map, shareReplay } from 'rxjs';
import { environment } from '../../environments/environment';
import { Agency } from './agencies.service';
import { Permissions } from './auth.service';

export interface User {
  id: string;
  displayName: string;
  permissions: Permissions;
}

export interface UserInformation {
  agencies: Agency[];
  preferences: UserPreferences;
}

export interface UserPreferences {
  messageEmailNotifications: boolean;
  reportByEmail: boolean;
  errorEmailNotifications: boolean;
}

@Injectable({
  providedIn: 'root',
})
export class UsersService {
  private readonly users = this.httpClient.get<User[]>(environment.endpoint + '/users').pipe(shareReplay(1));

  constructor(private httpClient: HttpClient) {}

  getUsers(): Observable<User[]> {
    return this.users.pipe(filter((a) => a.length > 0));
  }

  getUserById(id: string): Observable<User> {
    return this.users.pipe(map((users) => this.findById(users, id)));
  }

  getUsersByIds(ids: string[]): Observable<User[]> {
    return this.users.pipe(map((users) => ids.map((id) => this.findById(users, id))));
  }

  getUserInformation(): Observable<UserInformation> {
    return this.httpClient.get<UserInformation>(environment.endpoint + '/user-info');
  }

  updateUserPreferences(preferences: UserPreferences): Observable<void> {
    return this.httpClient.post<void>(environment.endpoint + '/user-preferences', preferences);
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
