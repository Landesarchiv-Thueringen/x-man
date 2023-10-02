// angular
import { Injectable } from '@angular/core';
import { DatePipe } from '@angular/common';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from '../../environments/environment';

// utility
import { Observable, BehaviorSubject, Subject, Subscriber } from 'rxjs';

export interface Message {
  id: string;
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
  id: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  processes: ProcessRecordObject[];
  recordObjectType: RecordObjectType;
}

export interface ProcessRecordObject {
  id: string;
  generalMetadata?: GeneralMetadata;
  archiveMetadata?: ArchiveMetadata;
  lifetime?: Lifetime;
  type?: string;
  documents: DocumentRecordObject[];
  recordObjectType: RecordObjectType;
}

export interface DocumentRecordObject {
  id: string;
  generalMetadata?: GeneralMetadata;
  type?: string;
  incomingDate?: string;
  outgoingDate?: string;
  documentDate?: string;
  recordObjectType: RecordObjectType;
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

export type StructureNodeType = 'message' | 'file' | 'process' | 'document';
export type RecordObjectType = 'file' | 'process' | 'document';

export interface DisplayText {
  title: string;
  subtitle?: string;
}

export interface StructureNode {
  id: string;
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  appraisal?: string;
  parentID?: string;
  children?: StructureNode[];
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

  constructor(private datePipe: DatePipe, private httpClient: HttpClient) {
    this.apiEndpoint = environment.endpoint;
    this.structureNodes = new Map<string, StructureNode>();
    this.nodesSubject = new BehaviorSubject<StructureNode[]>(
      this.getRootStructureNodes()
    );
    this.changedNodeSubject = new Subject<StructureNode>();
    this.getRecordObjectAppraisals().subscribe(
      (appraisals: RecordObjectAppraisal[]) => {
        this.appraisals = appraisals;
      }
    );
    this.getRecordObjectConfidentialities().subscribe(
      (confidentialities: RecordObjectConfidentiality[]) => {
        this.confidentialities = confidentialities;
      }
    );
  }

  processMessage(message: Message): StructureNode {
    const children: StructureNode[] = [];
    for (let recordObject of message.recordObjects) {
      if (recordObject.fileRecordObject) {
        children.push(
          this.getFileStructureNode(recordObject.fileRecordObject, message.id)
        );
      }
    }
    let displayText: DisplayText;
    switch (message.messageType.code) {
      case '0501':
        displayText = {
          title: 'Anbietung',
        };
        break;
      case '0503':
        displayText = {
          title: 'Abgabe',
        };
        break;
      default:
        throw new Error('unhandled message type');
    }
    const routerLink: string = 'details';
    const messageNode: StructureNode = {
      id: message.id,
      displayText: displayText,
      type: 'message',
      routerLink: routerLink,
      children: children,
    };
    this.structureNodes.set(messageNode.id, messageNode);
    this.nodesSubject.next(this.getRootStructureNodes());
    return messageNode;
  }

  getFileStructureNode(
    fileRecordObject: FileRecordObject,
    parentID?: string
  ): StructureNode {
    const children: StructureNode[] = [];
    for (let process of fileRecordObject.processes) {
      children.push(this.getProcessStructureNode(process, fileRecordObject.id));
    }
    const displayText: DisplayText = {
      title: 'Akte: ' + fileRecordObject.generalMetadata?.xdomeaID,
      subtitle: fileRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'akte/' + fileRecordObject.id;
    const fileNode: StructureNode = {
      id: fileRecordObject.id,
      displayText: displayText,
      type: 'file',
      routerLink: routerLink,
      appraisal: fileRecordObject.archiveMetadata?.appraisalCode,
      parentID: parentID,
      children: children,
    };
    this.addStructureNode(fileNode);
    return fileNode;
  }

  getProcessStructureNode(
    processRecordObject: ProcessRecordObject,
    parentID?: string
  ): StructureNode {
    const children: StructureNode[] = [];
    for (let document of processRecordObject.documents) {
      children.push(
        this.getDocumentStructureNode(document, processRecordObject.id)
      );
    }
    const displayText: DisplayText = {
      title: 'Vorgang: ' + processRecordObject.generalMetadata?.xdomeaID,
      subtitle: processRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'vorgang/' + processRecordObject.id;
    const processNode: StructureNode = {
      id: processRecordObject.id,
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

  getDocumentStructureNode(
    documentRecordObject: DocumentRecordObject,
    parentID?: string
  ): StructureNode {
    const displayText: DisplayText = {
      title: 'Dokument: ' + documentRecordObject.generalMetadata?.xdomeaID,
      subtitle: documentRecordObject.generalMetadata?.subject,
    };
    const routerLink: string = 'dokument/' + documentRecordObject.id;
    const documentNode: StructureNode = {
      id: documentRecordObject.id,
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

  propagateNodeChangeToParents(node: StructureNode): void {
    if (!node.parentID) {
      throw new Error('no parent for node change propagation existing');
    }
    const parent: StructureNode | undefined = this.structureNodes.get(
      node.parentID
    );
    if (!parent) {
      throw new Error('parent node doesn"t exist, ID: ' + node.parentID);
    }
    if (!parent.children) {
      throw new Error('parent and children are not connected');
    }
    const nodeIndex: number = parent.children.findIndex(
      (child: StructureNode) => child.id === node.id
    );
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
    return this.httpClient.get<FileRecordObject>(
      this.apiEndpoint + '/file-record-object/' + id
    );
  }

  getProcessRecordObject(id: string): Observable<ProcessRecordObject> {
    return this.httpClient.get<ProcessRecordObject>(
      this.apiEndpoint + '/process-record-object/' + id
    );
  }

  getDocumentRecordObject(id: string): Observable<DocumentRecordObject> {
    return this.httpClient.get<DocumentRecordObject>(
      this.apiEndpoint + '/document-record-object/' + id
    );
  }

  get0501Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0501');
  }

  get0503Messages(): Observable<Message[]> {
    return this.httpClient.get<Message[]>(this.apiEndpoint + '/messages/0503');
  }

  finalizeMessageAppraisal(id: string): Observable<void> {
    const url = this.apiEndpoint + '/finalize-message-appraisal/' + id;
    const body = {};
    const options = {};
    return this.httpClient.patch<void>(url, body, options);
  }

  getRecordObjectAppraisals(): Observable<RecordObjectAppraisal[]> {
    if (this.appraisals) {
      return new Observable(
        (subscriber: Subscriber<RecordObjectAppraisal[]>) => {
          subscriber.next(this.appraisals);
          subscriber.complete();
        }
      );
    } else {
      return this.httpClient.get<RecordObjectAppraisal[]>(
        this.apiEndpoint + '/record-object-appraisals'
      );
    }
  }

  getRecordObjectConfidentialities(): Observable<
    RecordObjectConfidentiality[]
  > {
    if (this.confidentialities) {
      return new Observable(
        (subscriber: Subscriber<RecordObjectConfidentiality[]>) => {
          subscriber.next(this.confidentialities);
          subscriber.complete();
        }
      );
    } else {
      return this.httpClient.get<RecordObjectConfidentiality[]>(
        this.apiEndpoint + '/record-object-confidentialities'
      );
    }
  }

  setFileRecordObjectAppraisal(
    id: string,
    appraisalCode: string
  ): Observable<FileRecordObject> {
    const url = this.apiEndpoint + '/file-record-object-appraisal';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('appraisal', appraisalCode),
    };
    return this.httpClient.patch<FileRecordObject>(url, body, options);
  }

  setProcessRecordObjectAppraisal(
    id: string,
    appraisalCode: string
  ): Observable<ProcessRecordObject> {
    const url = this.apiEndpoint + '/process-record-object-appraisal';
    const body = {};
    const options = {
      params: new HttpParams().set('id', id).set('appraisal', appraisalCode),
    };
    if (this.structureNodes.get(id)) {
      this.structureNodes.get(id)!.appraisal = appraisalCode;
    }
    return this.httpClient.patch<ProcessRecordObject>(url, body, options);
  }

  updateStructureNode(
    recordObject: FileRecordObject | ProcessRecordObject | DocumentRecordObject
  ): void {
    const node: StructureNode | undefined = this.structureNodes.get(
      recordObject.id
    );
    if (node) {
      let changedNode: StructureNode;
      switch (recordObject.recordObjectType) {
        case 'file': {
          changedNode = this.getFileStructureNode(
            recordObject as FileRecordObject,
            node.parentID
          );
          break;
        }
        case 'process': {
          changedNode = this.getProcessStructureNode(
            recordObject as ProcessRecordObject,
            node.parentID
          );
          break;
        }
        case 'document': {
          changedNode = this.getDocumentStructureNode(
            recordObject as DocumentRecordObject,
            node.parentID
          );
          break;
        }
      }
      // we accept you my node
      this.propagateNodeChangeToParents(changedNode);
      this.nodesSubject.next(this.getRootStructureNodes());
      this.changedNodeSubject.next(changedNode);
    } else {
      console.error(
        'no structure node for record object with ID: ' + recordObject.id
      );
    }
  }

  getRecordObjectAppraisalByCode(
    code: string | undefined,
    appraisals: RecordObjectAppraisal[]
  ): RecordObjectAppraisal | null {
    if (!code) {
      return null;
    }
    const appraisal = appraisals.find(
      (appraisal: RecordObjectAppraisal) => appraisal.code === code
    );
    if (!appraisal) {
      throw new Error(
        'record object appraisal with code <' + code + "> wasn't found"
      );
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
