import { Component } from '@angular/core';

import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { switchMap, tap } from 'rxjs';
import { MessageService, PrimaryDocumentData } from '../../../../services/message.service';
import { notNull } from '../../../../utils/predicates';
import { MessagePageService } from '../../message-page.service';
import {
  FileAnalysisTableComponent,
  FilePropertyDefinition,
} from './file-analysis/file-analysis-table/file-analysis-table.component';
import { FileResult } from './file-analysis/results';

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
  standalone: true,
  imports: [FileAnalysisTableComponent],
})
export class PrimaryDocumentsTableComponent {
  results?: FileResult[];
  getResult = async (id: string) => this.results?.find((result) => result.id === id);
  properties: FilePropertyDefinition[] = [
    { key: 'filenameComplete', label: 'Dateiname', inTable: false },
    { key: 'recordId', label: 'Dokument', inTable: false },
    { key: 'filename' },
    { key: 'mimeType' },
    { key: 'formatVersion' },
    { key: 'status' },
  ];

  constructor(
    private messageService: MessageService,
    private messagePage: MessagePageService,
  ) {
    let processId: string;
    this.messagePage
      .observeMessage()
      .pipe(
        takeUntilDestroyed(),
        tap((message) => (processId = message.messageHead.processID)),
        switchMap((message) => this.messageService.getPrimaryDocumentsData(message.messageHead.processID)),
      )
      .subscribe((primaryDocuments) => {
        const mapping = primaryDocumentToFileResult.bind(null, processId);
        this.results = primaryDocuments.map(mapping).filter(notNull);
      });
  }
}

function primaryDocumentToFileResult(processId: string, primaryDocument: PrimaryDocumentData): FileResult | undefined {
  if (!primaryDocument.formatVerification) {
    return undefined;
  }
  return {
    id: primaryDocument.filename,
    filename:
      primaryDocument.filenameOriginal ||
      primaryDocument.filename.replace(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}_/, ''),
    info: {
      recordId: {
        value: primaryDocument.recordId,
        routerLink: ['nachricht', processId, '0503', 'dokument', primaryDocument.recordId],
      },
      filenameOriginal: { value: primaryDocument.filenameOriginal },
      filenameComplete: { value: primaryDocument.filename },
    },
    toolResults: primaryDocument.formatVerification,
  };
}
