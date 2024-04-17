import { DatePipe } from '@angular/common';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, Subscriber, filter, shareReplay } from 'rxjs';
import { environment } from '../../environments/environment';
import { notNull } from '../utils/predicates';

export interface Message {
  id: string;
  messageType: MessageType;
  creationTime: string;
  xdomeaVersion: string;
  schemaValidation: boolean;
  messageHead: MessageHead;
  formatVerificationComplete: boolean;
  primaryDocumentCount: number;
  verificationCompleteCount: number;
  fileRecordObjects?: FileRecordObject[];
  processRecordObjects?: ProcessRecordObject[];
  documentRecordObjects?: DocumentRecordObject[];
}

export interface MessageType {
  id: number;
  code: string;
}

export interface MessageHead {
  id: number;
  processID: string;
  creationTime: string;
  sender: Contact;
  receiver: Contact;
}

export interface Contact {
  id: number;
  agencyIdentification?: AgencyIdentification;
  institution?: Institution;
}

export interface AgencyIdentification {
  id: number;
  code?: string;
  prefix?: string;
}

export interface Institution {
  id: number;
  name?: string;
  abbreviation?: string;
}

export interface FileRecordObject {
  id: string;
  xdomeaID: string;
  messageID: string;
  recordObjectType: RecordObjectType;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  subfiles: FileRecordObject[];
  processes: ProcessRecordObject[];
}

export interface ProcessRecordObject {
  id: string;
  xdomeaID: string;
  messageID: string;
  recordObjectType: RecordObjectType;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  subprocesses: ProcessRecordObject[];
  documents: DocumentRecordObject[];
}

export interface DocumentRecordObject {
  id: string;
  xdomeaID: string;
  messageID: string;
  recordObjectType: RecordObjectType;
  generalMetadata?: GeneralMetadata;
  type?: string;
  incomingDate?: string;
  outgoingDate?: string;
  documentDate?: string;
  versions?: DocumentVersion[];
  attachments?: DocumentRecordObject[];
}

export interface DocumentVersion {
  id: number;
  versionID: string;
  formats: Format[];
}

export interface Format {
  id: number;
  code: string;
  otherName?: string;
  version: string;
  primaryDocument: PrimaryDocument;
}

export interface PrimaryDocument {
  id: number;
  fileName: string;
  fileNameOriginal?: string;
  creatorName?: string;
  creationTime?: string;
  formatVerification?: FormatVerification;
}

export interface FormatVerification {
  fileIdentificationResults: ToolResult[];
  fileValidationResults: ToolResult[];
  summary: Summary;
}

export interface ToolResult {
  toolName: string;
  toolVersion: string;
  toolOutput: string;
  outputFormat: 'text' | 'json' | 'csv';
  extractedFeatures: { [key: string]: string };
  error: string;
}

export interface Summary {
  [key: string]: Feature;
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

export interface GeneralMetadata {
  id: number;
  subject?: string;
  xdomeaID?: string;
  filePlan?: FilePlan;
  confidentialityLevel?: ConfidentialityLevel;
  medium?: Medium;
}

export interface ConfidentialityLevel {
  code: string;
  shortDesc: string;
  desc: string;
}

export interface ArchiveMetadata {
  id: number;
  appraisalCode: string;
  appraisalRecommCode: string;
}

export interface AppraisalCode {
  id: number;
  code: string;
  shortDesc: string;
  desc: string;
}

export interface Medium {
  code: string;
  desc: string;
  shortDesc: string;
}

export interface FilePlan {
  id: number;
  xdomeaID?: number;
}

export interface Lifetime {
  id: number;
  start?: string;
  end?: string;
}

export type RecordObjectType = 'file' | 'process' | 'document';

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  private apiEndpoint: string;
  private appraisalCodes = new BehaviorSubject<AppraisalCode[] | null>(null);
  private confidentialityLevelCodelist?: ConfidentialityLevel[];

  private featureOrder: Map<string, number>;
  private overviewFeatures: string[];

  private cachedMessageId?: string;
  private cachedMessage?: Observable<Message>;

  constructor(
    private datePipe: DatePipe,
    private httpClient: HttpClient,
  ) {
    this.apiEndpoint = environment.endpoint;
    this.fetchAppraisalCodelist().subscribe((codes) => this.appraisalCodes.next(codes));
    this.getConfidentialityLevelCodelist().subscribe((confidentialityLevelCodelist: ConfidentialityLevel[]) => {
      this.confidentialityLevelCodelist = confidentialityLevelCodelist;
    });
    this.overviewFeatures = ['relativePath', 'fileName', 'fileSize', 'puid', 'mimeType', 'formatVersion', 'valid'];
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

  getMessage(id: string): Observable<Message> {
    if (id == null) {
      throw new Error('called getMessage with empty string');
    }
    if (id !== this.cachedMessageId) {
      this.cachedMessageId = id;
      this.cachedMessage = this.httpClient
        .get<Message>(this.apiEndpoint + '/message/' + id)
        .pipe(shareReplay({ bufferSize: 1, refCount: true }));
    }
    return this.cachedMessage!;
  }

  getFileRecordObject(id: string): Observable<FileRecordObject> {
    return this.httpClient.get<FileRecordObject>(this.apiEndpoint + '/file-record-object/' + id);
  }

  getProcessRecordObject(id: string): Observable<ProcessRecordObject> {
    return this.httpClient.get<ProcessRecordObject>(this.apiEndpoint + '/process-record-object/' + id);
  }

  getDocumentRecordObject(id: string): Observable<DocumentRecordObject> {
    return this.httpClient.get<DocumentRecordObject>(this.apiEndpoint + '/document-record-object/' + id);
  }

  get0501Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0501');
  }

  get0503Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0503');
  }

  getPrimaryDocument(messageID: string, primaryDocumentID: number): Observable<Blob> {
    const url = this.apiEndpoint + '/primary-document';
    const options = {
      params: new HttpParams().set('messageID', messageID).set('primaryDocumentID', primaryDocumentID),
      responseType: 'blob' as 'json', // https://github.com/angular/angular/issues/18586
    };
    return this.httpClient.get<Blob>(url, options);
  }

  getPrimaryDocuments(id: string): Observable<PrimaryDocument[]> {
    if (!id) {
      throw new Error('called getPrimaryDocuments with null ID');
    }
    const url = this.apiEndpoint + '/primary-documents/' + id;
    return this.httpClient.get<PrimaryDocument[]>(url);
  }

  finalizeMessageAppraisal(messageId: string): Observable<void> {
    const url = this.apiEndpoint + '/finalize-message-appraisal/' + messageId;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  archive0503Message(messageId: string, collectionId?: number): Observable<void> {
    let url = this.apiEndpoint + '/archive-0503-message/' + messageId;
    if (collectionId) {
      url += '?collectionId=' + collectionId;
    }
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  getAppraisalCodelist(): Observable<AppraisalCode[]> {
    return this.appraisalCodes.pipe(filter(notNull));
  }

  private fetchAppraisalCodelist(): Observable<AppraisalCode[]> {
    return this.httpClient.get<AppraisalCode[]>(this.apiEndpoint + '/appraisal-codelist');
  }

  getConfidentialityLevelCodelist(): Observable<ConfidentialityLevel[]> {
    if (this.confidentialityLevelCodelist) {
      return new Observable((subscriber: Subscriber<ConfidentialityLevel[]>) => {
        subscriber.next(this.confidentialityLevelCodelist);
        subscriber.complete();
      });
    } else {
      return this.httpClient.get<ConfidentialityLevel[]>(this.apiEndpoint + '/confidentiality-level-codelist');
    }
  }

  getRecordObjectAppraisalByCode(code: string | undefined, appraisals: AppraisalCode[]): AppraisalCode | null {
    if (!code) {
      return null;
    }
    const appraisal = appraisals.find((appraisal: AppraisalCode) => appraisal.code === code);
    if (!appraisal) {
      throw new Error('record object appraisal with code <' + code + "> wasn't found");
    }
    return appraisal;
  }

  areAllRecordObjectsAppraised(id: string): Observable<boolean> {
    return this.httpClient.get<boolean>(this.apiEndpoint + '/all-record-objects-appraised/' + id);
  }

  getMessageTypeCode(id: string): Observable<string> {
    return this.httpClient.get<string>(this.apiEndpoint + '/message-type-code/' + id);
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
}
