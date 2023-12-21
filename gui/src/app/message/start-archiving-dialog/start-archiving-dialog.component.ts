// angular
import { Component, Inject } from '@angular/core';

// material
import { MatDialogRef } from '@angular/material/dialog';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';

// project
import { MessageService } from '../message.service';

export interface DialogData {
  messageID: string;
}

@Component({
  selector: 'app-start-archiving-dialog',
  templateUrl: './start-archiving-dialog.component.html',
  styleUrls: ['./start-archiving-dialog.component.scss']
})
export class StartArchivingDialogComponent {
  archivingProcessStarted: boolean;

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: DialogData,
    private dialogRef: MatDialogRef<StartArchivingDialogComponent>,
    private messageService: MessageService,
  ) {
    this.archivingProcessStarted = true;
  }
}
