import { Component, TemplateRef, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { CollectionDetailsComponent } from './collection-details.component';
import { Collection, CollectionsService } from './collections.service';

@Component({
  selector: 'app-collections',
  standalone: true,
  imports: [
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatInputModule,
  ],
  templateUrl: './collections.component.html',
  styleUrl: './collections.component.scss',
})
export class CollectionsComponent {
  @ViewChild('newCollectionDialog') newCollectionDialog!: TemplateRef<unknown>;

  dataSource = new MatTableDataSource<Collection>();
  displayedColumns: string[] = ['icon', 'name'];
  newCollectionNameControl = new FormControl('', Validators.required);

  constructor(
    private collectionsService: CollectionsService,
    private dialog: MatDialog,
  ) {
    this.collectionsService
      .getCollections()
      .pipe(takeUntilDestroyed())
      .subscribe((collections) => (this.dataSource.data = collections));
  }

  openDetails(collection: Collection) {
    this.dialog.open(CollectionDetailsComponent, { data: collection });
  }

  newCollection() {
    const dialogRef = this.dialog.open(this.newCollectionDialog, { panelClass: 'new-collection-dialog' });
    dialogRef.afterClosed().subscribe((name) => {
      if (name) {
        this.collectionsService.createCollection({ name });
      }
    });
  }
}
