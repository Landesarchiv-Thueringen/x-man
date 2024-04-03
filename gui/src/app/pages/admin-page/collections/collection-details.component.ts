import { CommonModule } from '@angular/common';
import { Component, Inject, TemplateRef, ViewChild } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { Observable } from 'rxjs';
import { Agency } from '../../../services/agencies.service';
import { Collection, CollectionsService } from './collections.service';

/**
 * Collection metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-collection-details',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatInputModule,
    MatListModule,
    ReactiveFormsModule,
  ],
  templateUrl: './collection-details.component.html',
  styleUrl: './collection-details.component.scss',
})
export class CollectionDetailsComponent {
  @ViewChild('deleteDialog') deleteDialogTemplate!: TemplateRef<unknown>;

  readonly isNew = this.collection == null;
  form = new FormGroup({
    name: new FormControl(this.collection?.name ?? 'Neuer Bestand', {
      nonNullable: true,
      validators: Validators.required,
    }),
    dimagId: new FormControl(this.collection?.dimagId, {
      nonNullable: true,
      validators: Validators.required,
    }),
  });
  institutions?: Observable<Agency[]>;

  constructor(
    private dialogRef: MatDialogRef<CollectionDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public collection: Collection,
    private dialog: MatDialog,
    private collectionsService: CollectionsService,
  ) {
    if (collection) {
      this.institutions = this.collectionsService.getInstitutionsForCollection(this.collection.id);
    }
  }

  save() {
    const updatedCollection: Omit<Collection, 'id'> = this.form.getRawValue();
    this.dialogRef.close(updatedCollection);
  }

  /**
   * Deletes this collection after getting user confirmation and closes the dialog.
   */
  deleteCollection() {
    const dialogRef = this.dialog.open(this.deleteDialogTemplate);
    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.collectionsService.deleteCollection(this.collection);
        this.dialogRef.close();
      }
    });
  }
}
