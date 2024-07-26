import { Component } from '@angular/core';

import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { firstValueFrom, switchMap, tap } from 'rxjs';
import {
  FileAnalysisTableComponent,
  FilePropertyDefinition,
} from '../../../../features/file-analysis/file-analysis-table/file-analysis-table.component';
import { FileResult } from '../../../../features/file-analysis/results';
import { MessageService, PrimaryDocumentInfo } from '../../../../services/message.service';
import { notNull } from '../../../../utils/predicates';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
  standalone: true,
  imports: [FileAnalysisTableComponent],
})
export class PrimaryDocumentsTableComponent {
  processId!: string;
  results?: FileResult[];
  getDetails = async (id: string) => {
    const data = await firstValueFrom(
      this.messageService.getPrimaryDocumentData(this.processId, id),
    );
    return data.formatVerification;
  };
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
    this.messagePage
      .observeMessage()
      .pipe(
        takeUntilDestroyed(),
        tap((message) => (this.processId = message.messageHead.processID)),
        switchMap((message) =>
          this.messageService.getPrimaryDocumentsInfo(message.messageHead.processID),
        ),
      )
      .subscribe((info) => {
        const mapping = primaryDocumentToFileResult.bind(null, this.processId);
        this.results = info.map(mapping).filter(notNull);
      });
  }
}

function primaryDocumentToFileResult(
  processId: string,
  primaryDocument: PrimaryDocumentInfo,
): FileResult | undefined {
  if (!primaryDocument.formatVerificationSummary) {
    return undefined;
  }
  return {
    id: primaryDocument.filename,
    filename:
      primaryDocument.filenameOriginal ||
      primaryDocument.filename.replace(
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}_/,
        '',
      ),
    info: {
      recordId: {
        value: primaryDocument.recordId,
        routerLink: ['nachricht', processId, '0503', 'dokument', primaryDocument.recordId],
      },
      filenameOriginal: { value: primaryDocument.filenameOriginal },
      filenameComplete: { value: primaryDocument.filename },
    },
    summary: primaryDocument.formatVerificationSummary,
  };
}
