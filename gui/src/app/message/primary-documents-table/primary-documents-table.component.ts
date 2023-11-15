// angular
import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// material
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator } from '@angular/material/paginator';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Feature, MessageService, PrimaryDocument } from '../message.service';

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
  featureOrder: Map<string, number>;
  overviewFeatures: string[];

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  
  constructor(
    private messageService: MessageService,
    private route: ActivatedRoute
  ) {
    this.dataSource = new MatTableDataSource<FileOverview>([]);
    this.tableColumnList = [];
    this.generatedTableColumnList = ['fileName'];
    this.overviewFeatures = [
      'relativePath',
      'fileName',
      'fileSize',
      'puid',
      'mimeType',
      'formatVersion',
      'valid',
    ];
    this.featureOrder = new Map<string, number>([
      ['relativePath', 1],
      ['fileName', 2],
      ['fileSize', 3],
      ['puid', 4],
      ['mimeType', 5],
      ['formatVersion', 6],
      ['encoding', 7],
      ['', 101],
      ['wellFormed', 1001],
      ['valid', 1002],
    ]);
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
      let fileOverview: FileOverview = {};
      if (primaryDocument.formatVerification) {
        for (let featureKey in primaryDocument.formatVerification.summary) {
          featureKeys.push(featureKey);
          fileOverview['fileName'] = { value: primaryDocument.fileName };
          fileOverview[featureKey] = {
            value: primaryDocument.formatVerification.summary[featureKey].values[0].value,
            confidence: primaryDocument.formatVerification.summary[featureKey].values[0].score,
            feature: primaryDocument.formatVerification.summary[featureKey],
          };
        } 
      }
      data.push(fileOverview);
    }
    this.dataSource.data = data;
    const features = [...new Set(featureKeys)];
    const selectedFeatures = this.selectOverviewFeatures(features);
    const sortedFeatures = this.sortFeatures(selectedFeatures);
    this.generatedTableColumnList = sortedFeatures;
    this.tableColumnList = sortedFeatures.concat(['actions']);
  }

  sortFeatures(features: string[]): string[] {
    return features.sort((f1: string, f2: string) => {
      const featureOrder = this.featureOrder;
      let orderF1: number | undefined = featureOrder.get(f1);
      if (!orderF1) {
        orderF1 = featureOrder.get('');
      }
      let orderF2: number | undefined = featureOrder.get(f2);
      if (!orderF2) {
        orderF2 = featureOrder.get('');
      }
      if (orderF1! < orderF2!) {
        return -1;
      } else if (orderF1! > orderF2!) {
        return 1;
      }
      return 0;
    });
  }

  selectOverviewFeatures(features: string[]): string[] {
    const overviewFeatures: string[] = this.overviewFeatures;
    return features.filter((feature: string) => {
      return overviewFeatures.includes(feature);
    });
  }

  openDetails(fileOverview: FileOverview): void {
    // const id = fileOverview['id']?.value;
    // const fileResult =  this.fileAnalysisService.getFileResult(id);
    // if (fileResult) {
    //   this.dialog.open(FileOverviewComponent, {
    //     data: {
    //       fileResult: fileResult
    //     }
    //   });
    // } else {
    //   console.error('file result not found');
    // }
  }
}
