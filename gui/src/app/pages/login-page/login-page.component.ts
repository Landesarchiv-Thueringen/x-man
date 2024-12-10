import { animate, style, transition, trigger } from '@angular/animations';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, inject } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { LoginService } from '../../services/login.service';

@Component({
    selector: 'app-login-page',
    imports: [
        MatFormFieldModule,
        MatInputModule,
        MatCardModule,
        MatButtonModule,
        ReactiveFormsModule,
    ],
    templateUrl: './login-page.component.html',
    styleUrl: './login-page.component.scss',
    animations: [
        trigger('slideIn', [
            transition(':enter', [style({ height: 0 }), animate('200ms', style({ height: '*' }))]),
        ]),
    ]
})
export class LoginPageComponent implements OnInit {
  private auth = inject(AuthService);
  private loginService = inject(LoginService);
  private router = inject(Router);

  invalidCredentials = false;

  readonly loginForm = new FormGroup({
    username: new FormControl('', Validators.required),
    password: new FormControl('', Validators.required),
  });

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
