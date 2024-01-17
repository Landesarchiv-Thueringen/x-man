import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Subscription, interval, startWith, switchMap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Process, ProcessService } from '../process.service';

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
      // TODO: fix
      switch (property) {
        case 'receivedAt':
          return item.receivedAt;
        case 'institution':
          return item.institution ? item.institution : '';
        case 'message0501':
          return (!!item.message0501).toString();
        case 'appraisalComplete':
          return item.message0501 ? item.message0501.appraisalComplete.toString() : (!!item.message0501).toString();
        case 'received0502':
          return (!!item.message0501 && !!item.message0503).toString();
        case 'message0503':
          return (!!item.message0503).toString();
        case 'archivingComplete':
          return (!!item.processState.archiving.complete).toString();
        default:
          throw new Error('sorting error: unhandled column');
      }
    };
    // refetch processes every `updateInterval` milliseconds
    this.processSubscription = interval(environment.updateInterval)
      .pipe(
        startWith(void 0), // initial fetch
        switchMap(() => this.processService.getProcesses()),
      )
      .subscribe({
        error: (error) => {
          console.error(error);
        },
        next: (processes: Process[]) => {
          if (JSON.stringify(this.dataSource.data) !== JSON.stringify(processes)) {
            console.log('Updated processes', processes);
            this.dataSource.data = processes;
          }
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

  downloadReport(process: Process) {
    this.processService.getReport(process.xdomeaID).subscribe((report) => {
      const a = document.createElement('a');
      document.body.appendChild(a);
      a.download = `Übernahmebericht ${process.agency.abbreviation} ${process.receivedAt}.pdf`;
      a.href = window.URL.createObjectURL(report);
      a.click();
      document.body.removeChild(a);
    });
  }
}
