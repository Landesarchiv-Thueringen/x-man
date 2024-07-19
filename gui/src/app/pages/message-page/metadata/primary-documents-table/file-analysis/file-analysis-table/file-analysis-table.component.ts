import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, Input, ViewChild } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { FileOverviewComponent } from '../file-overview/file-overview.component';
import { FileFeaturePipe } from '../pipes/file-feature.pipe';
import { FileResult, RowValue } from '../results';
import { StatusIcons, StatusIconsService } from '../status-icons.service';

type FileOverview = {
  id: string;
  values: { [key in string]?: RowValue };
  icons: StatusIcons;
};

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
  standalone: true,
  imports: [
    CommonModule,
    FileFeaturePipe,
    MatButtonModule,
    MatIconModule,
    MatPaginatorModule,
    MatSortModule,
    MatTableModule,
  ],
})
export class FileAnalysisTableComponent implements AfterViewInit {
  dataSource: MatTableDataSource<FileOverview>;

  private _results?: FileResult[];
  @Input()
  get results(): FileResult[] | undefined {
    return this._results;
  }
  set results(value: FileResult[] | undefined) {
    this._results = value;
    this.processFileInformation(value ?? []);
  }

  @Input() getResult!: (id: string) => Promise<FileResult | undefined>;

  private _properties: FilePropertyDefinition[] = [
    { key: 'path', label: 'Pfad' },
    { key: 'filename' },
    { key: 'fileSize', label: 'Dateigröße' },
    { key: 'puid' },
    { key: 'mimeType' },
    { key: 'formatVersion' },
    { key: 'status' },
  ];
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
  @Input()
  get properties(): FilePropertyDefinition[] {
    return this._properties;
  }
  set properties(value: FilePropertyDefinition[]) {
    this._properties = value;
    this.tableProperties = value.filter((p) => p.inTable !== false);
    this.columns = this.tableProperties.map(({ key }) => key);
  }

  tableProperties = this.properties.filter((p) => p.inTable !== false);
  columns = this.tableProperties.map(({ key }) => key);

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(
    private dialog: MatDialog,
    private statusIcons: StatusIconsService,
  ) {
    this.dataSource = new MatTableDataSource<FileOverview>([]);
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
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

  private processFileInformation(results: FileResult[]): void {
    const data: FileOverview[] = [];
    for (let result of results) {
      let fileOverview: FileOverview = {
        id: result.id,
        icons: this.statusIcons.getIcons(result),
        values: {
          filename: { value: result.filename },
          puid: { value: result.toolResults.summary['puid']?.values[0].value },
          mimeType: { value: result.toolResults.summary['mimeType']?.values[0].value },
          formatVersion: { value: result.toolResults.summary['formatVersion']?.values[0].value },
          ...result.info,
        },
      };
      data.push(fileOverview);
    }
    this.dataSource.data = data;
  }

  async openDetails(fileOverview: FileOverview): Promise<void> {
    const id = fileOverview.id;
    const fileResult = await this.getResult(id);
    if (fileResult) {
      this.dialog.open(FileOverviewComponent, {
        data: {
          fileResult: fileResult,
          properties: this.properties,
        },
        autoFocus: false,
        width: '1200px',
        maxWidth: '80vw',
      });
    } else {
      console.error('file result not found');
    }
  }
}
