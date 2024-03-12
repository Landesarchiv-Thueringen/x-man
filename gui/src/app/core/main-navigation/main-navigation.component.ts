import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatDividerModule } from '@angular/material/divider';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { RouterModule } from '@angular/router';
import { UserDetailsComponent } from '../../pages/admin-page/users/user-details.component';
import { AboutService } from '../../services/about.service';
import { AuthService } from '../../services/auth.service';
import { UsersService } from '../../services/users.service';
import { AboutDialogComponent } from '../about-dialog/about-dialog.component';

@Component({
  selector: 'app-main-navigation',
  templateUrl: './main-navigation.component.html',
  styleUrls: ['./main-navigation.component.scss'],
  standalone: true,
  imports: [
    MatDividerModule,
    MatToolbarModule,
    CommonModule,
    MatIconModule,
    MatButtonModule,
    RouterModule,
    MatDialogModule,
  ],
})
export class MainNavigationComponent {
  loginInformation = this.auth.observeLoginInformation();
  aboutInformation = this.about.aboutInformation;

  constructor(
    private auth: AuthService,
    private dialog: MatDialog,
    private users: UsersService,
    private about: AboutService,
  ) {}

  openUserDetails() {
    const user = this.auth.getCurrentLoginInformation()!.user;
    this.users.getUserInformation().subscribe(({ agencies, preferences }) => {
      this.dialog.open(UserDetailsComponent, { data: { user, agencies, preferences } });
    });
  }

  openAboutDialog() {
    this.dialog.open(AboutDialogComponent);
  }

  logout(): void {
    this.auth.logout();
  }
}
