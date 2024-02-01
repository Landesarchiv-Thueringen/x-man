import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { Agency } from '../agencies/agencies.service';
import { User } from './users.service';

interface UserDetailsData {
  user: User;
  agencies: Agency[];
}

/**
 * User metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-user-details',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, MatChipsModule, MatListModule, MatIconModule],
  templateUrl: './user-details.component.html',
  styleUrl: './user-details.component.scss',
})
export class UserDetailsComponent {
  hasPermissions = Object.values(this.data.user.permissions).some(isTrue);

  constructor(
    private dialogRef: MatDialogRef<UserDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public data: UserDetailsData,
  ) {}
}

function isTrue(value: boolean) {
  return value;
}
