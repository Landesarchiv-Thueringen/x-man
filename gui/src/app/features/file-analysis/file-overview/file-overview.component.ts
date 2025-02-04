import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { FilePropertyDefinition } from '../file-analysis-table/file-analysis-table.component';
import { FileFeaturePipe } from '../pipes/file-feature.pipe';
import { FeatureValue, FileAnalysis, RowValue } from '../results';
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
  filename: string;
  info: { [key: string]: RowValue };
  analysis: FileAnalysis;
  properties: FilePropertyDefinition[];
}

interface FileFeature {
  value?: string | boolean | number;
  confidence?: number;
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
  data = inject<DialogData>(MAT_DIALOG_DATA);
  private dialog = inject(MatDialog);

  readonly analysis: FileAnalysis = this.data.analysis;
  dataSource = new MatTableDataSource<FileFeatures>();
  tableColumnList: string[] = [];
  infoProperties = this.data.properties.filter(
    (p) =>
      p.key !== 'filename' &&
      p.key !== 'status' &&
      !OVERVIEW_FEATURES.includes(p.key as (typeof OVERVIEW_FEATURES)[number]),
  );

  constructor() {
    this.initTableData();
  }

  initTableData(): void {
    const features = this.analysis.features;
    if (!features) {
      return;
    }
    let featureNames: OverviewFeature[] = [];
    const toolNames = this.analysis.toolResults.map((r) => r.toolName);
    for (const featureKey in features) {
      if (isOverviewFeature(featureKey)) {
        featureNames.push(featureKey);
      }
    }
    this.dataSource.data = this.getTableRows(toolNames, featureNames, this.analysis.features);
  }

  getTableRows(
    toolNames: string[],
    featureNames: string[],
    featureValues: { [key: string]: FeatureValue[] },
  ): FileFeatures[] {
    const rows: FileFeatures[] = [this.getCumulativeResult(featureNames, featureValues)];
    const sortedFeatures: string[] = sortFeatures([...ALWAYS_VISIBLE_COLUMNS, ...featureNames]);
    this.tableColumnList = ['tool', ...sortedFeatures];
    if (this.analysis.summary.error) {
      this.tableColumnList.push('error');
    }
    for (let toolName of toolNames) {
      const fileFeatures: FileFeatures = {};
      fileFeatures['tool'] = {
        value: toolName,
      };
      for (let featureName of featureNames) {
        for (let featureValue of featureValues[featureName]) {
          if (this.featureOfTool(featureValue, toolName)) {
            const toolConfidence = featureValue.supportingTools[toolName];
            fileFeatures[featureName] = {
              value: featureValue.value,
              confidence: toolConfidence,
            };
          }
        }
      }
      if (this.analysis.toolResults.find((r) => r.toolName === toolName)?.error) {
        fileFeatures['error'] = {
          icon: 'error',
        };
      }
      rows.push(fileFeatures);
    }
    return rows;
  }

  getCumulativeResult(
    featureNames: string[],
    featureValues: { [key: string]: FeatureValue[] },
  ): FileFeatures {
    const features: FileFeatures = {};
    features['tool'] = {
      value: 'Gesamtergebnis',
    };
    for (let featureName of featureNames) {
      // result with highest confidence
      const featureValue = featureValues[featureName][0];
      features[featureName] = {
        value: featureValue.value,
        confidence: featureValue.score,
      };
    }
    return features;
  }

  featureOfTool(featureValue: FeatureValue, toolName: string): boolean {
    return featureValue.supportingTools[toolName] != null;
  }

  showToolOutput(toolName: string): void {
    const toolResult = this.analysis.toolResults.find((r) => r.toolName === toolName);
    if (toolResult) {
      this.dialog.open(ToolOutputComponent, {
        data: {
          toolName,
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
    a.href =
      'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(this.analysis, null, 2));
    a.click();
    document.body.removeChild(a);
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
