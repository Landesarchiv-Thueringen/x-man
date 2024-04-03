import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { Agency } from '../../../../services/agencies.service';
import { CollectionsService } from '../../../admin-page/collections/collections.service';

export interface StartArchivingDialogData {
  agency: Agency;
}

@Component({
  selector: 'app-start-archiving-dialog',
  templateUrl: './start-archiving-dialog.component.html',
  styleUrls: ['./start-archiving-dialog.component.scss'],
  standalone: true,
  imports: [CommonModule, MatDialogModule, MatButtonModule, MatSelectModule, MatFormFieldModule, ReactiveFormsModule],
})
export class StartArchivingDialogComponent {
  collections = this.collectionsService.getCollections();
  collectionControl = new FormControl(this.data.agency.collectionId, {
    validators: Validators.required,
  });

  constructor(
    private dialogRef: MatDialogRef<StartArchivingDialogComponent>,
    @Inject(MAT_DIALOG_DATA) private data: StartArchivingDialogData,
    private collectionsService: CollectionsService,
  ) {}

  startArchivingProcess() {
    this.dialogRef.close({
      startedArchivingProcess: true,
      collectionId: this.collectionControl.value,
    });
  }
}
