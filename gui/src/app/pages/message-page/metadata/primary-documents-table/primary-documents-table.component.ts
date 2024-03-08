import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator, MatPaginatorModule } from '@angular/material/paginator';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { v4 as uuidv4 } from 'uuid';

import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { switchMap } from 'rxjs';
import { Feature, MessageService, PrimaryDocument } from '../../../../services/message.service';
import { MessagePageService } from '../../message-page.service';
import { FileOverviewComponent } from '../primary-document/primary-document-metadata.component';
import { FileFeaturePipe } from '../tool-output/file-attribut-de.pipe';
import { BreakOpportunitiesPipe } from './break-opportunities.pipe';
import { StatusIcons, StatusIconsService } from './status-icons.service';

const OVERVIEW_FEATURES = [
  'relativePath',
  'fileName',
  'fileSize',
  'puid',
  'mimeType',
  'formatVersion',
  'valid',
] as const;
export type OverviewFeature = (typeof OVERVIEW_FEATURES)[number];

type FileOverview = {
  [key in OverviewFeature]?: FileFeature;
} & {
  id: FileFeature;
  icons: StatusIcons;
};

interface FileFeature {
  value: string;
  confidence?: number;
  feature?: Feature;
  tooltip?: string;
}

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
  standalone: true,
  imports: [
    BreakOpportunitiesPipe,
    CommonModule,
    FileFeaturePipe,
    MatButtonModule,
    MatIconModule,
    MatPaginatorModule,
    MatTableModule,
    MatToolbarModule,
  ],
})
export class PrimaryDocumentsTableComponent implements AfterViewInit {
  dataSource: MatTableDataSource<FileOverview>;
  generatedTableColumnList: string[];
  tableColumnList: string[];
  primaryDocuments: Map<string, PrimaryDocument>;

  @ViewChild(MatPaginator) paginator!: MatPaginator;

  constructor(
    private dialog: MatDialog,
    private messageService: MessageService,
    private statusIcons: StatusIconsService,
    private messagePage: MessagePageService,
  ) {
    this.primaryDocuments = new Map<string, PrimaryDocument>();
    this.dataSource = new MatTableDataSource<FileOverview>([]);
    this.tableColumnList = [];
    this.generatedTableColumnList = ['fileName'];
    this.messagePage
      .observeMessage()
      .pipe(
        takeUntilDestroyed(),
        switchMap((message) => this.messageService.getPrimaryDocuments(message.id)),
      )
      .subscribe((primaryDocuments) => this.processFileInformation(primaryDocuments));
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
  }

  processFileInformation(primaryDocuments: PrimaryDocument[]): void {
    const featureKeys: string[] = ['fileName'];
    const data: FileOverview[] = [];
    for (let primaryDocument of primaryDocuments) {
      if (!primaryDocument.formatVerification) {
        continue;
      }
      const primaryDocumentID: string = uuidv4();

      let fileOverview: FileOverview = {
        id: { value: primaryDocumentID },
        icons: this.statusIcons.getIcons(primaryDocument),
      };
      fileOverview['fileName'] = { value: primaryDocument.fileName };
      for (let featureKey in primaryDocument.formatVerification.summary) {
        if (isOverviewFeature(featureKey) && featureKey !== 'valid') {
          featureKeys.push(featureKey);
          fileOverview[featureKey] = {
            value: primaryDocument.formatVerification.summary[featureKey].values[0].value,
            confidence: primaryDocument.formatVerification.summary[featureKey].values[0].score,
            feature: primaryDocument.formatVerification.summary[featureKey],
          };
        }
      }
      this.primaryDocuments.set(primaryDocumentID, primaryDocument);
      data.push(fileOverview);
    }
    this.dataSource.data = data;
    const sortedFeatures = this.messageService.sortFeatures(featureKeys);
    this.generatedTableColumnList = sortedFeatures;
    this.tableColumnList = sortedFeatures.concat(['status']);
  }

  openDetails(fileOverview: FileOverview): void {
    if (fileOverview) {
      const id: string = fileOverview['id'].value;
      const primaryDocument = this.primaryDocuments.get(id);
      this.dialog.open(FileOverviewComponent, {
        autoFocus: false,
        data: {
          primaryDocument: primaryDocument,
        },
      });
    }
  }
}

function isOverviewFeature(feature: string): feature is OverviewFeature {
  return (OVERVIEW_FEATURES as readonly string[]).includes(feature);
}
