import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

export interface Message {
  id: number;
  messageHead: MessageHead;
  recordObjects: RecordObject[];
}

export interface MessageType {
  id: number;
  code: string;
}

export interface MessageHead {
  id: number;
  processID: string;
  sender: Contact;
  receiver: Contact;
}

export interface Contact {
  id: number;
  agencyIdentification: AgencyIdentification;
  institution: Institution;
}

export interface AgencyIdentification {
  id: number;
  code: Code;
  prefix: Code;
}

export interface Institution {
  id: number;
  name: string;
  abbreviation: string;
}

export interface RecordObject {
  id: number;
  fileRecordObject?: FileRecordObject;
}

export interface FileRecordObject {
  id: number;
  generalMetadata: GeneralMetadata;
  lifetime: Lifetime;
}

export interface GeneralMetadata {
  id: number;
  subject: string;
  xdomeaID: string;
  filePlan: FilePlan;
}

export interface FilePlan {
  id: number;
  xdomeaID: number;
}

export interface Lifetime {
  id: number;
  start: string;
  end: string;
}

export interface Code {
  id: number;
  code: string;
  name: string;
}

@Injectable({
  providedIn: 'root'
})
export class MessageService {

  constructor(private httpClient: HttpClient) { }
}
