import { Component, inject } from '@angular/core';
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
    imports: [
        MatDialogModule,
        MatSelectModule,
        MatFormFieldModule,
        ReactiveFormsModule,
        MatButtonModule,
    ],
    templateUrl: './packaging-dialog.component.html',
    styleUrl: './packaging-dialog.component.scss'
})
export class PackagingDialogComponent {
  private dialogRef = inject<MatDialogRef<PackagingDialogComponent>>(MatDialogRef);
  private data = inject<PackagingDialogData>(MAT_DIALOG_DATA);
  private formBuilder = inject(FormBuilder);
  private packagingService = inject(PackagingService);

  packagingChoices = [...packagingChoices.map((option) => ({ ...option }))];
  form = this.formBuilder.group({
    packaging: 'root',
  });

  constructor() {
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
