import { CommonModule } from '@angular/common';
import { Component, TemplateRef, viewChild, inject } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import {
  MAT_DIALOG_DATA,
  MatDialog,
  MatDialogModule,
  MatDialogRef,
} from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { MatSelectModule } from '@angular/material/select';
import { Observable } from 'rxjs';
import { Agency } from '../../../services/agencies.service';
import { ArchiveCollection, CollectionsService } from './collections.service';

/**
 * Collection metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
    selector: 'app-collection-details',
    imports: [
        CommonModule,
        MatButtonModule,
        MatDialogModule,
        MatExpansionModule,
        MatFormFieldModule,
        MatInputModule,
        MatListModule,
        MatSelectModule,
        ReactiveFormsModule,
    ],
    templateUrl: './collection-details.component.html',
    styleUrl: './collection-details.component.scss'
})
export class CollectionDetailsComponent {
  private dialogRef = inject<MatDialogRef<CollectionDetailsComponent>>(MatDialogRef);
  collection = inject<ArchiveCollection>(MAT_DIALOG_DATA);
  private dialog = inject(MatDialog);
  private collectionsService = inject(CollectionsService);

  readonly deleteDialogTemplate = viewChild.required<TemplateRef<unknown>>('deleteDialog');

  readonly isNew = this.collection == null;
  readonly form = new FormGroup({
    name: new FormControl(this.collection?.name ?? 'Neuer Bestand', {
      nonNullable: true,
      validators: Validators.required,
    }),
    dimagId: new FormControl(this.collection?.dimagId, {
      nonNullable: true,
      validators: Validators.required,
    }),
  });
  readonly dimagIds = this.collectionsService.getDimagIds();
  readonly agencies?: Observable<Agency[]>;

  constructor() {
    const collection = this.collection;

    if (collection) {
      this.agencies = this.collectionsService.getAgenciesForCollection(this.collection.id);
    }
  }

  save() {
    const updatedCollection: Omit<ArchiveCollection, 'id'> = this.form.getRawValue();
    this.dialogRef.close(updatedCollection);
  }

  /**
   * Deletes this collection after getting user confirmation and closes the dialog.
   */
  deleteCollection() {
    const dialogRef = this.dialog.open(this.deleteDialogTemplate());
    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.collectionsService.deleteCollection(this.collection);
        this.dialogRef.close();
      }
    });
  }
}
