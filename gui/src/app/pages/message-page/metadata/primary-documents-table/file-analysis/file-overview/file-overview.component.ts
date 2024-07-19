import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { FilePropertyDefinition } from '../file-analysis-table/file-analysis-table.component';
import { FileFeaturePipe } from '../pipes/file-feature.pipe';
import { FeatureValue, FileResult, Summary, ToolConfidence, ToolResult } from '../results';
import { StatusIconsService } from '../status-icons.service';
import { ToolOutputComponent } from '../tool-output/tool-output.component';

const OVERVIEW_FEATURES = ['puid', 'mimeType', 'formatVersion', 'valid'] as const;
type OverviewFeature = (typeof OVERVIEW_FEATURES)[number];

const featureOrder = new Map<string, number>([
  ['puid', 4],
  ['mimeType', 5],
  ['formatVersion', 6],
  ['encoding', 7],
  ['', 101],
  ['wellFormed', 1001],
  ['valid', 1002],
]);

interface DialogData {
  fileResult: FileResult;
  properties: FilePropertyDefinition[];
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
  templateUrl: './file-overview.component.html',
  styleUrls: ['./file-overview.component.scss'],
  standalone: true,
  imports: [
    CommonModule,
    FileFeaturePipe,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatTableModule,
    RouterModule,
  ],
})
export class FileOverviewComponent {
  readonly fileResult: FileResult = this.data.fileResult;
  readonly icons = this.statusIcons.getIcons(this.data.fileResult);
  dataSource = new MatTableDataSource<FileFeatures>();
  tableColumnList: string[] = [];
  infoProperties = this.data.properties.filter(
    (p) =>
      p.key !== 'filename' &&
      p.key !== 'status' &&
      !OVERVIEW_FEATURES.includes(p.key as (typeof OVERVIEW_FEATURES)[number]),
  );

  constructor(
    @Inject(MAT_DIALOG_DATA) public data: DialogData,
    private statusIcons: StatusIconsService,
    private dialog: MatDialog,
  ) {
    console.log('infoProperties', this.infoProperties);
    this.initTableData();
  }

  initTableData(): void {
    if (this.fileResult.toolResults.summary) {
      const summary = this.fileResult.toolResults.summary;
      const toolNames: string[] = [];
      let featureNames: OverviewFeature[] = [];
      let toolResults: ToolResult[] = this.fileResult.toolResults.fileIdentificationResults ?? [];
      if (this.fileResult.toolResults.fileValidationResults) {
        toolResults = toolResults.concat(this.fileResult.toolResults.fileValidationResults);
      }
      toolResults.forEach((toolResult: ToolResult) => {
        toolNames.push(toolResult.toolName);
      });
      for (let featureKey in summary) {
        if (isOverviewFeature(featureKey)) {
          featureNames.push(featureKey);
        }
      }
      this.dataSource.data = this.getTableRows(summary, toolNames, featureNames);
    }
  }

  getTableRows(summary: Summary, toolNames: string[], featureNames: string[]): FileFeatures[] {
    const rows: FileFeatures[] = [this.getCumulativeResult(summary, featureNames)];
    const sortedFeatures: string[] = sortFeatures([...ALWAYS_VISIBLE_COLUMNS, ...featureNames]);
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
        maxWidth: '80vw',
      });
    }
  }

  exportResult(): void {
    const a = document.createElement('a');
    document.body.appendChild(a);
    a.download = 'borg-results.json';
    a.href = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(this.fileResult, null, 2));
    a.click();
    document.body.removeChild(a);
  }

  private findToolResult(toolName: string): ToolResult | undefined {
    return [
      ...(this.fileResult.toolResults.fileIdentificationResults ?? []),
      ...(this.fileResult.toolResults.fileValidationResults ?? []),
    ].find((result) => result.toolName === toolName);
  }
}

function isOverviewFeature(feature: string): feature is OverviewFeature {
  return (OVERVIEW_FEATURES as readonly string[]).includes(feature);
}

/** Sorts feature keys and removes duplicates. */
function sortFeatures(features: string[]): string[] {
  features = [...new Set(features)];
  return features.sort((f1: string, f2: string) => {
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
