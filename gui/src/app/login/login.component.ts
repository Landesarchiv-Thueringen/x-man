import { animate, style, transition, trigger } from '@angular/animations';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { Router } from '@angular/router';
import { AuthService } from '../utility/authorization/auth.service';
import { LoginService } from './login.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [MatFormFieldModule, MatInputModule, MatCardModule, MatButtonModule, ReactiveFormsModule],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss',
  animations: [
    trigger('slideIn', [transition(':enter', [style({ height: 0 }), animate('200ms', style({ height: '*' }))])]),
  ],
})
export class LoginComponent implements OnInit {
  invalidCredentials = false;

  readonly loginForm = new FormGroup({
    username: new FormControl('', Validators.required),
    password: new FormControl('', Validators.required),
  });

  constructor(
    private auth: AuthService,
    private loginService: LoginService,
    private router: Router,
  ) {}

  ngOnInit(): void {
    if (this.auth.isLoggedIn()) {
      this.router.navigate(['/']);
    }
  }

  async login() {
    if (this.loginForm.valid) {
      this.invalidCredentials = false;
      try {
        await this.auth.login(this.loginForm.value.username!, this.loginForm.value.password!);
        if (this.loginService.afterLoginUrl) {
          this.router.navigateByUrl(this.loginService.afterLoginUrl);
          this.loginService.afterLoginUrl = null;
        } else {
          this.router.navigate(['/']);
        }
      } catch (e) {
        if ((e as HttpErrorResponse).status === 401) {
          this.invalidCredentials = true;
        }
      }
    }
  }
}
