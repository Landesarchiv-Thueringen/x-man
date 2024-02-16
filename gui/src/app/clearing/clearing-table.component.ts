import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Subscription, interval, startWith, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { ClearingDetailsComponent } from './clearing-details.component';
import { ClearingService, ProcessingError } from './clearing.service';

@Component({
  selector: 'app-clearing-table',
  templateUrl: './clearing-table.component.html',
  styleUrls: ['./clearing-table.component.scss'],
})
export class ClearingTableComponent implements AfterViewInit, OnDestroy {
  dataSource: MatTableDataSource<ProcessingError>;
  displayedColumns: string[];
  errorsSubscription?: Subscription;

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private clearingService: ClearingService,
    private dialog: MatDialog,
  ) {
    this.displayedColumns = ['detectedAt', 'agency', 'description', 'actions'];
    this.dataSource = new MatTableDataSource<ProcessingError>();
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
    // refetch errors every `updateInterval` milliseconds
    this.errorsSubscription = interval(environment.updateInterval)
      .pipe(
        startWith(void 0), // initial fetch
        switchMap(() => this.clearingService.getProcessingErrors()),
      )
      .subscribe({
        error: (error: any) => {
          console.error(error);
        },
        next: (errors: ProcessingError[]) => {
          this.dataSource.data = errors;
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
    const dialogRef = this.dialog.open(ClearingDetailsComponent, { data: processingError });
    dialogRef.afterClosed().subscribe((result) => {});
  }
}
