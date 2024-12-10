import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { Agency } from '../../../../services/agencies.service';
import { ConfigService } from '../../../../services/config.service';
import { PackagingStats } from '../../../../services/packaging.service';
import { CollectionsService } from '../../../admin-page/collections/collections.service';

export interface StartArchivingDialogData {
  agency: Agency;
  packagingStats: PackagingStats;
}

@Component({
    selector: 'app-start-archiving-dialog',
    templateUrl: './start-archiving-dialog.component.html',
    styleUrls: ['./start-archiving-dialog.component.scss'],
    imports: [
        CommonModule,
        MatDialogModule,
        MatButtonModule,
        MatSelectModule,
        MatFormFieldModule,
        ReactiveFormsModule,
    ]
})
export class StartArchivingDialogComponent {
  private dialogRef = inject<MatDialogRef<StartArchivingDialogComponent>>(MatDialogRef);
  private data = inject<StartArchivingDialogData>(MAT_DIALOG_DATA);
  private collectionsService = inject(CollectionsService);
  private configService = inject(ConfigService);

  collectionControl = new FormControl(this.data.agency.collectionId, {
    validators: Validators.required,
  });
  readonly packagingStats = this.data.packagingStats;
  readonly config = this.configService.config;
  readonly collections = toSignal(this.collectionsService.getCollections());

  startArchivingProcess() {
    this.dialogRef.close({
      startedArchivingProcess: true,
      collectionId: this.collectionControl.value,
    });
  }
}
