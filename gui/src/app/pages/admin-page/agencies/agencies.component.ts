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
import { AgenciesService, Agency } from '../../../services/agencies.service';
import { UsersService } from '../../../services/users.service';
import { Collection, CollectionsService } from '../collections/collections.service';
import { AgencyDetailsComponent } from './agency-details.component';

@Component({
  selector: 'app-agencies',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatSortModule, MatButtonModule, MatDialogModule, MatIconModule],
  templateUrl: './agencies.component.html',
  styleUrl: './agencies.component.scss',
})
export class AgenciesComponent implements AfterViewInit {
  @ViewChild(MatSort) sort!: MatSort;

  displayedColumns: string[] = ['icon', 'abbreviation', 'name', 'users', 'collectionId'];
  dataSource = new MatTableDataSource<Agency>();
  agencies = this.agenciesService.getAgencies();
  collections: Collection[] | null = null;

  constructor(
    private agenciesService: AgenciesService,
    private dialog: MatDialog,
    private usersService: UsersService,
    private collectionsService: CollectionsService,
  ) {
    this.agenciesService
      .getAgencies()
      .pipe(takeUntilDestroyed())
      .subscribe((agencies) => (this.dataSource.data = agencies));
    this.collectionsService
      .getCollections()
      .pipe(takeUntilDestroyed())
      .subscribe((collections) => (this.collections = collections));
  }

  ngAfterViewInit() {
    this.dataSource.sort = this.sort;
  }

  getCollectionName(agency: Agency): Observable<string> {
    if (agency.collectionId == null) {
      return of('');
    }
    return this.collectionsService.getCollectionById(agency.collectionId).pipe(map((c) => c?.name ?? ''));
  }

  getUserNames(agency: Agency): Observable<string> {
    return this.usersService
      .getUsersByIds(agency.users.map((user) => user.id))
      .pipe(map((user) => user.map((u) => u.displayName).join('; ')));
  }

  openDetails(agency: Partial<Agency>) {
    const dialogRef = this.dialog.open(AgencyDetailsComponent, { data: agency });
    dialogRef.afterClosed().subscribe((result) => {
      if (result) {
        if (agency.id == null) {
          this.agenciesService.createAgency(result);
        } else {
          this.agenciesService.updateAgency(agency.id, result);
        }
      }
    });
  }

  newAgency() {
    this.openDetails({
      name: 'Neue Abgebende Stelle',
      abbreviation: '',
      transferDirURL: '',
    });
  }
}
