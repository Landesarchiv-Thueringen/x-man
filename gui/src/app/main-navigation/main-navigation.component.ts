import { Component } from '@angular/core';
import { AuthService } from '../utility/authorization/auth.service';

@Component({
  selector: 'app-main-navigation',
  templateUrl: './main-navigation.component.html',
  styleUrls: ['./main-navigation.component.scss'],
})
export class MainNavigationComponent {
  loginInformation = this.auth.observeLoginInformation();

  constructor(private auth: AuthService) {}

  logout(): void {
    this.auth.logout();
  }
}
