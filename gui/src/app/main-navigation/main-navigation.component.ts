import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { UserDetailsComponent } from '../admin/users/user-details.component';
import { UsersService } from '../admin/users/users.service';
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
    private users: UsersService,
  ) {}

  openUserDetails() {
    const user = this.auth.getCurrentLoginInformation()!.user;
    this.users.getUserInformation().subscribe(({ agencies, preferences }) => {
      this.dialog.open(UserDetailsComponent, { data: { user, agencies, preferences } });
    });
  }

  logout(): void {
    this.auth.logout();
  }
}
