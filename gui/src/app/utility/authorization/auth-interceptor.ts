import { HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { LoginService } from 'src/app/login/login.service';
import { environment } from 'src/environments/environment';
import { AuthService } from './auth.service';

/**
 * Intercepts API request to handle user authorization against the backend.
 *
 * - Appends the JWT authorization token to every request if available
 * - Redirects to the login page when the server responded with "401
 *   Unauthorized" to any request.
 * - Redirects to the error page when the server responds with any other error
 *   code.
 */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(
    private router: Router,
    private auth: AuthService,
    private login: LoginService,
  ) {}

  intercept(req: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    if (!this.isApiRequest(req)) {
      return next.handle(req);
    }
    const token = this.auth.getToken();
    if (token) {
      req = req.clone({
        headers: req.headers.set('Authorization', `Bearer ${token}`),
      });
    }
    return next.handle(req).pipe(
      tap({
        error: (event) => {
          if (event instanceof HttpErrorResponse) {
            if (event.status === 401) {
              if (this.router.url !== '/login') {
                this.login.afterLoginUrl = this.router.url;
                this.router.navigate(['login']);
              }
            } else {
              this.router.navigate(['error', event.status], { skipLocationChange: true });
            }
          }
        },
      }),
    );
  }

  private isApiRequest(req: HttpRequest<unknown>): boolean {
    return req.url.startsWith(environment.endpoint + '/');
  }
}
