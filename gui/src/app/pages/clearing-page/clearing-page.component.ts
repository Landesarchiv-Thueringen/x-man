import { DatePipe } from '@angular/common';
import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator, MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { startWith, switchMap, tap } from 'rxjs';
import { ClearingService, ProcessingError } from '../../services/clearing.service';
import { ClearingDetailsComponent } from './clearing-details.component';

@Component({
  selector: 'app-clearing-page',
  templateUrl: './clearing-page.component.html',
  styleUrls: ['./clearing-page.component.scss'],
  standalone: true,
  imports: [
    DatePipe,
    MatButtonModule,
    MatPaginatorModule,
    MatSlideToggleModule,
    MatSortModule,
    MatTableModule,
    ReactiveFormsModule,
  ],
})
export class ClearingPageComponent implements AfterViewInit {
  dataSource: MatTableDataSource<ProcessingError>;
  displayedColumns: string[];
  showResolvedControl = new FormControl(window.localStorage.getItem('show-resolved-processing-errors') === 'true', {
    nonNullable: true,
  });
  lastSeenTime = this.clearingService.getLastSeenTime();

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private clearingService: ClearingService,
    private dialog: MatDialog,
  ) {
    this.displayedColumns = ['createdAt', 'agency', 'title'];
    this.dataSource = new MatTableDataSource<ProcessingError>();
    this.clearingService.markAllSeen();

    this.showResolvedControl.valueChanges
      .pipe(
        tap((showResolved) => window.localStorage.setItem('show-resolved-processing-errors', showResolved.toString())),
        startWith(this.showResolvedControl.value),
        switchMap(() => this.clearingService.observeProcessingErrors()),
        takeUntilDestroyed(),
      )
      .subscribe((errors: ProcessingError[]) => {
        // Since we cannot expect the server and client clocks to be in sync on
        // the millisecond, we take the time from the most recent error.
        this.clearingService.markAllSeen(this.getMostRecentErrorTime(errors));
        if (this.showResolvedControl.value) {
          this.dataSource.data = errors;
        } else {
          this.dataSource.data = errors.filter((error) => !error.resolved);
        }
      });
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.paginator.pageSize = this.getPageSize();
    this.dataSource.sort = this.sort;
  }

  trackTableRow(index: number, element: ProcessingError): string {
    return element.id;
  }

  isNew(node: ProcessingError) {
    return new Date(node.createdAt).valueOf() > this.lastSeenTime;
  }

  openDetails(processingError: Partial<ProcessingError>) {
    const dialogRef = this.dialog.open(ClearingDetailsComponent, { maxWidth: '80vw', data: processingError });
    dialogRef.afterClosed().subscribe((result) => {});
  }

  onPaginate(event: PageEvent): void {
    window.localStorage.setItem('main-table-page-size', event.pageSize.toString());
  }

  private getPageSize(): number {
    const savedPageSize = window.localStorage.getItem('main-table-page-size');
    if (savedPageSize) {
      return parseInt(savedPageSize);
    } else {
      return 10;
    }
  }

  private getMostRecentErrorTime(errors: ProcessingError[]): number {
    let result = 0;
    for (const e of errors) {
      const errorTime = new Date(e.createdAt).valueOf();
      if (errorTime > result) {
        result = errorTime;
      }
    }
    return result;
  }
}
