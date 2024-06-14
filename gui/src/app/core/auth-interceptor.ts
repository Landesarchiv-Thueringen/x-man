import { HttpErrorResponse, HttpHandlerFn, HttpRequest } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { tap } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { AuthService } from '../services/auth.service';
import { LoginService } from '../services/login.service';

/**
 * Intercepts API request to handle user authorization against the backend.
 *
 * - Appends the JWT authorization token to every request if available
 * - Redirects to the login page when the server responded with "401
 *   Unauthorized" to any request.
 * - Redirects to the error page when the server responds with any other error
 *   code.
 */
export function authInterceptor(req: HttpRequest<unknown>, next: HttpHandlerFn) {
  if (!isApiRequest(req)) {
    return next(req);
  }
  const auth = inject(AuthService);
  const router = inject(Router);
  const login = inject(LoginService);
  const token = auth.getToken();
  if (token) {
    req = req.clone({
      headers: req.headers.set('Authorization', `Bearer ${token}`),
    });
  }
  return next(req).pipe(
    tap({
      error: (event) => {
        if (event instanceof HttpErrorResponse) {
          if (event.status === 401) {
            // On the login page, 401 means invalid credentials.
            if (router.url !== '/login') {
              // On any other page, it means our token is invalid. We delete
              // it and let the user log back in again.
              login.afterLoginUrl = router.url;
              auth.logout();
            }
          } else {
            router.navigate(['fehler', event.status], { skipLocationChange: true });
          }
        }
      },
    }),
  );
}

function isApiRequest(req: HttpRequest<unknown>): boolean {
  return req.url.startsWith(environment.endpoint + '/');
}
