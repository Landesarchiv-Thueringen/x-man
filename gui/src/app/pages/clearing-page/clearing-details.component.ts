import { CommonModule } from '@angular/common';
import { Component, HostBinding, Inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { ClearingService, ProcessingError } from '../../services/clearing.service';
import { NotificationService } from '../../services/notification.service';
import { TasksService } from '../../services/tasks.service';
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
  @HostBinding('class.resolved') resolved = this.data.resolved;
  processingError: ProcessingError = this.data;

  constructor(
    private dialogRef: MatDialogRef<ClearingDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) private data: ProcessingError,
    private clearingService: ClearingService,
    private notificationService: NotificationService,
    private tasksService: TasksService,
    private authService: AuthService,
  ) {
    if (this.authService.isAdmin()) {
      this.clearingService
        .observeProcessingError(data.id)
        .pipe(takeUntilDestroyed())
        .subscribe((e) => {
          if (e) {
            this.processingError = e;
            this.resolved = e.resolved;
          } else {
            this.dialogRef.close();
          }
        });
    }
  }

  retryTask() {
    this.tasksService.retryTask(this.processingError.taskId).subscribe(() => {
      this.dialogRef.close();
      this.notificationService.show('Prozess wird wiederholt...');
    });
  }

  sendEmail() {
    let subject: string;
    switch (this.processingError.messageType) {
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
    let body = 'Beim Einlesen einer xdomea-Nachricht ist ein Fehler aufgetreten: ' + this.processingError.title;
    if (this.processingError.info || this.processingError.data) {
      body += '\n\nFehlerausgabe vom System:';
    if (this.processingError.info) {
        body += '\n' + this.processingError.info;
      }
      if (this.processingError.data) {
        body += '\n' + JSON.stringify(this.processingError.data, null, 2);
      }
    }
    const a = document.createElement('a');
    a.setAttribute(
      'href',
      `mailto:${this.processingError.agency!.contactEmail}?subject=${subject}&body=${encodeURIComponent(body)}`,
    );
    a.click();
  }

  markSolved() {
    this.clearingService.resolveError(this.processingError.id, 'mark-solved').subscribe(() => {
      this.dialogRef.close();
    });
  }

  markDone() {
    this.clearingService.resolveError(this.processingError.id, 'mark-done').subscribe(() => {
      this.dialogRef.close();
    });
  }

  reimportMessage() {
    this.clearingService.resolveError(this.processingError.id, 'reimport-message').subscribe(() => {
      this.notificationService.show('Nachricht wird neu eingelesen...');
      this.dialogRef.close();
    });
  }

  deleteMessage() {
    this.clearingService.resolveError(this.processingError.id, 'delete-message').subscribe(() => {
      this.notificationService.show('Nachricht gelöscht');
      this.dialogRef.close();
    });
  }

  deleteTransferFile() {
    this.clearingService.resolveError(this.processingError.id, 'delete-transfer-file').subscribe(() => {
      this.notificationService.show('Transferdatei gelöscht');
      this.dialogRef.close();
    });
  }

  ignoreTransferFiles() {
    this.clearingService.resolveError(this.processingError.id, 'ignore-transfer-files').subscribe(() => {
      this.notificationService.show('Dateien ignoriert');
      this.dialogRef.close();
    });
  }

  deleteTransferFiles() {
    this.clearingService.resolveError(this.processingError.id, 'delete-transfer-files').subscribe(() => {
      this.notificationService.show('Dateien gelöscht');
      this.dialogRef.close();
    });
  }
}
