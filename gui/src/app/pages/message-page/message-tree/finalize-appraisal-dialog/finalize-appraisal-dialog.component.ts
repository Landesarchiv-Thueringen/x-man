import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MessageService } from '../../../../services/message.service';

export interface DialogData {
  processId: string;
}

@Component({
    selector: 'app-finalize-appraisal-dialog',
    templateUrl: './finalize-appraisal-dialog.component.html',
    styleUrls: ['./finalize-appraisal-dialog.component.scss'],
    imports: [CommonModule, MatDialogModule, MatButtonModule, MatProgressSpinnerModule]
})
export class FinalizeAppraisalDialogComponent {
  data = inject<DialogData>(MAT_DIALOG_DATA);
  private dialogRef = inject<MatDialogRef<FinalizeAppraisalDialogComponent>>(MatDialogRef);
  private messageService = inject(MessageService);

  loading = true;
  appraisalComplete?: boolean;

  constructor() {
    const data = this.data;

    this.messageService
      .areAllRecordObjectsAppraised(data.processId)
      .subscribe((appraisalComplete) => {
        this.loading = false;
        this.appraisalComplete = appraisalComplete;
      });
  }

  sendAppraisalMessage(): void {
    this.dialogRef.close({
      finalizeAppraisal: true,
    });
  }
}
