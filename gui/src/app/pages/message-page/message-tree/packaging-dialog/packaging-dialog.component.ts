import { Component, Inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { firstValueFrom } from 'rxjs';
import { packagingOptions, PackagingService } from '../../../../services/packaging.service';
import { printPackagingStats } from '../../packaging-stats.pipe';

export interface PackagingDialogData {
  processId: string;
  recordIds: string[];
}

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
  packagingOptions = [...packagingOptions.map((option) => ({ ...option }))];
  form = this.formBuilder.group({
    packaging: 'root',
  });

  constructor(
    private dialogRef: MatDialogRef<PackagingDialogComponent>,
    @Inject(MAT_DIALOG_DATA) private data: PackagingDialogData,
    private formBuilder: FormBuilder,
    private packagingService: PackagingService,
  ) {
    this.populatePackagingOptions();
  }

  onSubmit(): void {
    this.dialogRef.close(this.form.value);
  }

  private async populatePackagingOptions() {
    const statsMap = await firstValueFrom(
      this.packagingService.getPackagingStats(this.data.processId, this.data.recordIds),
    );
    for (const option of this.packagingOptions) {
      option.disabled = !statsMap[option.value].deepestLevelHasItems;
      option.label = option.label + ` (${printPackagingStats(statsMap[option.value])})`;
    }
  }
}
