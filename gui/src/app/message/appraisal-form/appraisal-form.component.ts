// angular
import { Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';

// material
import { MatDialogRef } from '@angular/material/dialog';

// project
import {
  MessageService, RecordObjectAppraisal,
} from '../../message/message.service';

@Component({
  selector: 'app-appraisal-form',
  templateUrl: './appraisal-form.component.html',
  styleUrls: ['./appraisal-form.component.scss']
})
export class AppraisalFormComponent {
  form: FormGroup;
  recordObjectAppraisals?: RecordObjectAppraisal[];
  selectedAppraisal?: string;

  constructor(
    private dialogRef: MatDialogRef<AppraisalFormComponent>,
    private formBuilder: FormBuilder,
    private messageService: MessageService,
  ) {
    this.form = this.formBuilder.group({
      appraisal: new FormControl<string | null>(null, Validators.required),
      appraisalNote: new FormControl<string | null>(null),
    });
    this.messageService.getRecordObjectAppraisals().subscribe(
      (appraisals: RecordObjectAppraisal[]) => {
      this.recordObjectAppraisals = appraisals;
    });
  }

  onSubmit(): void {
    if (this.form.valid) {
      this.dialogRef.close({
        appraisalCode: this.selectedAppraisal,
      })
    }
  }
}
