import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';
import { Collection, CollectionsService } from '../collections/collections.service';
import { UsersService } from '../users/users.service';
import { InstitutionDetailsComponent } from './institution-details.component';
import { Institution, InstitutionsService } from './institutions.service';
import { TransferDirectory } from './transfer-directory.service';

@Component({
  selector: 'app-institutions',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatSortModule, MatButtonModule, MatDialogModule, MatIconModule],
  templateUrl: './institutions.component.html',
  styleUrl: './institutions.component.scss',
})
export class InstitutionsComponent implements AfterViewInit {
  @ViewChild(MatSort) sort!: MatSort;

  displayedColumns: string[] = ['abbreviation', 'name', 'users', 'collectionId'];
  dataSource = new MatTableDataSource<Institution>();
  institutions = this.institutionsService.getInstitutions();
  collections: Collection[] | null = null;

  constructor(
    private institutionsService: InstitutionsService,
    private dialog: MatDialog,
    private usersService: UsersService,
    private collectionsService: CollectionsService,
  ) {
    this.institutionsService
      .getInstitutions()
      .pipe(takeUntilDestroyed())
      .subscribe((institutions) => (this.dataSource.data = institutions));
    this.collectionsService
      .getCollections()
      .pipe(takeUntilDestroyed())
      .subscribe((collections) => (this.collections = collections));
  }

  ngAfterViewInit() {
    this.dataSource.sort = this.sort;
  }

  getCollectionName(institution: Institution): Observable<string> {
    if (institution.collectionId == null) {
      return of('');
    }
    return this.collectionsService.getCollectionById(institution.collectionId).pipe(map((c) => c?.name ?? ''));
  }

  getArchivistNames(institution: Institution): Observable<string> {
    return this.usersService
      .getUsersByIds(institution.userIds)
      .pipe(map((user) => user.map((u) => u.displayName).join('; ')));
  }

  openDetails(institution: Partial<Institution>) {
    const dialogRef = this.dialog.open(InstitutionDetailsComponent, { data: institution });
    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        if (institution.id == null) {
          this.institutionsService.createInstitution(result);
        } else {
          this.institutionsService.updateInstitution(institution.id, result);
        }
      }
    });
  }

  newInstitution() {
    this.openDetails({
      name: 'Neue Abgebende Stelle',
      abbreviation: '',
      transferDirectory: {} as TransferDirectory,
      userIds: [],
    });
  }
}
