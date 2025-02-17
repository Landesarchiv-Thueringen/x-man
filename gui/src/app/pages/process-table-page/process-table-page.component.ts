import { animate, state, style, transition, trigger } from '@angular/animations';
import { DatePipe } from '@angular/common';
import { AfterViewInit, Component, viewChild, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatPaginator, MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { startWith, switchMap, tap } from 'rxjs';
import { Agency } from '../../services/agencies.service';
import { AuthService } from '../../services/auth.service';
import { ConfigService } from '../../services/config.service';
import { ProcessService, ProcessStep, SubmissionProcess } from '../../services/process.service';
import { UpdatesService } from '../../services/updates.service';
import { TaskStateIconComponent } from '../../shared/task-state-icon.component';
import { ProcessStepProgressPipe } from '../message-page/metadata/message-metadata/process-step-progress.pipe';

@Component({
    selector: 'app-process-table-page',
    templateUrl: './process-table-page.component.html',
    styleUrls: ['./process-table-page.component.scss'],
    animations: [
        trigger('expand', [
            state('false', style({
                height: 0,
                paddingTop: 0,
                paddingBottom: 0,
                visibility: 'hidden',
                overflow: 'hidden',
            })),
            transition('false => true', [
                style({ visibility: 'visible' }),
                animate(200, style({ height: '*' })),
            ]),
            transition('true => false', [style({ overflow: 'hidden' }), animate(200)]),
        ]),
    ],
    imports: [
        DatePipe,
        MatButtonModule,
        MatFormFieldModule,
        MatIconModule,
        MatInputModule,
        MatPaginatorModule,
        MatProgressSpinnerModule,
        MatSelectModule,
        MatSlideToggleModule,
        MatSortModule,
        MatTableModule,
        ProcessStepProgressPipe,
        ReactiveFormsModule,
        RouterModule,
        TaskStateIconComponent,
    ]
})
export class ProcessTablePageComponent implements AfterViewInit {
  private authService = inject(AuthService);
  private configService = inject(ConfigService);
  private formBuilder = inject(FormBuilder);
  private processService = inject(ProcessService);
  private updatesService = inject(UpdatesService);

  readonly dataSource: MatTableDataSource<SubmissionProcess> =
    new MatTableDataSource<SubmissionProcess>();
  readonly displayedColumns = [
    'agency',
    'note',
    'message0501',
    'appraisalComplete',
    'message0505',
    'message0503',
    'formatVerification',
    'archivingComplete',
  ] as const;
  readonly stateValues = [
    { value: 'message0501', viewValue: 'Anbietung erhalten' },
    { value: 'appraisalComplete', viewValue: 'Bewertung abgeschlossen' },
    { value: 'message0505', viewValue: 'Bewertung in DMS importiert' },
    { value: 'message0503', viewValue: 'Abgabe erhalten' },
    { value: 'formatVerification', viewValue: 'Formatverifikation abgeschlossen' },
    { value: 'archivingComplete', viewValue: 'Abgabe archiviert' },
  ] as const;
  readonly filter = this.formBuilder.group({
    string: new FormControl(''),
    agency: new FormControl<string | null>(null),
    state: new FormControl('' as ProcessTablePageComponent['stateValues'][number]['value']),
  });
  /** All agencies for which there are processes. Used for agencies filter field. */
  agencies: Agency[] = [];
  isAdmin = this.authService.isAdmin();
  allUsersControl = new FormControl(
    this.isAdmin && window.localStorage.getItem('show-all-user-processes') === 'true',
    {
      nonNullable: true,
    },
  );
  showFilters = window.localStorage.getItem('show-process-filters') === 'true';
  config = this.configService.config;

  readonly paginator = viewChild.required(MatPaginator);
  readonly sort = viewChild.required(MatSort);

  constructor() {
    this.dataSource.sortingDataAccessor = ((
      process: SubmissionProcess,
      property: ProcessTablePageComponent['displayedColumns'][number],
    ) => {
      switch (property) {
        case 'message0501':
          return process.processState.receive0501.completedAt ?? '';
        case 'appraisalComplete':
          return process.processState.appraisal.completedAt ?? '';
        case 'message0505':
          return process.processState.receive0505.completedAt ?? '';
        case 'message0503':
          return process.processState.receive0503.completedAt ?? '';
        case 'formatVerification':
          return process.processState.formatVerification.completedAt ?? '';
        case 'archivingComplete':
          return process.processState.archiving.completedAt ?? '';
        case 'agency':
          return process.agency.name;
        default:
          return process[property] ?? '';
      }
    }) as (data: SubmissionProcess, sortHeaderId: string) => string;
    // We use object instead of string for filter. Hence we cast to the "wrong" types in both assignments below.
    this.filter.valueChanges.subscribe((filter) => (this.dataSource.filter = filter as string));
    this.dataSource.filterPredicate = this.filterPredicate as (
      data: SubmissionProcess,
      filter: string,
    ) => boolean;

    // refetch processes every `updateInterval` milliseconds
    this.allUsersControl.valueChanges
      .pipe(
        tap((allUsers) =>
          window.localStorage.setItem('show-all-user-processes', allUsers.toString()),
        ),
        startWith(this.allUsersControl.value),
        switchMap(() => this.updatesService.observeCollection('submission_processes')),
        switchMap(() => this.processService.getProcesses(this.allUsersControl.value)),
        takeUntilDestroyed(),
      )
      .subscribe((processes: SubmissionProcess[]) => {
        this.dataSource.data = processes ?? [];
        this.populateAgencies(processes ?? []);
      });
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.paginator();
    this.paginator().pageSize = this.getPageSize();
    this.dataSource.sort = this.sort();
  }

  toggleFilters(): void {
    this.showFilters = !this.showFilters;
    window.localStorage.setItem('show-process-filters', this.showFilters.toString());
    if (!this.showFilters) {
      this.filter.setValue({ agency: null, state: null, string: null });
    }
  }

  private getPageSize(): number {
    const savedPageSize = window.localStorage.getItem('main-table-page-size');
    if (savedPageSize) {
      return parseInt(savedPageSize);
    } else {
      return 10;
    }
  }

  onPaginate(event: PageEvent): void {
    window.localStorage.setItem('main-table-page-size', event.pageSize.toString());
  }

  private populateAgencies(processes: SubmissionProcess[]): void {
    this.agencies = [];
    for (const { agency } of processes) {
      if (!this.agencies.some((a) => a.id === agency.id)) {
        this.agencies.push(agency);
      }
    }
  }

  /** The default filter predicate of MatTableDataSource that provides string matching on all data properties. */
  private readonly textFilterPredicate = this.dataSource.filterPredicate;

  /**
   * Custom filter predicate for our process data source.
   *
   * Note that we don't use "string" as type for filter. Instead we provide a
   * filter object and cast types where needed.
   */
  private filterPredicate = (
    process: SubmissionProcess,
    filter: ProcessTablePageComponent['filter']['value'],
  ): boolean => {
    return (
      // Match string field
      this.textFilterPredicate(process, filter.string ?? '') &&
      // Match agency field
      (() => {
        if (filter.agency) {
          return process.agency.id === filter.agency;
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
  private getCurrentState(
    process: SubmissionProcess,
  ): ProcessTablePageComponent['stateValues'][number]['value'] | null {
    for (const state of this.stateValues.map((v) => v.value).reverse()) {
      if (state === 'message0501' && process.processState.receive0501.complete) {
        return state;
      } else if (state === 'appraisalComplete' && process.processState.appraisal.complete) {
        return state;
      } else if (state === 'message0505' && process.processState.receive0505.complete) {
        return state;
      } else if (state === 'message0503' && process.processState.receive0503.complete) {
        return state;
      } else if (
        state === 'formatVerification' &&
        process.processState.formatVerification.complete
      ) {
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
  trackProcess(index: number, item: SubmissionProcess): string {
    return item.processId;
  }

  /**
   * Returns the number of processes that match the given agency.
   *
   * Other filters are applied as normal.
   */
  getElementsForAgency(agency: string | null): number {
    return this.dataSource.data.filter((process) =>
      this.filterPredicate(process, {
        ...(this.dataSource.filter as ProcessTablePageComponent['filter']['value']),
        agency,
      }),
    ).length;
  }

  /**
   * Returns the number of processes that match the given state filter.
   *
   * Other filters are applied as normal.
   */
  getElementsForState(state: ProcessTablePageComponent['stateValues'][number]['value']): number {
    return this.dataSource.data.filter((process) =>
      this.filterPredicate(process, {
        ...(this.dataSource.filter as ProcessTablePageComponent['filter']['value']),
        state,
      }),
    ).length;
  }

  hasUnresolvedError(process: SubmissionProcess): boolean {
    return process.unresolvedErrors > 0;
  }

  getErrorTime(processStep: ProcessStep): string | null {
    return processStep.hasError ? processStep.updatedAt : null;
  }
}
