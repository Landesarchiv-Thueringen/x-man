import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatTableModule } from '@angular/material/table';
import { ToolResult } from '../../message/message.service';
import { FileFeaturePipe } from './file-attribut-de.pipe';
import { PrettyPrintCsvPipe } from './pretty-print-csv.pipe';
import { PrettyPrintJsonPipe } from './pretty-print-json.pipe';

interface DialogData {
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
    FileFeaturePipe,
    PrettyPrintCsvPipe,
    PrettyPrintJsonPipe,
  ],
  standalone: true,
})
export class ToolOutputComponent {
  readonly toolResult = this.data.toolResult;

  constructor(@Inject(MAT_DIALOG_DATA) private data: DialogData) {}
}
