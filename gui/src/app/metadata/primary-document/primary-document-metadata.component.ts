import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { MatTableDataSource } from '@angular/material/table';
import {
  FeatureValue,
  MessageService,
  PrimaryDocument,
  Summary,
  ToolConfidence,
  ToolResult,
} from '../../message/message.service';
import { StatusIconsService } from '../primary-documents-table/status-icons.service';
import { ToolOutputComponent } from '../tool-output/tool-output.component';
interface DialogData {
  primaryDocument: PrimaryDocument;
}

interface FileFeature {
  value?: string;
  confidence?: number;
  colorizeConfidence?: boolean;
  icon?: string;
}

interface FileFeatures {
  [key: string]: FileFeature;
}

const ALWAYS_VISIBLE_COLUMNS = ['puid', 'mimeType'];

@Component({
  selector: 'app-file-overview',
  templateUrl: './primary-document-metadata.component.html',
  styleUrls: ['./primary-document-metadata.component.scss'],
})
export class FileOverviewComponent {
  readonly primaryDocument = this.data.primaryDocument;
  readonly icons = this.statusIcons.getIcons(this.data.primaryDocument);
  dataSource = new MatTableDataSource<FileFeatures>();
  tableColumnList: string[] = [];

  constructor(
    @Inject(MAT_DIALOG_DATA) private data: DialogData,
    private statusIcons: StatusIconsService,
    private dialog: MatDialog,
    private messageService: MessageService,
  ) {
    this.initTableData();
  }

  initTableData(): void {
    if (this.primaryDocument?.formatVerification?.summary) {
      const summary = this.primaryDocument.formatVerification.summary;
      const toolNames: string[] = [];
      let featureNames: string[] = [];
      let toolResults: ToolResult[] = this.primaryDocument.formatVerification.fileIdentificationResults;
      if (this.primaryDocument.formatVerification.fileValidationResults) {
        toolResults = toolResults.concat(this.primaryDocument.formatVerification.fileValidationResults);
      }
      toolResults.forEach((toolResult: ToolResult) => {
        toolNames.push(toolResult.toolName);
      });
      for (let featureKey in summary) {
        featureNames.push(featureKey);
      }
      featureNames = this.messageService.selectOverviewFeatures(featureNames);
      this.dataSource.data = this.getTableRows(summary, toolNames, featureNames);
    }
  }

  getTableRows(summary: Summary, toolNames: string[], featureNames: string[]): FileFeatures[] {
    const rows: FileFeatures[] = [this.getCumulativeResult(summary, featureNames)];
    const sortedFeatures: string[] = this.messageService.sortFeatures([...ALWAYS_VISIBLE_COLUMNS, ...featureNames]);
    this.tableColumnList = ['tool', ...sortedFeatures];
    if (this.icons.error) {
      this.tableColumnList.push('error');
    }
    for (let toolName of toolNames) {
      const featureValues: FileFeatures = {};
      featureValues['tool'] = {
        value: toolName,
      };
      for (let featureName of featureNames) {
        for (let featureValue of summary[featureName].values) {
          if (this.featureOfTool(featureValue, toolName)) {
            const toolInfo: ToolConfidence = featureValue.tools.find((toolInfo: ToolConfidence) => {
              return toolInfo.toolName === toolName;
            })!;
            featureValues[featureName] = {
              value: featureValue.value,
              confidence: toolInfo.confidence,
              colorizeConfidence: false,
            };
          }
        }
      }
      if (this.findToolResult(toolName)?.error) {
        featureValues['error'] = {
          icon: 'error',
        };
      }
      rows.push(featureValues);
    }
    return rows;
  }

  getCumulativeResult(summary: Summary, featureNames: string[]): FileFeatures {
    const features: FileFeatures = {};
    features['tool'] = {
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

  showToolOutput(toolName: string): void {
    const toolResult = this.findToolResult(toolName);
    if (toolResult) {
      this.dialog.open(ToolOutputComponent, {
        data: {
          toolResult,
        },
        autoFocus: false,
      });
    }
  }

  private findToolResult(toolName: string): ToolResult | undefined {
    return [
      ...(this.data.primaryDocument.formatVerification?.fileIdentificationResults ?? []),
      ...(this.data.primaryDocument.formatVerification?.fileValidationResults ?? []),
    ].find((result) => result.toolName === toolName);
  }
}
