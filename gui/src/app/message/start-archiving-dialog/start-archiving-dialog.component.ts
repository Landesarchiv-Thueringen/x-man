// angular
import { Component, Inject } from '@angular/core';

// material
import { MatDialogRef } from '@angular/material/dialog';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';

@Component({
  selector: 'app-start-archiving-dialog',
  templateUrl: './start-archiving-dialog.component.html',
  styleUrls: ['./start-archiving-dialog.component.scss']
})
export class StartArchivingDialogComponent {

  constructor(
    private dialogRef: MatDialogRef<StartArchivingDialogComponent>
  ) {}

  startArchivingProcess() {
    this.dialogRef.close({
      startedArchivingProcess: true,
    })
  }
}
