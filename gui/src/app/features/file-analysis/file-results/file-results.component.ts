import { CommonModule } from '@angular/common';
import {
  AfterViewInit,
  Component,
  computed,
  effect,
  inject,
  input,
  viewChild,
} from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import {
  FileDetailsComponent,
  DialogData as FileDetailsData,
} from '../file-details/file-details.component';
import { BreakOpportunitiesPipe } from '../pipes/break-opportunities.pipe';
import { FileAnalysis, FileResult } from '../results';

export type Columns = (typeof columns)[number];

const columns = [
  'path',
  'filename',
  'fileSize',
  'puid',
  'mimeType',
  'formatVersion',
  'status',
] as const;

interface ResultRow {
  id: string;
  filename: string;
  path: string | null;
  fileSize: string | null;
  puid: string | null;
  mimeType: string | null;
  formatVersion: string | null;
  status: Status;
}

interface Status {
  valid: boolean;
  invalid: boolean;
  warning: boolean;
  error: boolean;
}
type FilterKey = keyof Status;
type Filter = { key: FilterKey; label: string; icon: string };

@Component({
  selector: 'app-file-results',
  templateUrl: './file-results.component.html',
  styleUrls: ['./file-results.component.scss'],
  imports: [
    BreakOpportunitiesPipe,
    CommonModule,
    MatButtonModule,
    MatChipsModule,
    MatIconModule,
    MatMenuModule,
    MatPaginatorModule,
    MatSortModule,
    MatTableModule,
  ],
})
export class FileResultsComponent implements AfterViewInit {
  readonly getDetails = input.required<(id: string) => Promise<FileAnalysis | undefined>>();
  readonly results = input<FileResult[]>();
  private resultsMap = computed(() =>
    (this.results() ?? []).reduce(
      (acc, result) => ((acc[result.id] = result), acc),
      {} as { [key: string]: FileResult },
    ),
  );

  private readonly dialog = inject(MatDialog);
  private readonly paginator = viewChild.required(MatPaginator);
  private readonly sort = viewChild.required(MatSort);

  readonly columns = input<Columns[]>();
  displayedColumns = computed(() => this.columns() ?? columns);
  dataSource: MatTableDataSource<ResultRow>;
  activeFilters = new Set<Filter>([]);
  readonly availableFilters: Filter[] = [
    { key: 'valid', label: 'Valide', icon: 'check' },
    { key: 'invalid', label: 'Invalide', icon: 'close' },
    { key: 'warning', label: 'Warnung', icon: 'warning' },
    { key: 'error', label: 'Fehler', icon: 'error' },
  ];

  constructor() {
    this.dataSource = new MatTableDataSource<ResultRow>([]);
    effect(() => this.updateFileResults(this.results() ?? []));
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator();
    this.dataSource.filterPredicate = (data: ResultRow, filter: string) =>
      this.filterPredicate(data, filter as unknown as Set<Filter>);
    this.dataSource.sort = this.sort();
    this.dataSource.sortingDataAccessor = (data: ResultRow, sortHeaderId: string): string => {
      switch (sortHeaderId) {
        case 'status':
          return Object.entries(data.status)
            .filter(([_, value]) => value)
            .map(([key, _]) => key)
            .join();
        default:
          return data[sortHeaderId as keyof ResultRow]
            ? (data[sortHeaderId as keyof ResultRow] as string)
            : '';
      }
    };
  }

  addFilter(filter: Filter) {
    this.activeFilters.add(filter);
    this.dataSource.filter = this.activeFilters as unknown as string;
  }

  removeFilter(filter: Filter) {
    this.activeFilters.delete(filter);
    this.dataSource.filter = this.activeFilters as unknown as string;
  }

  nItemsForFilter(filter: Filter): number {
    return this.dataSource.data.filter((item) => this.filterPredicate(item, new Set([filter])))
      .length;
  }

  private filterPredicate(data: ResultRow, filters?: Set<Filter>): boolean {
    if (!filters || filters.size === 0) {
      return true;
    }
    return [...filters.values()].some((filter) => data.status[filter.key] === true);
  }

  private updateFileResults(results: FileResult[]): void {
    const data: ResultRow[] = [];
    for (let result of results) {
      const row: ResultRow = {
        id: result.id,
        filename: result.filename,
        path: result.additionalMetadata?.['general:path']
          ? (result.additionalMetadata['general:path'].value as string)
          : null,
        fileSize: result.additionalMetadata?.['general:fileSize']
          ? (result.additionalMetadata['general:fileSize'].value as string)
          : null,
        puid: result.summary.puid,
        mimeType: result.summary.mimeType,
        formatVersion: result.summary.formatVersion,
        status: {
          valid: result.summary.valid,
          invalid: result.summary.invalid,
          warning: result.summary.formatUncertain || result.summary.validityConflict,
          error: result.summary.error,
        },
      };
      data.push(row);
    }
    this.dataSource.data = data;
  }

  async openDetails(overview: ResultRow): Promise<void> {
    const result = this.resultsMap()[overview.id];
    const analysis = await this.getDetails()(overview.id);
    if (result && analysis) {
      const data: FileDetailsData = {
        result,
        analysis,
      };
      this.dialog.open(FileDetailsComponent, {
        data: data,
        autoFocus: false,
        width: '80em',
        maxWidth: '80vw',
      });
    } else {
      console.error('file result not found');
    }
  }
}
