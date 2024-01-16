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
  selector: 'app-finalize-appraisal-dialog',
  templateUrl: './finalize-appraisal-dialog.component.html',
  styleUrls: ['./finalize-appraisal-dialog.component.scss'],
})
export class FinalizeAppraisalDialogComponent {
  appraisalComplete: boolean;

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: DialogData,
    private dialogRef: MatDialogRef<FinalizeAppraisalDialogComponent>,
    private messageService: MessageService,
  ) {
    this.appraisalComplete = false;
    this.messageService.areAllRecordObjectsAppraised(data.messageID).subscribe({
      error: (error: any) => {
        console.error(error);
      },
      next: (appraisalComplete: boolean) => {
        this.appraisalComplete = appraisalComplete;
      },
    });
  }

  sendAppraisalMessage(): void {
    this.dialogRef.close({
      finalizeAppraisal: true,
    });
  }
}
