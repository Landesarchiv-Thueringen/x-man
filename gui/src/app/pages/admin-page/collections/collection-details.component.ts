import { CommonModule } from '@angular/common';
import { Component, Inject, TemplateRef, ViewChild } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatListModule } from '@angular/material/list';
import { Collection, CollectionsService } from './collections.service';

/**
 * Collection metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-collection-details',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatListModule, MatDialogModule],
  templateUrl: './collection-details.component.html',
  styleUrl: './collection-details.component.scss',
})
export class CollectionDetailsComponent {
  @ViewChild('deleteDialog') deleteDialogTemplate!: TemplateRef<unknown>;
  institutions = this.collectionsService.getInstitutionsForCollection(this.collection.id);

  constructor(
    private dialogRef: MatDialogRef<CollectionDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) public collection: Collection,
    private dialog: MatDialog,
    private collectionsService: CollectionsService,
  ) {}

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
