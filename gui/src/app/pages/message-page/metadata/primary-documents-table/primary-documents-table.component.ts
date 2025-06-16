import { Component, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { FileResultsComponent } from '../../../../features/file-analysis/file-results/file-results.component';
import { FeatureValue, FileResult } from '../../../../features/file-analysis/results';
import { MessageService, PrimaryDocumentInfo } from '../../../../services/message.service';
import { notNull } from '../../../../utils/predicates';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
  imports: [FileResultsComponent],
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
  const additionalMetadata: {
    [key: string]: FeatureValue | undefined;
  } = {};
  if (primaryDocument.filenameOriginal) {
    additionalMetadata['general:filenameOriginal'] = {
      value: primaryDocument.filenameOriginal,
      label: 'ursprünglicher Dateiname',
      supportingTools: ['x-man'],
    };
  }
  if (primaryDocument.filename) {
    additionalMetadata['general:filenameComplete'] = {
      value: primaryDocument.filename,
      label: 'vollständiger Dateiname',
      supportingTools: ['x-man'],
    };
  }
  return {
    id: primaryDocument.filename,
    filename:
      primaryDocument.filenameOriginal ||
      primaryDocument.filename.replace(
        /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}_/,
        '',
      ),
    resourceLink: {
      sectionLabel: 'Dokument',
      linkLabel: primaryDocument.recordId,
      routerLink: ['nachricht', processId, '0503', 'dokument', primaryDocument.recordId],
    },
    additionalMetadata: additionalMetadata,
    summary: primaryDocument.formatVerificationSummary,
  };
}
