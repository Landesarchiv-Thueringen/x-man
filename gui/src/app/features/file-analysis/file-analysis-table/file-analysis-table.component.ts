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
import { FileOverviewComponent } from '../file-overview/file-overview.component';
import { BreakOpportunitiesPipe } from '../pipes/break-opportunities.pipe';
import { FileFeaturePipe } from '../pipes/file-feature.pipe';
import { FileAnalysis, FileResult, RowValue } from '../results';

export interface StatusIcons {
  valid: boolean;
  invalid: boolean;
  warning: boolean;
  error: boolean;
}

type FileOverview = {
  id: string;
  values: { [key in string]?: RowValue };
  icons: StatusIcons;
};

type FilterKey = keyof StatusIcons;
type Filter = { key: FilterKey; label: string; icon: string };

/** Properties to show in the table and / or the details dialog. */
export interface FilePropertyDefinition {
  /**
   * The property key.
   *
   * Either `filename`, a tool-result property, or a field of `info`.
   */
  key: string;
  /**
   * A label to be used column header.
   *
   * Can be omitted for native properties (`filename`, tool results).
   */
  label?: string;
  /**
   * Whether to show the property in the table. Default: true.
   *
   * When false, the property will be shown only in the details dialog.
   */
  inTable?: boolean;
}

@Component({
  selector: 'app-file-analysis-table',
  templateUrl: './file-analysis-table.component.html',
  styleUrls: ['./file-analysis-table.component.scss'],
  imports: [
    BreakOpportunitiesPipe,
    CommonModule,
    FileFeaturePipe,
    MatButtonModule,
    MatChipsModule,
    MatIconModule,
    MatMenuModule,
    MatPaginatorModule,
    MatSortModule,
    MatTableModule,
  ],
})
export class FileAnalysisTableComponent implements AfterViewInit {
  private dialog = inject(MatDialog);

  readonly results = input<FileResult[]>();
  readonly getDetails = input.required<(id: string) => Promise<FileAnalysis | undefined>>();
  /**
   * Properties to be displayed in the table and in the details dialog.
   *
   * Available values are combined from
   *  - filename
   *  - puid, mimeType, and formatVersion from tool results
   *  - any field provided with the `info` property on file results
   *
   * You have to provide labels for columns based on properties from `info`.
   */
  readonly properties = input<FilePropertyDefinition[]>([
    { key: 'path', label: 'Pfad' },
    { key: 'filename' },
    { key: 'fileSize', label: 'Dateigröße' },
    { key: 'puid' },
    { key: 'mimeType' },
    { key: 'formatVersion' },
    { key: 'status' },
  ]);

  readonly paginator = viewChild.required(MatPaginator);
  readonly sort = viewChild.required(MatSort);

  private resultsMap = computed(() =>
    (this.results() ?? []).reduce(
      (acc, result) => ((acc[result.id] = result), acc),
      {} as { [key: string]: FileResult },
    ),
  );
  readonly tableProperties = computed(() => this.properties().filter((p) => p.inTable !== false));
  readonly columns = computed(() => this.tableProperties().map(({ key }) => key));

  dataSource: MatTableDataSource<FileOverview>;
  activeFilters = new Set<Filter>([]);
  readonly availableFilters: Filter[] = [
    { key: 'valid', label: 'Valide', icon: 'check' },
    { key: 'invalid', label: 'Invalide', icon: 'close' },
    { key: 'warning', label: 'Warnung', icon: 'warning' },
    { key: 'error', label: 'Fehler', icon: 'error' },
  ];

  constructor() {
    this.dataSource = new MatTableDataSource<FileOverview>([]);
    effect(() => this.processFileInformation(this.results() ?? []));
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator();
    this.dataSource.filterPredicate = (data: FileOverview, filter: string) =>
      this.filterPredicate(data, filter as unknown as Set<Filter>);
    this.dataSource.sort = this.sort();
    this.dataSource.sortingDataAccessor = (data, sortHeaderId) => {
      switch (sortHeaderId) {
        case 'status':
          return Object.entries(data.icons)
            .filter(([_, value]) => value)
            .map(([key, _]) => key)
            .join();
        default:
          return data.values[sortHeaderId]?.value ?? '';
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

  async openDetails(overview: FileOverview): Promise<void> {
    const analysis = await this.getDetails()(overview.id);
    if (analysis) {
      this.dialog.open(FileOverviewComponent, {
        data: {
          filename: overview.values['filename']!.value,
          info: this.resultsMap()[overview.id].info,
          analysis,
          properties: this.properties(),
        },
        autoFocus: false,
        width: '1200px',
        maxWidth: '80vw',
      });
    } else {
      console.error('file result not found');
    }
  }

  private filterPredicate(data: FileOverview, filters?: Set<Filter>): boolean {
    if (!filters || filters.size === 0) {
      return true;
    }
    return [...filters.values()].some((filter) => data.icons[filter.key] === true);
  }

  private processFileInformation(results: FileResult[]): void {
    const data: FileOverview[] = [];
    for (let result of results) {
      let fileOverview: FileOverview = {
        id: result.id,
        icons: {
          valid: result.summary.valid,
          invalid: result.summary.invalid,
          warning: result.summary.formatUncertain || result.summary.validityConflict,
          error: result.summary.error,
        },
        values: {
          filename: { value: result.filename },
          puid: { value: result.summary.puid },
          mimeType: { value: result.summary.mimeType },
          formatVersion: { value: result.summary.formatVersion },
          ...result.info,
        },
      };
      data.push(fileOverview);
    }
    this.dataSource.data = data;
  }
}
