import { DatePipe } from '@angular/common';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map, shareReplay } from 'rxjs';
import { FileAnalysis, Summary } from '../features/file-analysis/results';

export interface Message {
  messageType: MessageType;
  creationTime: string;
  xdomeaVersion: string;
  messageHead: MessageHead;
}

export type MessageType = '0501' | '0502' | '0503' | '0504' | '0505' | '0506' | '0507';

export interface MessageHead {
  processID: string;
  creationTime: string;
  sender: Contact;
  receiver: Contact;
}

export interface Contact {
  agencyIdentification?: AgencyIdentification;
  institution?: Institution;
}

export interface AgencyIdentification {
  code: string;
  prefix: string;
}

export interface Institution {
  name: string;
  abbreviation: string;
}

export type FormatVerification = FileAnalysis;

export interface PrimaryDocumentInfo {
  recordId: string;
  filename: string;
  filenameOriginal: string;
  formatVerificationSummary?: Summary;
}

export interface PrimaryDocumentData {
  recordId: string;
  filename: string;
  filenameOriginal: string;
  creatorName: string;
  creationTime: string;
  formatVerification?: FormatVerification;
}

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  private datePipe = inject(DatePipe);
  private httpClient = inject(HttpClient);

  private cachedMessageId?: {
    processId: string;
    messageType: MessageType;
  };
  private cachedMessage?: Observable<Message>;

  getMessage(processId: string, messageType: MessageType): Observable<Message> {
    if (!processId || !messageType) {
      throw new Error('called getMessage with empty string');
    }
    if (
      this.cachedMessageId?.processId !== processId ||
      this.cachedMessageId?.messageType !== messageType
    ) {
      this.cachedMessageId = { processId, messageType };
      this.cachedMessage = this.httpClient
        .get<Message>('/api/message/' + processId + '/' + messageType)
        .pipe(shareReplay({ bufferSize: 1, refCount: true }));
    }
    return this.cachedMessage!;
  }

  getPrimaryDocument(processId: string, filename: string): Observable<Blob> {
    const url = '/api/primary-document';
    const options = {
      params: new HttpParams().set('processID', processId).set('filename', filename),
      responseType: 'blob' as 'json', // https://github.com/angular/angular/issues/18586
    };
    return this.httpClient.get<Blob>(url, options);
  }

  getPrimaryDocumentsInfo(processId: string): Observable<PrimaryDocumentInfo[]> {
    if (!processId) {
      throw new Error('called getPrimaryDocuments with null ID');
    }
    const url = '/api/primary-documents-info/' + processId;
    return this.httpClient.get<PrimaryDocumentInfo[]>(url);
  }

  getPrimaryDocumentData(processId: string, filename: string): Observable<PrimaryDocumentData> {
    if (!processId) {
      throw new Error('called getPrimaryDocuments with null ID');
    }
    const url = '/api/primary-document-data/' + processId + '/' + encodeURIComponent(filename);
    return this.httpClient.get<PrimaryDocumentData>(url);
  }

  finalizeMessageAppraisal(messageId: string): Observable<void> {
    const url = '/api/finalize-message-appraisal/' + messageId;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  archive0503Message(processId: string, collectionId: string): Observable<void> {
    let url = '/api/archive-0503-message/' + processId;
    if (collectionId) {
      url += '?collectionId=' + collectionId;
    }
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  areAllRecordObjectsAppraised(processId: string): Observable<boolean> {
    return this.httpClient.get<boolean>('/api/all-record-objects-appraised/' + processId);
  }

  /**
   * Returns null if the xml node or its text contents are null, because that means the date was not
   * provided in the message. Returns the text content of the xml node if the text content is no
   * parsable date to show the malformed date in the ui. Returns formatted date string if text
   * content is parsable.
   */
  getDateText(dateText: string | null | undefined): string | null {
    if (dateText) {
      const timestamp: number = Date.parse(dateText);
      if (Number.isNaN(timestamp)) {
        return dateText;
      } else {
        const date: Date = new Date(timestamp);
        return this.datePipe.transform(date);
      }
    }
    return null;
  }

  reimportMessage(processId: string, type: MessageType): Observable<void> {
    return this.httpClient
      .post('/api/message/' + processId + '/' + type + '/reimport', null)
      .pipe(map(() => void 0));
  }

  deleteMessage(processId: string, type: MessageType): Observable<void> {
    return this.httpClient.delete('/api/message/' + processId + '/' + type).pipe(map(() => void 0));
  }
}
