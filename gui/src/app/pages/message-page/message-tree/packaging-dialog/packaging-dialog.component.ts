import { Component, Inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { firstValueFrom } from 'rxjs';
import { packagingChoices, PackagingService } from '../../../../services/packaging.service';
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
  packagingChoices = [...packagingChoices.map((option) => ({ ...option }))];
  form = this.formBuilder.group({
    packaging: 'root',
  });

  constructor(
    private dialogRef: MatDialogRef<PackagingDialogComponent>,
    @Inject(MAT_DIALOG_DATA) private data: PackagingDialogData,
    private formBuilder: FormBuilder,
    private packagingService: PackagingService,
  ) {
    this.populatePackagingChoices();
  }

  onSubmit(): void {
    this.dialogRef.close(this.form.value);
  }

  private async populatePackagingChoices() {
    const statsMap = await firstValueFrom(
      this.packagingService.getPackagingStats(this.data.processId, this.data.recordIds),
    );
    for (const choice of this.packagingChoices) {
      choice.disabled = !statsMap[choice.value].deepestLevelHasItems;
      choice.label = choice.label + ` (${printPackagingStats(statsMap[choice.value])})`;
    }
  }
}
