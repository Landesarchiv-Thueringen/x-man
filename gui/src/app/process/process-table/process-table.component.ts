import { animate, state, style, transition, trigger } from '@angular/animations';
import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';
import { FormBuilder, FormControl } from '@angular/forms';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Subscription, interval, startWith, switchMap, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Task } from '../../admin/tasks/tasks.service';
import { ProcessingError } from '../../clearing/clearing.service';
import { AuthService } from '../../utility/authorization/auth.service';
import { Process, ProcessService, ProcessStep } from '../process.service';

@Component({
  selector: 'app-process-table',
  templateUrl: './process-table.component.html',
  styleUrls: ['./process-table.component.scss'],
  animations: [
    trigger('expand', [
      state('false', style({ height: 0, paddingTop: 0, paddingBottom: 0, visibility: 'hidden', overflow: 'hidden' })),
      transition('false => true', [style({ visibility: 'visible' }), animate(200, style({ height: '*' }))]),
      transition('true => false', [style({ overflow: 'hidden' }), animate(200)]),
    ]),
  ],
})
export class ProcessTableComponent implements AfterViewInit, OnDestroy {
  readonly dataSource: MatTableDataSource<Process> = new MatTableDataSource<Process>();
  readonly displayedColumns: string[] = [
    'institution',
    'note',
    'message0501',
    'appraisalComplete',
    'received0502',
    'message0503',
    'formatVerification',
    'archivingComplete',
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
    institution: new FormControl(''),
    state: new FormControl('' as ProcessTableComponent['stateValues'][number]['value']),
  });
  /** All institutions for which there are processes. Used for institutions filter field. */
  institutions: string[] = [];
  isAdmin = this.authService.isAdmin();
  allUsersControl = new FormControl(window.localStorage.getItem('show-all-user-processes') === 'true', {
    nonNullable: true,
  });
  showFilters = window.localStorage.getItem('show-process-filters') === 'true';

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private processService: ProcessService,
    private formBuilder: FormBuilder,
    private authService: AuthService,
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
    this.processSubscription = this.allUsersControl.valueChanges
      .pipe(
        tap((allUsers) => window.localStorage.setItem('show-all-user-processes', allUsers.toString())),
        startWith(this.allUsersControl.value),
        switchMap(() =>
          interval(environment.updateInterval).pipe(
            startWith(void 0), // initial fetch
          ),
        ),
        switchMap(() => this.processService.getProcesses(this.allUsersControl.value)),
      )
      .subscribe({
        error: (error) => {
          console.error(error);
        },
        next: (processes: Process[]) => {
          this.dataSource.data = processes;
          this.populateInstitutions(processes);
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

  toggleFilters(): void {
    this.showFilters = !this.showFilters;
    window.localStorage.setItem('show-process-filters', this.showFilters.toString());
    if (!this.showFilters) {
      this.filter.setValue({ institution: null, state: null, string: null });
    }
  }

  private populateInstitutions(processes: Process[]): void {
    this.institutions = [...new Set(processes.map((p) => p.institution))];
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
      // Match institution field
      (() => {
        if (filter.institution) {
          return process.institution === filter.institution;
        } else {
          return true;
        }
      })() &&
      // Match state field
      (() => {
        if (filter.state) {
          return filter.state === this.getCurrentState(process);
        } else {
          return true;
        }
      })()
    );
  };

  /** Returns the highest process state that the process completed. */
  private getCurrentState(process: Process): ProcessTableComponent['stateValues'][number]['value'] | null {
    for (const state of this.stateValues.map((v) => v.value).reverse()) {
      if (state === 'message0501' && process.processState.receive0501.complete) {
        return state;
      } else if (state === 'appraisalComplete' && process.processState.appraisal.complete) {
        return state;
      } else if (state === 'received0502' && process.processState.receive0505.complete) {
        return state;
      } else if (state === 'message0503' && process.processState.receive0503.complete) {
        return state;
      } else if (state === 'formatVerification' && process.processState.formatVerification.complete) {
        return state;
      } else if (state === 'archivingComplete' && process.processState.archiving.complete) {
        return state;
      }
    }
    return null;
  }

  /**
   * Implements a track-by predicate for Angular to match table rows on data updates.
   */
  trackProcess(index: number, item: Process): string {
    return item.id;
  }

  /**
   * Returns the number of processes that match the given institution.
   *
   * Other filters are applied as normal.
   */
  getElementsForInstitution(institution: string): number {
    return this.dataSource.data.filter((process) =>
      this.filterPredicate(process, {
        ...(this.dataSource.filter as ProcessTableComponent['filter']['value']),
        institution,
      }),
    ).length;
  }

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

  hasUnresolvedError(process: Process): boolean {
    return process.processingErrors.some((processingError) => !processingError.resolved);
  }

  isStepFailed(processStep: ProcessStep): boolean {
    return this.getUnresolvedErrors(processStep).length > 0;
  }

  getUnresolvedErrors(processStep: ProcessStep): ProcessingError[] {
    return processStep.processingErrors.filter((processingError) => !processingError.resolved);
  }

  getErrorTime(processStep: ProcessStep): string | null {
    return this.getUnresolvedErrors(processStep)[0]?.detectedAt;
  }

  getRunningTask(processStep: ProcessStep): Task | null {
    return processStep.tasks.find((task) => task.state === 'running') ?? null;
  }
}
