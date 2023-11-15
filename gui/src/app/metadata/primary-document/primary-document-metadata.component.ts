// angular
import { Component, Inject } from '@angular/core';

// material
import { MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatTableDataSource } from '@angular/material/table';

// project
import {
  FeatureValue,
  MessageService,
  PrimaryDocument,
  Summary,
  ToolConfidence,
  ToolResult,
} from '../../message/message.service';

interface DialogData {
  primaryDocument: PrimaryDocument;
}

interface FileFeature {
  value: string;
  confidence?: number;
  colorizeConfidence?: boolean;
}

interface FileFeatures {
  [key: string]: FileFeature;
}

@Component({
  selector: 'app-file-overview',
  templateUrl: './primary-document-metadata.component.html',
  styleUrls: ['./primary-document-metadata.component.scss'],
})
export class FileOverviewComponent {
  readonly primaryDocument: PrimaryDocument;
  dataSource: MatTableDataSource<FileFeatures>;
  tableColumnList: string[];
  constructor(
    @Inject(MAT_DIALOG_DATA) private data: DialogData,
    private messageService: MessageService
  ) {
    this.dataSource = new MatTableDataSource<FileFeatures>();
    this.tableColumnList = ['Attribut'];
    console.log(this.data);
    this.primaryDocument = this.data.primaryDocument;
    this.initTableData();
  }

  initTableData(): void {
    if (this.primaryDocument?.formatVerification?.summary) {
      const summary = this.primaryDocument.formatVerification.summary;
      const toolNames: string[] = [];
      let featureNames: string[] = [];
      let toolResults: ToolResult[] =
        this.primaryDocument.formatVerification.fileIdentificationResults;
      if (this.primaryDocument.formatVerification.fileValidationResults) {
        toolResults = toolResults.concat(
          this.primaryDocument.formatVerification.fileValidationResults
        );
      }
      toolResults.forEach((toolResult: ToolResult) => {
        toolNames.push(toolResult.toolName);
      });
      for (let featureKey in summary) {
        featureNames.push(featureKey);
      }
      featureNames = this.messageService.selectOverviewFeatures(featureNames);
      this.dataSource.data = this.getTableRows(
        summary,
        toolNames,
        featureNames
      );
    }
  }

  getTableRows(
    summary: Summary,
    toolNames: string[],
    featureNames: string[]
  ): FileFeatures[] {
    const rows: FileFeatures[] = [
      this.getCumulativeResult(summary, featureNames),
    ];
    const sortedFeatures: string[] =
      this.messageService.sortFeatures(featureNames);
    this.tableColumnList = ['Werkzeug', ...sortedFeatures];
    for (let toolName of toolNames) {
      const featureValues: FileFeatures = {};
      featureValues['Werkzeug'] = {
        value: toolName,
      };
      for (let featureName of featureNames) {
        for (let featureValue of summary[featureName].values) {
          if (this.featureOfTool(featureValue, toolName)) {
            const toolInfo: ToolConfidence = featureValue.tools.find(
              (toolInfo: ToolConfidence) => {
                return toolInfo.toolName === toolName;
              }
            )!;
            featureValues[featureName] = {
              value: featureValue.value,
              confidence: toolInfo.confidence,
              colorizeConfidence: false,
            };
          }
        }
      }
      rows.push(featureValues);
    }
    return rows;
  }

  getCumulativeResult(summary: Summary, featureNames: string[]): FileFeatures {
    const features: FileFeatures = {};
    features['Werkzeug'] = {
      value: 'Gesamtergebnis',
    };
    for (let featureName of featureNames) {
      // result with highest confidence
      const featureValue = summary[featureName].values[0];
      features[featureName] = {
        value: featureValue.value,
        confidence: featureValue.score,
        colorizeConfidence: true,
      };
    }
    return features;
  }

  featureOfTool(featureValue: FeatureValue, toolName: string): boolean {
    for (let tool of featureValue.tools) {
      if (tool.toolName === toolName) {
        return true;
      }
    }
    return false;
  }
}
