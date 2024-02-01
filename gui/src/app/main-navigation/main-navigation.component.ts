import { HttpClient } from '@angular/common/http';
import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { environment } from '../../environments/environment';
import { Agency } from '../admin/agencies/agencies.service';
import { UserDetailsComponent } from '../admin/users/user-details.component';
import { AuthService } from '../utility/authorization/auth.service';

@Component({
  selector: 'app-main-navigation',
  templateUrl: './main-navigation.component.html',
  styleUrls: ['./main-navigation.component.scss'],
})
export class MainNavigationComponent {
  loginInformation = this.auth.observeLoginInformation();

  constructor(
    private auth: AuthService,
    private dialog: MatDialog,
    private httpClient: HttpClient,
  ) {}

  openUserDetails() {
    const user = this.auth.getCurrentLoginInformation()!.user;
    this.httpClient.get<Agency[]>(environment.endpoint + '/agencies/my').subscribe((agencies) => {
      this.dialog.open(UserDetailsComponent, { data: { user, agencies } });
    });
  }

  logout(): void {
    this.auth.logout();
  }
}
