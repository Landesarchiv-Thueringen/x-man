import { DatePipe } from '@angular/common';
import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator, MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { Subscription, interval, startWith, switchMap, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
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
export class ClearingPageComponent implements AfterViewInit, OnDestroy {
  dataSource: MatTableDataSource<ProcessingError>;
  displayedColumns: string[];
  errorsSubscription?: Subscription;
  showResolvedControl = new FormControl(window.localStorage.getItem('show-resolved-processing-errors') === 'true', {
    nonNullable: true,
  });

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private clearingService: ClearingService,
    private dialog: MatDialog,
  ) {
    this.displayedColumns = ['detectedAt', 'agency', 'description'];
    this.dataSource = new MatTableDataSource<ProcessingError>();
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.paginator.pageSize = this.getPageSize();
    this.dataSource.sort = this.sort;

    // refetch errors every `updateInterval` milliseconds
    this.errorsSubscription = this.showResolvedControl.valueChanges
      .pipe(
        tap((showResolved) => window.localStorage.setItem('show-resolved-processing-errors', showResolved.toString())),
        startWith(this.showResolvedControl.value),
        switchMap(() =>
          interval(environment.updateInterval).pipe(
            startWith(void 0), // initial fetch
          ),
        ),
        switchMap(() => this.clearingService.getProcessingErrors()),
      )
      .subscribe({
        error: (error: any) => {
          console.error(error);
        },
        next: (errors: ProcessingError[]) => {
          if (this.showResolvedControl.value) {
            this.dataSource.data = errors;
          } else {
            this.dataSource.data = errors.filter((error) => !error.resolved);
          }
        },
      });
  }

  ngOnDestroy(): void {
    this.errorsSubscription?.unsubscribe();
  }

  trackTableRow(index: number, element: ProcessingError): number {
    return element.id;
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
}
