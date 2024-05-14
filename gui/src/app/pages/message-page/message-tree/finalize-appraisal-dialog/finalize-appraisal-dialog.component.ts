import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MessageService } from '../../../../services/message.service';

export interface DialogData {
  processId: string;
}

@Component({
  selector: 'app-finalize-appraisal-dialog',
  templateUrl: './finalize-appraisal-dialog.component.html',
  styleUrls: ['./finalize-appraisal-dialog.component.scss'],
  standalone: true,
  imports: [CommonModule, MatDialogModule, MatButtonModule],
})
export class FinalizeAppraisalDialogComponent {
  appraisalComplete: boolean;

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: DialogData,
    private dialogRef: MatDialogRef<FinalizeAppraisalDialogComponent>,
    private messageService: MessageService,
  ) {
    this.appraisalComplete = false;
    this.messageService.areAllRecordObjectsAppraised(data.processId).subscribe({
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
