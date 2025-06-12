import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatTableModule } from '@angular/material/table';
import { MatTabsModule } from '@angular/material/tabs';
import { PrettyPrintCsvPipe } from '../pipes/pretty-print-csv.pipe';
import { PrettyPrintJsonPipe } from '../pipes/pretty-print-json.pipe';
import { PrettyPrintXmlPipe } from '../pipes/pretty-print-xml.pipe';
import { ToolResult } from '../results';

interface DialogData {
  toolName: string;
  toolResult: ToolResult;
}

@Component({
  selector: 'app-tool-output',
  templateUrl: './tool-output.component.html',
  styleUrls: ['./tool-output.component.scss'],
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatExpansionModule,
    MatTableModule,
    PrettyPrintCsvPipe,
    PrettyPrintJsonPipe,
    PrettyPrintXmlPipe,
    MatTabsModule,
  ],
})
export class ToolOutputComponent {
  private data = inject<DialogData>(MAT_DIALOG_DATA);

  readonly toolName = this.data.toolName;
  readonly toolResult = this.data.toolResult;
  readonly showFeatures =
    this.data.toolResult.features && Object.keys(this.data.toolResult.features).length > 0;
}
