import { CommonModule } from '@angular/common';
import { Component, HostBinding, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { RouterModule } from '@angular/router';
import { NotificationService } from '../utility/notification/notification.service';
import { ClearingService, ProcessingError } from './clearing.service';

@Component({
  selector: 'app-clearing-details',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatMenuModule,
    RouterModule,
  ],
  templateUrl: './clearing-details.component.html',
  styleUrl: './clearing-details.component.scss',
})
export class ClearingDetailsComponent {
  @HostBinding('class.resolved') readonly resolved = this.data.resolved;

  json: string;

  constructor(
    private dialogRef: MatDialogRef<ClearingDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public data: ProcessingError,
    private clearingService: ClearingService,
    private notificationService: NotificationService,
  ) {
    this.json = JSON.stringify(data, null, 2);
  }

  reimportMessage() {
    this.clearingService.resolveError(this.data.id, 'reimport-message').subscribe(() => {
      this.notificationService.show('Nachricht wird neu eingelesen...');
      this.dialogRef.close();
    });
  }

  deleteMessage() {
    this.clearingService.resolveError(this.data.id, 'delete-message').subscribe(() => {
      this.notificationService.show('Nachricht gelöscht');
      this.dialogRef.close();
    });
  }
}
