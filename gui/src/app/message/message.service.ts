// angular
import { DatePipe } from '@angular/common';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment';

// utility
import { BehaviorSubject, Observable, Subject, Subscriber } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import { Process } from '../process/process.service';

export interface Message {
  id: string;
  messageType: MessageType;
  creationTime: string;
  xdomeaVersion: string;
  schemaValidation: boolean;
  messageHead: MessageHead;
  appraisalComplete: boolean;
  formatVerificationComplete: boolean;
  primaryDocumentCount: number;
  verificationCompleteCount: number;
  fileRecordObjects: FileRecordObject[];
  processRecordObjects: ProcessRecordObject[];
  documentRecordObjects: DocumentRecordObject[];
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

export interface FileRecordObject {
  id: string;
  messageID: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  processes: ProcessRecordObject[];
  subfiles: FileRecordObject[];
  recordObjectType: RecordObjectType;
}

export interface ProcessRecordObject {
  id: string;
  messageID: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  documents: DocumentRecordObject[];
  recordObjectType: RecordObjectType;
}

export interface DocumentRecordObject {
  id: string;
  messageID: string;
  generalMetadata?: GeneralMetadata;
  type?: string;
  incomingDate?: string;
  outgoingDate?: string;
  documentDate?: string;
  recordObjectType: RecordObjectType;
  versions?: DocumentVersion[];
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
  confidentialityCode?: string;
}

export interface RecordObjectConfidentiality {
  id: number;
  code: string;
  desc: string;
  shortDesc: string;
}

export interface ArchiveMetadata {
  id: number;
  appraisalCode: string;
  appraisalRecommCode: string;
  internalAppraisalNote?: string;
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

export type StructureNodeType = 'message' | 'file' | 'subfile' | 'process' | 'document' | 'primaryDocuments';
export type RecordObjectType = 'file' | 'process' | 'document';

export interface DisplayText {
  title: string;
  subtitle?: string;
}

export interface StructureNode {
  id: string;
  selected: boolean;
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  appraisal?: string;
  parentID?: string;
  children?: StructureNode[];
}

export interface MultiAppraisalResponse {
  updatedFileRecordObjects: FileRecordObject[];
  updatedProcessRecordObjects: ProcessRecordObject[];
}

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  apiEndpoint: string;
  appraisals?: RecordObjectAppraisal[];
  confidentialities?: RecordObjectConfidentiality[];

  structureNodes: Map<string, StructureNode>;
  nodesSubject: BehaviorSubject<StructureNode[]>;
  changedNodeSubject: Subject<StructureNode>;

  featureOrder: Map<string, number>;
  overviewFeatures: string[];

  constructor(
    private datePipe: DatePipe,
    private httpClient: HttpClient,
  ) {
    this.apiEndpoint = environment.endpoint;
    this.structureNodes = new Map<string, StructureNode>();
    this.nodesSubject = new BehaviorSubject<StructureNode[]>(this.getRootStructureNodes());
    this.changedNodeSubject = new Subject<StructureNode>();
    this.getRecordObjectAppraisals().subscribe((appraisals: RecordObjectAppraisal[]) => {
      this.appraisals = appraisals;
    });
    this.getRecordObjectConfidentialities().subscribe((confidentialities: RecordObjectConfidentiality[]) => {
      this.confidentialities = confidentialities;
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

  processMessage(process: Process, message: Message): StructureNode {
    const children: StructureNode[] = [];
    if (message.messageType?.code === '0503') {
      children.push(this.getPrimaryDocumentsNode(message.id));
    }
    for (let fileRecordObject of message.fileRecordObjects) {
      children.push(this.getFileStructureNode(fileRecordObject, false, message.id));
    }
    let displayText: DisplayText;
    switch (message.messageType.code) {
      case '0501':
        displayText = {
          title: 'Anbietung',
          subtitle: process.agency.name,
        };
        break;
      case '0503':
        displayText = {
          title: 'Abgabe',
          subtitle: process.agency.name,
        };
        break;
      default:
        throw new Error('unhandled message type');
    }
    const routerLink: string = 'details';
    const messageNode: StructureNode = {
      id: message.id,
      selected: true,
      displayText: displayText,
      type: 'message',
      routerLink: routerLink,
      children: children,
    };
    this.structureNodes.set(messageNode.id, messageNode);
    this.nodesSubject.next(this.getRootStructureNodes());
    return messageNode;
  }

  getPrimaryDocumentsNode(messageID: string): StructureNode {
    const displayText: DisplayText = {
      title: 'Formatverifikation',
      subtitle: 'Primärdateien',
    };
    const routerLink: string = 'formatverifikation';
    const primaryDocumentsNode: StructureNode = {
      id: uuidv4(),
      selected: false,
      displayText: displayText,
      type: 'primaryDocuments',
      routerLink: routerLink,
      parentID: messageID,
    };
    return primaryDocumentsNode;
  }

  getFileStructureNode(fileRecordObject: FileRecordObject, subfile: boolean, parentID?: string): StructureNode {
    const children: StructureNode[] = [];
    // generate child nodes for all sub files (de: Teilakten)
    for (let subfile of fileRecordObject.subfiles) {
      children.push(this.getFileStructureNode(subfile, true, fileRecordObject.id));
    }
    // generate child nodes for all processes (de: Vorgänge)
    for (let process of fileRecordObject.processes) {
      children.push(this.getProcessStructureNode(process, fileRecordObject.id));
    }
    const nodeName = subfile ? 'Teilakte' : 'Akte';
    const displayText: DisplayText = {
      title: nodeName + ': ' + fileRecordObject.generalMetadata?.xdomeaID,
      subtitle: fileRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'akte/' + fileRecordObject.id;
    const type = subfile ? 'subfile' : 'file';
    const fileNode: StructureNode = {
      id: fileRecordObject.id,
      selected: false,
      displayText: displayText,
      type: type,
      routerLink: routerLink,
      appraisal: fileRecordObject.archiveMetadata?.appraisalCode,
      parentID: parentID,
      children: children,
    };
    this.addStructureNode(fileNode);
    return fileNode;
  }

  getProcessStructureNode(processRecordObject: ProcessRecordObject, parentID?: string): StructureNode {
    const children: StructureNode[] = [];
    for (let document of processRecordObject.documents) {
      children.push(this.getDocumentStructureNode(document, processRecordObject.id));
    }
    const displayText: DisplayText = {
      title: 'Vorgang: ' + processRecordObject.generalMetadata?.xdomeaID,
      subtitle: processRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'vorgang/' + processRecordObject.id;
    const processNode: StructureNode = {
      id: processRecordObject.id,
      selected: false,
      displayText: displayText,
      type: 'process',
      routerLink: routerLink,
      appraisal: processRecordObject.archiveMetadata?.appraisalCode,
      parentID: parentID,
      children: children,
    };
    this.addStructureNode(processNode);
    return processNode;
  }

  getDocumentStructureNode(documentRecordObject: DocumentRecordObject, parentID?: string): StructureNode {
    const displayText: DisplayText = {
      title: 'Dokument: ' + documentRecordObject.generalMetadata?.xdomeaID,
      subtitle: documentRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'dokument/' + documentRecordObject.id;
    const documentNode: StructureNode = {
      id: documentRecordObject.id,
      selected: false,
      displayText: displayText,
      type: 'document',
      routerLink: routerLink,
      parentID: parentID,
    };
    this.addStructureNode(documentNode);
    return documentNode;
  }

  watchStructureNodes(): Observable<StructureNode[]> {
    return this.nodesSubject.asObservable();
  }

  watchNodeChanges(): Observable<StructureNode> {
    return this.changedNodeSubject.asObservable();
  }

  addStructureNode(node: StructureNode): void {
    this.structureNodes.set(node.id, node);
  }

  getStructureNode(id: string): StructureNode | undefined {
    return this.structureNodes.get(id);
  }

  propagateNodeChangeToParents(node: StructureNode): void {
    if (!node.parentID) {
      throw new Error('no parent for node change propagation existing');
    }
    const parent: StructureNode | undefined = this.structureNodes.get(node.parentID);
    if (!parent) {
      throw new Error('parent node doesn"t exist, ID: ' + node.parentID);
    }
    if (!parent.children) {
      throw new Error('parent and children are not connected');
    }
    const nodeIndex: number = parent.children.findIndex((child: StructureNode) => child.id === node.id);
    if (nodeIndex === -1) {
      throw new Error('parent and child are not connected');
    }
    // replace old node with updated version
    parent.children[nodeIndex] = node;
  }

  getMessage(id: string): Observable<Message> {
    return this.httpClient.get<Message>(this.apiEndpoint + '/message/' + id);
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
    const url = this.apiEndpoint + '/primary-documents/' + id;
    return this.httpClient.get<PrimaryDocument[]>(url);
  }

  finalizeMessageAppraisal(id: string): Observable<void> {
    const url = this.apiEndpoint + '/finalize-message-appraisal/' + id;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  archive0503Message(id: string): Observable<void> {
    const url = this.apiEndpoint + '/archive-0503-message/' + id;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  getRecordObjectAppraisals(): Observable<RecordObjectAppraisal[]> {
    if (this.appraisals) {
      return new Observable((subscriber: Subscriber<RecordObjectAppraisal[]>) => {
        subscriber.next(this.appraisals);
        subscriber.complete();
      });
    } else {
      return this.httpClient.get<RecordObjectAppraisal[]>(this.apiEndpoint + '/record-object-appraisals');
    }
  }

  getRecordObjectConfidentialities(): Observable<RecordObjectConfidentiality[]> {
    if (this.confidentialities) {
      return new Observable((subscriber: Subscriber<RecordObjectConfidentiality[]>) => {
        subscriber.next(this.confidentialities);
        subscriber.complete();
      });
    } else {
      return this.httpClient.get<RecordObjectConfidentiality[]>(this.apiEndpoint + '/record-object-confidentialities');
    }
  }

  setFileRecordObjectAppraisal(id: string, appraisalCode: string): Observable<FileRecordObject> {
    const url = this.apiEndpoint + '/file-record-object-appraisal';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('appraisal', appraisalCode),
    };
    return this.httpClient.patch<FileRecordObject>(url, body, options);
  }

  setFileRecordObjectAppraisalNote(id: string, note?: string | null): Observable<FileRecordObject> {
    const url = this.apiEndpoint + '/file-record-object-appraisal-note';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('note', note ? note : ''),
    };
    return this.httpClient.patch<FileRecordObject>(url, body, options);
  }

  setProcessRecordObjectAppraisal(id: string, appraisalCode: string): Observable<ProcessRecordObject> {
    const url = this.apiEndpoint + '/process-record-object-appraisal';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('appraisal', appraisalCode),
    };
    return this.httpClient.patch<ProcessRecordObject>(url, body, options);
  }

  setProcessRecordObjectAppraisalNote(id: string, note: string | null): Observable<ProcessRecordObject> {
    const url = this.apiEndpoint + '/process-record-object-appraisal-note';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('note', note ? note : ''),
    };
    return this.httpClient.patch<ProcessRecordObject>(url, body, options);
  }

  setAppraisalForMultipleRecorcObjects(
    recordObjectIDs: string[],
    appraisalCode: string,
    appraisalNote: string | null,
  ): Observable<MultiAppraisalResponse> {
    const fileRecordObjectIDs: string[] = [];
    const processRecordObjectIDs: string[] = [];
    for (let id of recordObjectIDs) {
      const node: StructureNode | undefined = this.structureNodes.get(id);
      if (!node) {
        throw new Error('record object ID not found');
      }
      if (node.type === 'file') {
        fileRecordObjectIDs.push(node.id);
      } else if (node.type === 'process') {
        processRecordObjectIDs.push(node.id);
      } else {
        throw new Error('appraisal can only be set for file and process record objects');
      }
    }
    const url = this.apiEndpoint + '/multi-appraisal';
    const body = {
      fileRecordObjectIDs: fileRecordObjectIDs,
      processRecordObjectIDs: processRecordObjectIDs,
      appraisalCode: appraisalCode,
      appraisalNote: appraisalNote,
    };
    const options = {};
    return this.httpClient.patch<MultiAppraisalResponse>(url, body, options);
  }

  updateStructureNodeForRecordObject(
    recordObject: FileRecordObject | ProcessRecordObject | DocumentRecordObject,
  ): void {
    const node: StructureNode | undefined = this.structureNodes.get(recordObject.id);
    if (node) {
      let changedNode: StructureNode;
      switch (recordObject.recordObjectType) {
        case 'file': {
          changedNode = this.getFileStructureNode(
            recordObject as FileRecordObject,
            node.type === 'subfile',
            node.parentID,
          );
          break;
        }
        case 'process': {
          changedNode = this.getProcessStructureNode(recordObject as ProcessRecordObject, node.parentID);
          break;
        }
        case 'document': {
          changedNode = this.getDocumentStructureNode(recordObject as DocumentRecordObject, node.parentID);
          break;
        }
      }
      // we accept you my node
      this.propagateNodeChangeToParents(changedNode);
      this.nodesSubject.next(this.getRootStructureNodes());
      this.changedNodeSubject.next(changedNode);
    } else {
      console.error('no structure node for record object with ID: ' + recordObject.id);
    }
  }

  updateStructureNode(changedNode: StructureNode) {
    this.structureNodes.set(changedNode.id, changedNode);
    this.propagateNodeChangeToParents(changedNode);
    this.nodesSubject.next(this.getRootStructureNodes());
    this.changedNodeSubject.next(changedNode);
  }

  getRecordObjectAppraisalByCode(
    code: string | undefined,
    appraisals: RecordObjectAppraisal[],
  ): RecordObjectAppraisal | null {
    if (!code) {
      return null;
    }
    const appraisal = appraisals.find((appraisal: RecordObjectAppraisal) => appraisal.code === code);
    if (!appraisal) {
      throw new Error('record object appraisal with code <' + code + "> wasn't found");
    }
    return appraisal;
  }

  isMessageAppraisalComplete(id: string): Observable<boolean> {
    return this.httpClient.get<boolean>(this.apiEndpoint + '/message-appraisal-complete/' + id);
  }

  areAllRecordObjectsAppraised(id: string): Observable<boolean> {
    return this.httpClient.get<boolean>(this.apiEndpoint + '/all-record-objects-appraised/' + id);
  }

  getMessageTypeCode(id: string): Observable<string> {
    return this.httpClient.get<string>(this.apiEndpoint + '/message-type-code/' + id);
  }

  sortFeatures(features: string[]): string[] {
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

  /**
   * The message tree has only one root node which is the message node.
   */
  private getRootStructureNodes(): StructureNode[] {
    for (let node of this.structureNodes.values()) {
      if (!node.parentID) {
        return [node];
      }
    }
    return [];
  }
}
