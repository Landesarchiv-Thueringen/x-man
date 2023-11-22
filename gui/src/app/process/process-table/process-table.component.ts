// angular
import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Process, ProcessService } from '../process.service';

// utility
import { interval, switchMap, Subscription } from 'rxjs';

@Component({
  selector: 'app-process-table',
  templateUrl: './process-table.component.html',
  styleUrls: ['./process-table.component.scss'],
})
export class ProcessTableComponent implements AfterViewInit, OnDestroy {
  dataSource: MatTableDataSource<Process>;
  displayedColumns: string[];
  processSubscription: Subscription;

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(private processService: ProcessService) {
    this.displayedColumns = [
      'institution',
      'message0501',
      'appraisalComplete',
      'received0502',
      'message0503',
      'formatVerification',
      'archivingComplete',
      'actions',
    ];
    this.dataSource = new MatTableDataSource<Process>();
    this.dataSource.sortingDataAccessor = (item: Process, property: string) => {
      switch (property) {
        case 'receivedAt':
          return item.receivedAt;
        case 'institution':
          return item.institution ? item.institution : '';
        case 'message0501':
          return (!!item.message0501).toString();
        case 'appraisalComplete':
          return item.message0501
            ? item.message0501.appraisalComplete.toString()
            : (!!item.message0501).toString();
        case 'received0502':
          return (!!item.message0501 && !!item.message0503).toString()
        case 'message0503':
          return (!!item.message0503).toString();
        case 'archivingComplete':
          return (!!item.archivingComplete).toString();
        default:
          throw new Error('sorting error: unhandled column');
      }
    };
    this.processService.getProcesses().subscribe({
      error: (error) => {
        console.error(error);
      },
      next: (processes: Process[]) => {
        console.log(processes);
        this.dataSource.data = processes;
      },
    });
    // refetch processes every 30 seconds
    this.processSubscription = interval(30000)
      .pipe(switchMap(() => this.processService.getProcesses()))
      .subscribe({
        error: (error) => {
          console.error(error);
        },
        next: (processes: Process[]) => {
          console.log(processes);
          this.dataSource.data = processes;
        },
      });
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  ngOnDestroy(): void {
    this.processSubscription.unsubscribe();
  }
}
