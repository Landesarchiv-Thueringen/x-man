import { Injectable, inject } from '@angular/core';
import {
  ActivatedRouteSnapshot,
  CanActivateFn,
  Router,
  RouterStateSnapshot,
} from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { LoginService } from '../services/login.service';

export const isLoggedIn: CanActivateFn = (route, state) =>
  inject(AuthGuards).isLoggedIn(route, state);
export const isAdmin: CanActivateFn = (route, state) => inject(AuthGuards).isAdmin(route, state);

@Injectable({ providedIn: 'root' })
export class AuthGuards {
  constructor(
    private auth: AuthService,
    private login: LoginService,
    private router: Router,
  ) {}

  /** Redirects to the login page if the user is not logged in. */
  isLoggedIn(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot,
  ): Observable<boolean> | Promise<boolean> | boolean {
    if (!this.auth.isLoggedIn()) {
      this.login.afterLoginUrl = state.url;
      this.router.navigate(['login']);
      return false;
    }
    return true;
  }

  isAdmin(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot,
  ): Observable<boolean> | Promise<boolean> | boolean {
    if (!this.isLoggedIn(route, state)) {
      return false;
    }
    if (!this.auth.isAdmin()) {
      this.router.navigate(['fehler/403']);
      return false;
    }
    return true;
  }
}
