import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatBadgeModule } from '@angular/material/badge';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatDividerModule } from '@angular/material/divider';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { RouterModule } from '@angular/router';
import { UserDetailsComponent } from '../../pages/admin-page/users/user-details.component';
import { AuthService } from '../../services/auth.service';
import { ClearingService } from '../../services/clearing.service';
import { UpdatesService } from '../../services/updates.service';
import { UsersService } from '../../services/users.service';

@Component({
    selector: 'app-main-navigation',
    templateUrl: './main-navigation.component.html',
    styleUrls: ['./main-navigation.component.scss'],
    imports: [
        CommonModule,
        MatBadgeModule,
        MatButtonModule,
        MatDialogModule,
        MatDividerModule,
        MatIconModule,
        MatToolbarModule,
        RouterModule,
    ]
})
export class MainNavigationComponent {
  private auth = inject(AuthService);
  private dialog = inject(MatDialog);
  private users = inject(UsersService);
  private clearing = inject(ClearingService);
  private updates = inject(UpdatesService);

  loginInformation = this.auth.observeLoginInformation();
  unseenProcessingErrors = this.clearing.observeNumberUnseen();
  connectionState = this.updates.state;

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
