// project
import { Component, OnDestroy, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { ClearingService, ProcessingError } from '../clearing.service';

// utility
import { interval, switchMap, Subscription } from 'rxjs';

@Component({
  selector: 'app-clearing-table',
  templateUrl: './clearing-table.component.html',
  styleUrls: ['./clearing-table.component.scss'],
})
export class ClearingTableComponent implements OnDestroy {
  dataSource: MatTableDataSource<ProcessingError>;
  displayedColumns: string[];
  errorsSubscription: Subscription;

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(private clearingService: ClearingService) {
    this.displayedColumns = [
      'detectedAt',
      'description',
      'transferDirPath',
      'messageStorePath',
      'actions',
    ];
    this.dataSource = new MatTableDataSource<ProcessingError>();
    this.clearingService.getProcessingErrors().subscribe({
      error: (error: any) => {
        console.error(error);
      },
      next: (errors: ProcessingError[]) => {
        this.dataSource.data = errors
      },
    });
    // refetch errors every minute
    this.errorsSubscription = interval(60000)
    .pipe(switchMap(() => this.clearingService.getProcessingErrors()))
    .subscribe({
      error: (error: any) => {
        console.error(error);
      },
      next: (errors: ProcessingError[]) => {
        console.log(errors);
        this.dataSource.data = errors
      },
    });
  }

  ngOnDestroy(): void {
    this.errorsSubscription.unsubscribe();
  }
}
