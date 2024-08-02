import { Component } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { PackagingOption, packagingOptions } from '../../../../services/recordOptions.service';

@Component({
  selector: 'app-packaging-dialog',
  standalone: true,
  imports: [
    MatDialogModule,
    MatSelectModule,
    MatFormFieldModule,
    ReactiveFormsModule,
    MatButtonModule,
  ],
  templateUrl: './packaging-dialog.component.html',
  styleUrl: './packaging-dialog.component.scss',
})
export class PackagingDialogComponent {
  readonly packagingOptions = packagingOptions;
  form = this.formBuilder.group({
    packaging: '' as PackagingOption,
  });

  constructor(
    private dialogRef: MatDialogRef<PackagingDialogComponent>,
    private formBuilder: FormBuilder,
  ) {}

  onSubmit(): void {
    this.dialogRef.close(this.form.value);
  }
}
