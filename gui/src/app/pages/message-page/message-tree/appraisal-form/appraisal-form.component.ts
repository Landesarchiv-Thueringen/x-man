import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import {
  FormBuilder,
  FormControl,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { appraisalDescriptions } from '../../../../services/appraisal.service';

@Component({
    selector: 'app-appraisal-form',
    templateUrl: './appraisal-form.component.html',
    styleUrls: ['./appraisal-form.component.scss'],
    imports: [
        CommonModule,
        MatButtonModule,
        MatDialogModule,
        MatFormFieldModule,
        MatInputModule,
        MatSelectModule,
        ReactiveFormsModule,
    ]
})
export class AppraisalFormComponent {
  form: FormGroup;
  appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({ code, ...d }));
  selectedAppraisal?: string;

  constructor(
    private dialogRef: MatDialogRef<AppraisalFormComponent>,
    private formBuilder: FormBuilder,
  ) {
    this.form = this.formBuilder.group({
      appraisal: new FormControl<string | null>(null, Validators.required),
      appraisalNote: new FormControl<string | null>(null),
    });
  }

  onSubmit(): void {
    if (this.form.valid) {
      this.dialogRef.close({
        appraisalCode: this.selectedAppraisal,
        appraisalNote: this.form.get('appraisalNote')!.value,
      });
    }
  }
}
