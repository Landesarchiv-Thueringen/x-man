import { PercentPipe } from '@angular/common';
import { Component, inject, input, OnInit } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatRippleModule } from '@angular/material/core';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { DialogData, FormatDetailsComponent } from '../format-details/format-details.component';
import { BreakOpportunitiesPipe } from '../pipes/break-opportunities.pipe';
import { FeatureValuePipe } from '../pipes/feature-value.pipe';
import { ToolsPipe } from '../pipes/tools.pipe';
import { FeatureValue, FileAnalysis, ToolResult } from '../results';

interface FormatRow {
  setIndex: number;
  puid: FeatureValue | undefined;
  mimeType: FeatureValue | undefined;
  formatVersion: FeatureValue | undefined;
  valid: FeatureValue | undefined;
  score: number;
  tools: string[];
}

@Component({
  selector: 'app-file-format',
  imports: [
    MatTableModule,
    PercentPipe,
    FeatureValuePipe,
    MatIconModule,
    MatRippleModule,
    BreakOpportunitiesPipe,
    ToolsPipe,
    MatButtonModule,
  ],
  templateUrl: './file-format.component.html',
  styleUrl: './file-format.component.scss',
})
export class FileFormatComponent implements OnInit {
  private readonly dialog = inject(MatDialog);
  readonly fileAnalysis = input.required<FileAnalysis>();
  toolResults: ToolResult[] = [];
  resultUncertain: boolean = false;
  displayedColumns: string[] = ['puid', 'mimeType', 'formatVersion', 'valid', 'tools', 'score'];
  rows: FormatRow[] = [];

  ngOnInit(): void {
    if (this.fileAnalysis().featureSets.length > 0) {
      this.toolResults = this.fileAnalysis().toolResults;
      this.resultUncertain = this.fileAnalysis().summary.formatUncertain;
      this.rows = this.fileAnalysis().featureSets.map((set, index) => {
        return {
          setIndex: index,
          puid: set.features['format:puid'],
          mimeType: set.features['format:mimeType'],
          formatVersion: set.features['format:version'],
          valid: set.features['format:valid'],
          score: set.score,
          tools: set.supportingTools,
        };
      });
    }
  }

  showResultDetails(setIndex: number): void {
    const featureSet = this.fileAnalysis().featureSets[setIndex];
    const toolResults = this.fileAnalysis().toolResults;
    if (featureSet && toolResults) {
      const data: DialogData = {
        analysis: this.fileAnalysis(),
        featureSet: featureSet,
        toolResults: toolResults,
      };
      this.dialog.open(FormatDetailsComponent, {
        data: data,
        autoFocus: false,
        width: '70em',
        maxWidth: '80vw',
      });
    }
  }
}
