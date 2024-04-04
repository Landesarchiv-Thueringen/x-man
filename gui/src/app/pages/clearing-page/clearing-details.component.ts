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
import { ClearingService, ProcessingError } from '../../services/clearing.service';
import { NotificationService } from '../../services/notification.service';
import { BreakOpportunitiesPipe } from '../../shared/break-opportunities.pipe';

@Component({
  selector: 'app-clearing-details',
  standalone: true,
  imports: [
    BreakOpportunitiesPipe,
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

  sendEmail() {
    let subject: string;
    switch (this.data.message?.messageType.code) {
      case '0501':
        subject = 'Fehler bei xdomea-Anbietung';
        break;
      case '0503':
        subject = 'Fehler bei xdomea-Abgabe';
        break;
      case '0505':
        subject = 'Fehler bei xdomea-Bewertungsbestätigung';
        break;
      default:
        subject = 'Fehler bei xdomea-Nachricht';
    }
    let body = 'Beim Einlesen einer xdomea-Nachricht ein ein Fehler aufgetreten: ' + this.data.description;
    if (this.data.additionalInfo) {
      body += '\n\nFehlerausgabe vom System:\n' + this.data.additionalInfo;
    }
    const a = document.createElement('a');
    a.setAttribute(
      'href',
      `mailto:${this.data.agency!.contactEmail}?subject=${subject}&body=${encodeURIComponent(body)}`,
    );
    a.click();
  }

  markSolved() {
    this.clearingService.resolveError(this.data.id, 'mark-solved').subscribe(() => {
      this.dialogRef.close();
    });
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
