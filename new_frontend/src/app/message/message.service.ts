// angular
import { Injectable } from '@angular/core';
import { DatePipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

// utility
import { Observable } from 'rxjs';

export interface Message {
  id: number;
  messageType: MessageType;
  creationTime: string;
  xdomeaVersion: string;
  messageHead: MessageHead;
  recordObjects: RecordObject[];
  appraisalComplete: boolean;
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
  code?: Code;
  prefix?: Code;
}

export interface Institution {
  id: number;
  name?: string;
  abbreviation?: string;
}

export interface RecordObject {
  id: number;
  fileRecordObject?: FileRecordObject;
}

export interface FileRecordObject {
  id: number;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  processes: ProcessRecordObject[];
}

export interface ProcessRecordObject {
  id: number;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  documents: DocumentRecordObject[];
}

export interface DocumentRecordObject {
  id: number;
  generalMetadata?: GeneralMetadata;
  type?: string;
  incomingDate?: string;
  outgoingDate?: string;
  documentDate?: string;
}

export interface GeneralMetadata {
  id: number;
  subject?: string;
  xdomeaID?: string;
  filePlan?: FilePlan;
}

export interface ArchiveMetadata {
  id: number;
  appraisalCode: string;
  appraisalRecommCode: string;
}

export interface RecordObjectAppraisal {
  id: number;
  code: string;
  shortDesc: string;
  desc: string;
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

export interface Code {
  id: number;
  code?: string;
  name?: string;
}

@Injectable({
  providedIn: 'root'
})
export class MessageService {

  apiEndpoint: string;

  constructor(
    private datePipe: DatePipe,
    private httpClient: HttpClient,
  ) {
    this.apiEndpoint = environment.endpoint;
  }

  getMessage(id: number): Observable<Message> {
    return this.httpClient.get<Message>(this.apiEndpoint + '/message/' + id);
  }

  getFileRecordObject(id: number): Observable<FileRecordObject> {
    return this.httpClient.get<FileRecordObject>(this.apiEndpoint + '/file-record-object/' + id);
  }

  getProcessRecordObject(id: number): Observable<ProcessRecordObject> {
    return this.httpClient.get<ProcessRecordObject>(this.apiEndpoint + '/process-record-object/' + id);
  }

  getDocumentRecordObject(id: number): Observable<DocumentRecordObject> {
    return this.httpClient.get<DocumentRecordObject>(this.apiEndpoint + '/document-record-object/' + id);
  }

  get0501Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0501');
  }

  get0503Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0503');
  }

  getRecordObjectAppraisals(): Observable<RecordObjectAppraisal[]> {
    return this.httpClient.get<RecordObjectAppraisal[]>(this.apiEndpoint + '/record-object-appraisals');
  }

  getRecordObjectAppraisalByCode(code: string | undefined, appraisals: RecordObjectAppraisal[]): RecordObjectAppraisal | null {
    if (!code) {
      return null;
    }
    const appraisal = appraisals.find((appraisal: RecordObjectAppraisal) => appraisal.code === code);
    if (!appraisal) {
      throw new Error('record object appraisal with code <' + code + '> wasn\'t found');
    }
    return appraisal;
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
