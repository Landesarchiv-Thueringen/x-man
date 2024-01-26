import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, firstValueFrom } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../../environments/environment';
import { User } from '../../admin/users/users.service';

export interface Permissions {
  admin: boolean;
}

interface LoginInformation {
  token: string;
  user: User;
}

/**
 * Provides methods to handle user authentication against the server.
 *
 * The user logs in with their LDAP username / password combination. They then
 * obtain a JWT form the server, which we then send with every request to the
 * server.
 */
@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private loginInformation = new BehaviorSubject<LoginInformation | null>(null);

  constructor(
    private httpClient: HttpClient,
    private router: Router,
  ) {
    const json = localStorage.getItem('loginInformation');
    if (json) {
      this.loginInformation.next(JSON.parse(json));
    }
  }

  getToken(): string | null {
    return this.loginInformation.value?.token ?? null;
  }

  getCurrentLoginInformation(): LoginInformation | null {
    return this.loginInformation.value;
  }

  observeLoginInformation(): Observable<LoginInformation | null> {
    return this.loginInformation;
  }

  isLoggedIn(): boolean {
    return this.loginInformation.value != null;
  }

  isAdmin(): boolean {
    return this.loginInformation.value?.user.permissions.admin ?? false;
  }

  async login(username: string, password: string): Promise<void> {
    const headers = new HttpHeaders({
      Authorization: 'Basic ' + btoa(`${username}:${password}`),
    });
    const observable = this.httpClient.get<LoginInformation>(environment.endpoint + '/login', { headers }).pipe(
      tap((loginInformation) => {
        this.loginInformation.next(loginInformation);
        localStorage.setItem('loginInformation', JSON.stringify(loginInformation));
      }),
    );
    await firstValueFrom(observable);
  }

  logout(): void {
    this.loginInformation.next(null);
    localStorage.removeItem('loginInformation');
    this.router.navigate(['login']);
  }
}
