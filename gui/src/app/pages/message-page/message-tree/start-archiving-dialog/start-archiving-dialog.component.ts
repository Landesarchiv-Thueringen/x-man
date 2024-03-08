import { Component } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';

@Component({
  selector: 'app-start-archiving-dialog',
  templateUrl: './start-archiving-dialog.component.html',
  styleUrls: ['./start-archiving-dialog.component.scss'],
  standalone: true,
  imports: [MatDialogModule, MatButtonModule],
})
export class StartArchivingDialogComponent {
  constructor(private dialogRef: MatDialogRef<StartArchivingDialogComponent>) {}

  startArchivingProcess() {
    this.dialogRef.close({
      startedArchivingProcess: true,
    });
  }
}
