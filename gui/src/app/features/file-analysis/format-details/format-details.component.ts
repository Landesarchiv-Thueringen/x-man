import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatRippleModule } from '@angular/material/core';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { RouterModule } from '@angular/router';
import { BreakOpportunitiesPipe } from '../pipes/break-opportunities.pipe';
import { FeatureValuePipe } from '../pipes/feature-value.pipe';
import { FeatureSet, FileAnalysis, ToolFeatureValue, ToolResult } from '../results';
import { ToolDetailsComponent } from '../tool-details/tool-details.component';

export interface DialogData {
  featureSet: FeatureSet;
  toolResults: ToolResult[];
  analysis: FileAnalysis;
}

interface ToolRow {
  toolName: string;
  puid: ToolFeatureValue | undefined;
  mimeType: ToolFeatureValue | undefined;
  formatVersion: ToolFeatureValue | undefined;
  valid: ToolFeatureValue | undefined;
  error: boolean;
}

@Component({
  selector: 'app-format-details',
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatTableModule,
    RouterModule,
    MatTabsModule,
    FeatureValuePipe,
    MatRippleModule,
    BreakOpportunitiesPipe,
  ],
  templateUrl: './format-details.component.html',
  styleUrl: './format-details.component.scss',
})
export class FormatDetailsComponent implements OnInit {
  private readonly data = inject<DialogData>(MAT_DIALOG_DATA);
  private readonly dialog = inject(MatDialog);
  private readonly featureSet: FeatureSet = this.data.featureSet;
  private readonly toolResults: ToolResult[] = this.data.toolResults;
  readonly rows: ToolRow[] = [];
  displayedColumns: string[] = ['tool', 'puid', 'mimeType', 'formatVersion', 'valid'];

  ngOnInit() {
    this.rows.push({
      toolName: 'Gesamtergebnis',
      puid: this.featureSet.features['format:puid'],
      mimeType: this.featureSet.features['format:mimeType'],
      formatVersion: this.featureSet.features['format:version'],
      valid: this.featureSet.features['format:valid'],
      error: false,
    });
    // gather all tool results which support the feature set
    for (let toolResult of this.toolResults.filter((tr) =>
      this.featureSet.supportingTools.includes(tr.id),
    )) {
      this.rows.push(this.getToolRow(toolResult));
    }
    // gather all tool results with errors
    for (let toolResult of this.toolResults.filter((tr) => tr.error)) {
      this.rows.push(this.getToolRow(toolResult));
    }
    if (this.rows.some((row) => row.error)) {
      this.displayedColumns.push('error');
    }
  }

  getToolRow(tr: ToolResult): ToolRow {
    return {
      toolName: tr.title,
      puid: tr.features['format:puid'],
      mimeType: tr.features['format:mimeType'],
      formatVersion: tr.features['format:version'],
      valid: tr.features['format:valid'],
      error: !!tr.error,
    };
  }

  showToolOutput(toolName: string): void {
    const toolResult = this.toolResults.find((r) => r.title === toolName);
    if (toolResult) {
      this.dialog.open(ToolDetailsComponent, {
        data: {
          toolName,
          toolResult,
        },
        autoFocus: false,
        height: '40em',
        width: '70em',
        maxWidth: '80vw',
      });
    }
  }
}
