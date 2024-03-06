import { A11yModule } from '@angular/cdk/a11y';
import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { AuthService } from '../../utility/authorization/auth.service';
import { ConfigService } from '../../utility/config.service';
import { Agency } from '../agencies/agencies.service';
import { User, UserPreferences, UsersService } from './users.service';

interface UserDetailsData {
  user: User;
  agencies?: Agency[];
  preferences?: UserPreferences;
}

/**
 * User metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-user-details',
  standalone: true,
  imports: [
    A11yModule,
    CommonModule,
    MatButtonModule,
    MatChipsModule,
    MatDialogModule,
    MatExpansionModule,
    MatIconModule,
    MatListModule,
    MatSlideToggleModule,
    ReactiveFormsModule,
  ],
  templateUrl: './user-details.component.html',
  styleUrl: './user-details.component.scss',
})
export class UserDetailsComponent {
  hasPermissions = Object.values(this.data.user.permissions).some(isTrue);
  isAdmin = this.auth.isAdmin();
  supportsEmailNotifications?: boolean;
  preferences?: FormGroup;

  constructor(
    private dialogRef: MatDialogRef<UserDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public data: UserDetailsData,
    private formBuilder: FormBuilder,
    private auth: AuthService,
    private config: ConfigService,
    private users: UsersService,
  ) {
    this.config.config
      .pipe(takeUntilDestroyed())
      .subscribe((config) => (this.supportsEmailNotifications = config.supportsEmailNotifications));
    if (this.data.preferences) {
      this.preferences = this.formBuilder.group(this.data.preferences);
      this.preferences.valueChanges.subscribe((preferences) => {
        this.users.updateUserPreferences(preferences).subscribe();
      });
    }
  }
}

function isTrue(value: boolean) {
  return value;
}
