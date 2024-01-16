// angular
import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// material
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator } from '@angular/material/paginator';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Feature, MessageService, PrimaryDocument } from '../../message/message.service';
import { FileOverviewComponent } from '../primary-document/primary-document-metadata.component';

// utility
import { v4 as uuidv4 } from 'uuid';

interface FileOverview {
  [key: string]: FileFeature;
}

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
        this.processFileInformations(primaryDocuments);
      },
    });
  }

  processFileInformations(primaryDocuments: PrimaryDocument[]): void {
    const featureKeys: string[] = ['fileName'];
    const data: FileOverview[] = [];
    for (let primaryDocument of primaryDocuments) {
      if (!primaryDocument.formatVerification) {
        continue;
      }
      const primaryDocumentID: string = uuidv4();
      let fileOverview: FileOverview = {};
      for (let featureKey in primaryDocument.formatVerification.summary) {
        featureKeys.push(featureKey);
        fileOverview['fileName'] = {
          value: primaryDocument.fileName,
        };
        fileOverview[featureKey] = {
          value: primaryDocument.formatVerification.summary[featureKey].values[0].value,
          confidence: primaryDocument.formatVerification.summary[featureKey].values[0].score,
          feature: primaryDocument.formatVerification.summary[featureKey],
        };
        fileOverview['id'] = { value: primaryDocumentID };
      }
      this.primaryDocuments.set(primaryDocumentID, primaryDocument);
      data.push(fileOverview);
    }
    this.dataSource.data = data;
    const features = [...new Set(featureKeys)];
    const selectedFeatures = this.messageService.selectOverviewFeatures(features);
    const sortedFeatures = this.messageService.sortFeatures(selectedFeatures);
    this.generatedTableColumnList = sortedFeatures;
    this.tableColumnList = sortedFeatures.concat(['actions']);
  }

  openDetails(fileOverview: FileOverview): void {
    if (fileOverview) {
      const id: string = fileOverview['id'].value;
      const primaryDocument: PrimaryDocument | undefined = this.primaryDocuments.get(id);
      this.dialog.open(FileOverviewComponent, {
        autoFocus: false,
        data: {
          primaryDocument: primaryDocument,
        },
      });
    }
  }
}
