import { A11yModule } from '@angular/cdk/a11y';
import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { Agency } from '../../../services/agencies.service';
import { AuthService } from '../../../services/auth.service';
import { ConfigService } from '../../../services/config.service';
import { User, UserPreferences, UsersService } from '../../../services/users.service';

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
    styleUrl: './user-details.component.scss'
})
export class UserDetailsComponent {
  private dialogRef = inject<MatDialogRef<UserDetailsComponent>>(MatDialogRef);
  data = inject<UserDetailsData>(MAT_DIALOG_DATA);
  private formBuilder = inject(FormBuilder);
  private auth = inject(AuthService);
  private configService = inject(ConfigService);
  private users = inject(UsersService);

  readonly hasPermissions = Object.values(this.data.user.permissions).some(isTrue);
  readonly isAdmin = this.auth.isAdmin();
  readonly config = this.configService.config;
  readonly preferences?: FormGroup;

  constructor() {
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
