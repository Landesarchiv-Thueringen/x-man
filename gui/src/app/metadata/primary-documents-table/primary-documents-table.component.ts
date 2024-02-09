import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator } from '@angular/material/paginator';
import { MatTableDataSource } from '@angular/material/table';
import { ActivatedRoute } from '@angular/router';
import { v4 as uuidv4 } from 'uuid';

import { Feature, MessageService, PrimaryDocument } from '../../message/message.service';
import { FileOverviewComponent } from '../primary-document/primary-document-metadata.component';
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
    private route: ActivatedRoute,
    private statusIcons: StatusIconsService,
  ) {
    this.primaryDocuments = new Map<string, PrimaryDocument>();
    this.dataSource = new MatTableDataSource<FileOverview>([]);
    this.tableColumnList = [];
    this.generatedTableColumnList = ['fileName'];
  }

  ngAfterViewInit(): void {
    this.dataSource.paginator = this.paginator;
    const messageID: string = this.route.parent!.snapshot.params['id'];
    this.messageService.getPrimaryDocuments(messageID).subscribe({
      error: (error: any) => {
        console.error(error);
      },
      next: (primaryDocuments: PrimaryDocument[]) => {
        this.processFileInformation(primaryDocuments);
      },
    });
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
