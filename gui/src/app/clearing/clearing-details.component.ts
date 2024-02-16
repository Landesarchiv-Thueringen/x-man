import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { RouterModule } from '@angular/router';
import { MessageService } from '../message/message.service';
import { NotificationService } from '../utility/notification/notification.service';
import { ProcessingError } from './clearing.service';

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
  json: string;

  constructor(
    private dialogRef: MatDialogRef<ClearingDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public data: ProcessingError,
    private messageService: MessageService,
    private notificationService: NotificationService,
  ) {
    this.json = JSON.stringify(data, null, 2);
  }

  reimportMessage() {
    this.messageService.reimportMessage(this.data.message.id).subscribe(() => {
      this.notificationService.show('Nachricht wird neu eingelesen...');
      this.dialogRef.close();
    });
  }

  deleteMessage() {
    this.messageService.deleteMessage(this.data.message.id).subscribe(() => {
      this.notificationService.show('Nachricht gelöscht');
      this.dialogRef.close();
    });
  }
}
