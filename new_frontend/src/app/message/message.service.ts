import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable } from 'rxjs';

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
  creationTime: string;
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

  apiEndpoint: string;

  constructor(private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
  }

  get0501Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0501');
  }

  get0503Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0503');
  }
}
