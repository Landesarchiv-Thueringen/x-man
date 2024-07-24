import { DatePipe } from '@angular/common';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, map, shareReplay } from 'rxjs';
import { environment } from '../../environments/environment';
import {
  FileAnalysis,
  Summary,
} from '../pages/message-page/metadata/primary-documents-table/file-analysis/results';

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

export interface ToolResult {
  toolName: string;
  toolVersion: string;
  toolOutput: string;
  outputFormat: 'text' | 'json' | 'csv';
  extractedFeatures: { [key: string]: string };
  error: string;
}

export interface Feature {
  key: string;
  values: FeatureValue[];
}

export interface FeatureValue {
  value: string;
  score: number;
  tools: ToolConfidence[];
}

export interface ToolConfidence {
  confidence: number;
  toolName: string;
}

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
  private apiEndpoint: string;

  private featureOrder: Map<string, number>;
  private overviewFeatures: string[];

  private cachedMessageId?: {
    processId: string;
    messageType: MessageType;
  };
  private cachedMessage?: Observable<Message>;

  constructor(
    private datePipe: DatePipe,
    private httpClient: HttpClient,
  ) {
    this.apiEndpoint = environment.endpoint;
    this.overviewFeatures = [
      'relativePath',
      'fileName',
      'fileSize',
      'puid',
      'mimeType',
      'formatVersion',
      'valid',
    ];
    this.featureOrder = new Map<string, number>([
      ['relativePath', 1],
      ['fileName', 2],
      ['fileSize', 3],
      ['puid', 4],
      ['mimeType', 5],
      ['formatVersion', 6],
      ['encoding', 7],
      ['', 101],
      ['wellFormed', 1001],
      ['valid', 1002],
    ]);
  }

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
        .get<Message>(this.apiEndpoint + '/message/' + processId + '/' + messageType)
        .pipe(shareReplay({ bufferSize: 1, refCount: true }));
    }
    return this.cachedMessage!;
  }

  getPrimaryDocument(processId: string, filename: string): Observable<Blob> {
    const url = this.apiEndpoint + '/primary-document';
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
    const url = this.apiEndpoint + '/primary-documents-info/' + processId;
    return this.httpClient.get<PrimaryDocumentInfo[]>(url);
  }

  getPrimaryDocumentData(processId: string, filename: string): Observable<PrimaryDocumentData> {
    if (!processId) {
      throw new Error('called getPrimaryDocuments with null ID');
    }
    const url =
      this.apiEndpoint + '/primary-document-data/' + processId + '/' + encodeURIComponent(filename);
    return this.httpClient.get<PrimaryDocumentData>(url);
  }

  finalizeMessageAppraisal(messageId: string): Observable<void> {
    const url = this.apiEndpoint + '/finalize-message-appraisal/' + messageId;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  archive0503Message(processId: string, collectionId: string): Observable<void> {
    let url = this.apiEndpoint + '/archive-0503-message/' + processId;
    if (collectionId) {
      url += '?collectionId=' + collectionId;
    }
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  areAllRecordObjectsAppraised(processId: string): Observable<boolean> {
    return this.httpClient.get<boolean>(
      this.apiEndpoint + '/all-record-objects-appraised/' + processId,
    );
  }

  sortFeatures(features: string[]): string[] {
    features = [...new Set(features)];
    return features.sort((f1: string, f2: string) => {
      const featureOrder = this.featureOrder;
      let orderF1: number | undefined = featureOrder.get(f1);
      if (!orderF1) {
        orderF1 = featureOrder.get('');
      }
      let orderF2: number | undefined = featureOrder.get(f2);
      if (!orderF2) {
        orderF2 = featureOrder.get('');
      }
      if (orderF1! < orderF2!) {
        return -1;
      } else if (orderF1! > orderF2!) {
        return 1;
      }
      return 0;
    });
  }

  selectOverviewFeatures(features: string[]): string[] {
    const overviewFeatures: string[] = this.overviewFeatures;
    return features.filter((feature: string) => {
      return overviewFeatures.includes(feature);
    });
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
      .post(this.apiEndpoint + '/message/' + processId + '/' + type + '/reimport', null)
      .pipe(map(() => void 0));
  }

  deleteMessage(processId: string, type: MessageType): Observable<void> {
    return this.httpClient
      .delete(this.apiEndpoint + '/message/' + processId + '/' + type)
      .pipe(map(() => void 0));
  }
}
