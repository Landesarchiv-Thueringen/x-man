import { Component, inject } from '@angular/core';

import { firstValueFrom } from 'rxjs';
import { FileResult } from '../../../../features/file-analysis/results';
import {
  FilePropertyDefinition,
  ResultsComponent,
} from '../../../../features/file-analysis/results/results.component';
import { MessageService, PrimaryDocumentInfo } from '../../../../services/message.service';
import { notNull } from '../../../../utils/predicates';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
  imports: [ResultsComponent],
})
export class PrimaryDocumentsTableComponent {
  private messageService = inject(MessageService);
  private messagePage = inject(MessagePageService);

  processId = this.messagePage.processId;
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
    { key: 'filename', label: 'Dateiname' },
    { key: 'mimeType', label: 'MIME-Type' },
    { key: 'formatVersion', label: 'Formatversion' },
    { key: 'status', label: 'Status' },
  ];

  constructor() {
    this.messageService.getPrimaryDocumentsInfo(this.processId).subscribe((info) => {
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
