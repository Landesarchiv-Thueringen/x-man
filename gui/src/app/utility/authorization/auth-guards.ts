import { Injectable, inject } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivateFn, Router, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { LoginService } from 'src/app/login/login.service';
import { AuthService } from './auth.service';

export const isLoggedIn: CanActivateFn = (route, state) => inject(AuthGuards).isLoggedIn(route, state);

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
    }
    return true;
  }
}