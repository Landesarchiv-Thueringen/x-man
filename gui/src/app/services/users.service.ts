import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, shareReplay } from 'rxjs';
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
  private httpClient = inject(HttpClient);

  private readonly users = this.httpClient.get<User[]>('/api/users').pipe(shareReplay(1));

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
    return this.httpClient.get<UserInformation>('/api/user-info');
  }

  updateUserPreferences(preferences: UserPreferences): Observable<void> {
    return this.httpClient.post<void>('/api/user-preferences', preferences);
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
