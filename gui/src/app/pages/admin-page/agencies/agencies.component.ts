import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, effect, viewChild, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { Observable, of } from 'rxjs';
import { map } from 'rxjs/operators';
import { AgenciesService, Agency } from '../../../services/agencies.service';
import { ConfigService } from '../../../services/config.service';
import { UsersService } from '../../../services/users.service';
import { ArchiveCollection, CollectionsService } from '../collections/collections.service';
import { AgencyDetailsComponent } from './agency-details.component';

@Component({
    selector: 'app-agencies',
    imports: [
        CommonModule,
        MatButtonModule,
        MatDialogModule,
        MatIconModule,
        MatSortModule,
        MatTableModule,
    ],
    templateUrl: './agencies.component.html',
    styleUrl: './agencies.component.scss'
})
export class AgenciesComponent implements AfterViewInit {
  private agenciesService = inject(AgenciesService);
  private dialog = inject(MatDialog);
  private usersService = inject(UsersService);
  private collectionsService = inject(CollectionsService);
  private configService = inject(ConfigService);

  readonly sort = viewChild.required(MatSort);

  displayedColumns: string[] = ['icon', 'abbreviation', 'name', 'users'];
  dataSource = new MatTableDataSource<Agency>();
  agencies = this.agenciesService.observeAgencies();
  collections: ArchiveCollection[] | null = null;

  constructor() {
    this.agenciesService
      .observeAgencies()
      .pipe(takeUntilDestroyed())
      .subscribe((agencies) => (this.dataSource.data = agencies));
    this.collectionsService
      .observeCollections()
      .pipe(takeUntilDestroyed())
      .subscribe((collections) => (this.collections = collections));
    effect(() => {
      if (this.configService.config()?.archiveTarget === 'dimag') {
        this.displayedColumns.push('collectionId');
      }
    });
  }

  ngAfterViewInit() {
    this.dataSource.sort = this.sort();
  }

  getCollectionName(agency: Agency): Observable<string> {
    if (agency.collectionId == null) {
      return of('');
    }
    return this.collectionsService
      .getCollectionById(agency.collectionId)
      .pipe(map((c) => c?.name ?? ''));
  }

  getUserNames(agency: Agency): Observable<string> {
    return this.usersService
      .getUsersByIds(agency.users ?? [])
      .pipe(map((user) => user.map((u) => u.displayName).join('; ')));
  }

  openDetails(agency: Partial<Agency>) {
    const dialogRef = this.dialog.open(AgencyDetailsComponent, { data: agency, maxWidth: '80vw' });
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
