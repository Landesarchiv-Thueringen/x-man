import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Router } from '@angular/router';
import { BehaviorSubject, Observable, firstValueFrom } from 'rxjs';
import { tap } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { User } from './users.service';

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
  private httpClient = inject(HttpClient);
  private router = inject(Router);

  private loginInformation = new BehaviorSubject<LoginInformation | null>(null);

  constructor() {
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
      Authorization: 'Basic ' + toBase64(`${username}:${password}`),
    });
    const observable = this.httpClient
      .get<LoginInformation>(environment.endpoint + '/login', { headers })
      .pipe(
        tap((loginInformation) => {
          this.loginInformation.next(loginInformation);
          localStorage.setItem('loginInformation', JSON.stringify(loginInformation));
        }),
      );
    await firstValueFrom(observable);
  }

  /**
   * Removes saved login information and navigates to the login page.
   */
  logout(): void {
    this.loginInformation.next(null);
    localStorage.removeItem('loginInformation');
    this.router.navigate(['login']);
  }
}

/**
 * Converts the a string to base64, correctly handling unicode characters.
 */
function toBase64(s: string): string {
  // Adapted from
  // https://developer.mozilla.org/en-US/docs/Web/API/Window/btoa#unicode_strings.
  const bytes = new TextEncoder().encode(s);
  const binString = Array.from(bytes, (byte) => String.fromCodePoint(byte)).join('');
  return btoa(binString);
}
