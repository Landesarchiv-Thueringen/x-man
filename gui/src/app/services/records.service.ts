import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { AppraisalCode } from './appraisal.service';
import { MessageType } from './message.service';

export interface Records {
  files?: FileRecord[];
  processes?: ProcessRecord[];
  documents?: DocumentRecord[];
}

export interface FileRecord {
  recordId: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type: string;
  subfiles: FileRecord[];
  processes: ProcessRecord[];
}

export interface ProcessRecord {
  recordId: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type: string;
  subprocesses: ProcessRecord[];
  documents: DocumentRecord[];
}

export interface DocumentRecord {
  recordId: string;
  generalMetadata?: GeneralMetadata;
  type: string;
  incomingDate: string;
  outgoingDate: string;
  documentDate: string;
  versions?: DocumentVersion[];
  attachments?: DocumentRecord[];
}

export interface DocumentVersion {
  versionID: string;
  formats: Format[];
}

export interface GeneralMetadata {
  subject: string;
  recordNumber: string;
  leadership: string;
  fileManager: string;
  filePlan?: FilePlan;
  confidentialityLevel?: ConfidentialityLevel;
  medium?: Medium;
}

export type ConfidentialityLevel = '001' | '002' | '003' | '004' | '005';
export type Medium = '001' | '002' | '003';

export interface ArchiveMetadata {
  appraisalCode: AppraisalCode;
  appraisalRecommCode: AppraisalCode;
}

export interface FilePlan {
  filePlanNumber: string;
  subject: string;
}

export interface Lifetime {
  start: string;
  end: string;
}

export interface Format {
  code: string;
  otherName: string;
  version: string;
  primaryDocument: PrimaryDocument;
}

export interface PrimaryDocument {
  filename: string;
  filenameOriginal: string;
  creatorName: string;
  creationTime: string;
}

@Injectable({
  providedIn: 'root',
})
export class RecordsService {
  constructor(private httpClient: HttpClient) {}

  getRootRecords(processId: string, messageType: MessageType): Observable<Records> {
    const url = environment.endpoint + '/root-records/' + processId + '/' + messageType;
    return this.httpClient.get<Records>(url);
  }
}
