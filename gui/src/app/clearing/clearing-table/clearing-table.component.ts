// project
import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { ClearingService, ProcessingError } from '../clearing.service';
import { environment } from '../../../environments/environment';

// utility
import { interval, switchMap, Subscription, startWith } from 'rxjs';

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

  constructor(private clearingService: ClearingService) {
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
          if (JSON.stringify(this.dataSource.data) !== JSON.stringify(errors)) {
            console.log('Updated errors', errors);
            this.dataSource.data = errors;
          }
        },
      });
  }

  ngOnDestroy(): void {
    this.errorsSubscription?.unsubscribe();
  }
}
