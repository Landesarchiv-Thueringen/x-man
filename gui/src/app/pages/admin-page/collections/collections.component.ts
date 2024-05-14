import { AfterViewInit, Component, TemplateRef, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { CollectionDetailsComponent } from './collection-details.component';
import { ArchiveCollection, CollectionsService } from './collections.service';

@Component({
  selector: 'app-collections',
  standalone: true,
  imports: [MatTableModule, MatButtonModule, MatIconModule, MatSortModule],
  templateUrl: './collections.component.html',
  styleUrl: './collections.component.scss',
})
export class CollectionsComponent implements AfterViewInit {
  @ViewChild('newCollectionDialog') newCollectionDialog!: TemplateRef<unknown>;
  @ViewChild(MatSort) sort!: MatSort;

  dataSource = new MatTableDataSource<ArchiveCollection>();
  displayedColumns: string[] = ['icon', 'name', 'dimagId'];
  newCollectionNameControl = new FormControl('', Validators.required);

  constructor(
    private collectionsService: CollectionsService,
    private dialog: MatDialog,
  ) {
    this.collectionsService
      .observeCollections()
      .pipe(takeUntilDestroyed())
      .subscribe((collections) => (this.dataSource.data = collections));
  }

  ngAfterViewInit(): void {
    this.dataSource.sort = this.sort;
  }

  openDetails(collection: ArchiveCollection) {
    const dialogRef = this.dialog.open(CollectionDetailsComponent, { data: collection });
    dialogRef.afterClosed().subscribe((updatedCollection) => {
      if (updatedCollection) {
        this.collectionsService.updateCollection(collection.id, updatedCollection);
      }
    });
  }

  newCollection() {
    const dialogRef = this.dialog.open(CollectionDetailsComponent);
    dialogRef.afterClosed().subscribe((collection) => {
      if (collection) {
        this.collectionsService.createCollection(collection);
      }
    });
  }
}
