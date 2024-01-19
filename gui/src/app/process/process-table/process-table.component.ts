import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';
import { FormBuilder, FormControl } from '@angular/forms';
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
  readonly dataSource: MatTableDataSource<Process> = new MatTableDataSource<Process>();
  readonly displayedColumns: string[] = [
    'institution',
    'message0501',
    'appraisalComplete',
    'received0502',
    'message0503',
    'formatVerification',
    'archivingComplete',
    'actions',
  ];
  readonly processSubscription: Subscription;
  readonly stateValues = [
    { value: 'message0501', viewValue: 'Anbietung erhalten' },
    { value: 'appraisalComplete', viewValue: 'Bewertung abgeschlossen' },
    { value: 'received0502', viewValue: 'Bewertung in VIS importiert' },
    { value: 'message0503', viewValue: 'Abgabe erhalten' },
    { value: 'formatVerification', viewValue: 'Formatverifikation abgeschlossen' },
    { value: 'archivingComplete', viewValue: 'Abgabe archiviert' },
  ] as const;
  readonly filter = this.formBuilder.group({
    string: new FormControl(''),
    state: new FormControl('' as ProcessTableComponent['stateValues'][number]['value']),
  });

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private processService: ProcessService,
    private formBuilder: FormBuilder,
  ) {
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
    // We use object instead of string for filter. Hence we cast to the "wrong" types in both assignments below.
    this.filter.valueChanges.subscribe((filter) => (this.dataSource.filter = filter as string));
    this.dataSource.filterPredicate = this.filterPredicate as (data: Process, filter: string) => boolean;

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

  /** The default filter predicate of MatTableDataSource that provides string matching on all data properties. */
  private readonly textFilterPredicate = this.dataSource.filterPredicate;

  /**
   * Custom filter predicate for our process data source.
   *
   * Note that we don't use "string" as type for filter. Instead we provide a
   * filter object and cast types where needed.
   */
  private filterPredicate = (process: Process, filter: ProcessTableComponent['filter']['value']): boolean => {
    return (
      // Match string field
      this.textFilterPredicate(process, filter.string ?? '') &&
      // Match state field
      (() => {
        switch (filter.state) {
          case 'message0501':
            return process.processState.receive0501.complete && !process.processState.appraisal.complete;
          case 'appraisalComplete':
            return (
              process.processState.appraisal.complete &&
              !process.processState.receive0505.complete &&
              !process.processState.receive0503.complete
            );
          case 'received0502':
            return process.processState.receive0505.complete && !process.processState.receive0503.complete;
          case 'message0503':
            return process.processState.receive0503.complete && !process.processState.formatVerification.complete;
          case 'formatVerification':
            return process.processState.formatVerification.complete && !process.processState.archiving.complete;
          case 'archivingComplete':
            return process.processState.archiving.complete;
          default:
            return true;
        }
      })()
    );
  };

  /**
   * Returns the number of processes that match the given state filter.
   *
   * Other filters are applied as normal.
   */
  getElementsForState(state: ProcessTableComponent['stateValues'][number]['value']): number {
    return this.dataSource.data.filter((process) =>
      this.filterPredicate(process, { ...(this.dataSource.filter as ProcessTableComponent['filter']['value']), state }),
    ).length;
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
