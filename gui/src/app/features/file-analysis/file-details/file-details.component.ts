import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatTabsModule } from '@angular/material/tabs';
import { FileFormatComponent } from '../file-format/file-format.component';
import { MetadataComponent } from '../metadata/metadata-table.component';
import { FileAnalysis, RowValue } from '../results';

interface DialogData {
  filename: string;
  info: { [key: string]: RowValue };
  analysis: FileAnalysis;
}

@Component({
  selector: 'app-file-details',
  templateUrl: './file-details.component.html',
  styleUrls: ['./file-details.component.scss'],
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatTabsModule,
    MetadataComponent,
    FileFormatComponent,
  ],
})
export class FileDetailsComponent {
  data = inject<DialogData>(MAT_DIALOG_DATA);
  readonly analysis: FileAnalysis = this.data.analysis;

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
