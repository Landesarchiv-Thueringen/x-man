import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class LoginService {
  /** The URL to which to redirect after the user logged in successfully. */
  afterLoginUrl: string | null = null;
}
