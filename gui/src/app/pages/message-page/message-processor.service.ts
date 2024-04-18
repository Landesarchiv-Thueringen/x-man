import { Injectable } from '@angular/core';
import { Subject, first, firstValueFrom, map, startWith } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import { Config, ConfigService } from '../../services/config.service';
import {
  DocumentRecordObject,
  FileRecordObject,
  GeneralMetadata,
  Message,
  ProcessRecordObject,
} from '../../services/message.service';
import { Process } from '../../services/process.service';
import { notNull } from '../../utils/predicates';

export type StructureNodeType =
  | 'message'
  | 'file'
  | 'subfile'
  | 'process'
  | 'subprocess'
  | 'document'
  | 'attachment'
  | 'primaryDocuments';

export interface StructureNode {
  id: string;
  type: StructureNodeType;
  title: string;
  subtitle?: string;
  parentID?: string;
  xdomeaID?: string;
  routerLink?: string;
  generalMetadata?: GeneralMetadata;
  /**
   * Whether the node can be appraised by the user with the UI.
   *
   * Note that this is different from whether there is an appraisal field for
   * the node in xdomea and from whether we should include the node when sending
   * appraisal requests to the backend. Both is true if and only if the node's
   * type is one of 'file' | 'subfile' | 'process' | 'subprocess'.
   */
  canBeAppraised: boolean;
  children?: StructureNode[];
}

/**
 * Processes a given message into a tree structure.
 */
@Injectable()
export class MessageProcessorService {
  private config?: Config;
  private readonly nodes: Map<string, StructureNode> = new Map<string, StructureNode>();
  private messageProcessed = new Subject<void>();

  constructor(private configService: ConfigService) {}

  /**
   * Processes the given message into a tree structure and returns the tree's
   * root node.
   *
   * Also save's the generated tree in the service, so subsequent calls to
   * `getNode` will return nodes from the new tree.
   */
  async processMessage(process: Process, message: Message): Promise<StructureNode> {
    this.config = await firstValueFrom(this.configService.config);
    this.nodes.clear();
    // Create message node
    const messageNode = this.getMessageNode(process, message);
    // Create file-verification node
    if (message.messageType?.code === '0503') {
      messageNode.children!.push(this.getPrimaryDocumentsNode(message.id));
    }
    // Add file nodes
    for (const fileRecordObject of message.fileRecordObjects ?? []) {
      messageNode.children!.push(this.getFileStructureNode(fileRecordObject, messageNode));
    }
    // Add process nodes
    message.processRecordObjects ??= [];
    for (const processRecordObject of message.processRecordObjects ?? []) {
      messageNode.children!.push(this.getProcessStructureNode(processRecordObject, messageNode));
    }
    // Add document nodes
    for (const documentRecordObject of message.documentRecordObjects ?? []) {
      messageNode.children!.push(this.getDocumentStructureNode(documentRecordObject, messageNode));
    }
    this.messageProcessed.next();
    return messageNode;
  }

  /**
   * Returns the structure node with the given ID.
   *
   * Throws an error if the node cannot be found in the tree.
   */
  getNode(id: string): StructureNode {
    const node = this.nodes.get(id);
    if (node == null) {
      throw new Error('node not found: ' + id);
    }
    return node;
  }

  /**
   * Returns a promise that resolves to the structure node with the given ID as
   * soon as it becomes available in the tree.
   *
   * For a node to become available `processMessage` has to be called with a
   * message containing the node.
   */
  async getNodeWhenReady(id: string): Promise<StructureNode> {
    return firstValueFrom(
      this.messageProcessed.pipe(
        startWith(void 0),
        map(() => this.nodes.get(id)),
        first(notNull),
      ),
    );
  }

  private canBeAppraised(type: StructureNode['type'], parent: StructureNode): boolean {
    switch (this.config!.appraisalLevel) {
      case 'root':
        return parent.type === 'message';
      case 'all':
        return type === 'file' || type === 'subfile' || type === 'process' || type === 'subprocess';
      default:
        console.error('called canBeAppraised when config was not ready');
        return false;
    }
  }

  private getMessageNode(process: Process, message: Message): StructureNode {
    let title: string;
    switch (message.messageType.code) {
      case '0501':
        title = 'Anbietung';
        break;
      case '0503':
        title = 'Abgabe';
        break;
      default:
        throw new Error('unhandled message type');
    }
    const messageNode: StructureNode = {
      id: message.id,
      title,
      subtitle: process.agency.name,
      type: 'message',
      routerLink: 'details',
      canBeAppraised: false,
      children: [],
    };
    this.nodes.set(messageNode.id, messageNode);
    return messageNode;
  }

  private getPrimaryDocumentsNode(messageID: string): StructureNode {
    const routerLink: string = 'formatverifikation';
    const primaryDocumentsNode: StructureNode = {
      id: uuidv4(),
      title: 'Formatverifikation',
      subtitle: 'Primärdateien',
      type: 'primaryDocuments',
      routerLink: routerLink,
      parentID: messageID,
      canBeAppraised: false,
    };
    this.nodes.set(primaryDocumentsNode.id, primaryDocumentsNode);
    return primaryDocumentsNode;
  }

  private getFileStructureNode(fileRecordObject: FileRecordObject, parent: StructureNode): StructureNode {
    const children: StructureNode[] = [];
    const type = parent.type.endsWith('file') ? 'subfile' : 'file';
    const nodeName = type === 'file' ? 'Akte' : 'Teilakte';
    const routerLink: string = 'akte/' + fileRecordObject.id;
    const fileNode: StructureNode = {
      id: fileRecordObject.id,
      title: nodeName + ': ' + fileRecordObject.generalMetadata?.xdomeaID,
      subtitle: fileRecordObject.generalMetadata?.subject,
      xdomeaID: fileRecordObject.xdomeaID,
      type,
      routerLink,
      parentID: parent.id,
      generalMetadata: fileRecordObject.generalMetadata,
      children,
      canBeAppraised: this.canBeAppraised(type, parent),
    };
    // generate child nodes for all subfiles (de: Teilakten)
    if (fileRecordObject.subfiles) {
      for (let subfile of fileRecordObject.subfiles) {
        children.push(this.getFileStructureNode(subfile, fileNode));
      }
    }
    // generate child nodes for all processes (de: Vorgänge)
    if (fileRecordObject.processes) {
      for (let process of fileRecordObject.processes) {
        children.push(this.getProcessStructureNode(process, fileNode));
      }
    }

    this.nodes.set(fileNode.id, fileNode);
    return fileNode;
  }

  private getProcessStructureNode(processRecordObject: ProcessRecordObject, parent: StructureNode): StructureNode {
    const children: StructureNode[] = [];
    const routerLink: string = 'vorgang/' + processRecordObject.id;
    const type = parent.type.endsWith('process') ? 'subprocess' : 'process';
    const nodeName = type === 'process' ? 'Vorgang' : 'Teilvorgang';
    const processNode: StructureNode = {
      id: processRecordObject.id,
      title: nodeName + ': ' + processRecordObject.generalMetadata?.xdomeaID,
      subtitle: processRecordObject.generalMetadata?.subject,
      xdomeaID: processRecordObject.xdomeaID,
      type: type,
      routerLink: routerLink,
      parentID: parent.id,
      generalMetadata: processRecordObject.generalMetadata,
      canBeAppraised: this.canBeAppraised(type, parent),
      children: children,
    };
    // generate child nodes for all subprocesses (de: Teilvorgänge)
    if (processRecordObject.subprocesses) {
      for (let subprocess of processRecordObject.subprocesses) {
        children.push(this.getProcessStructureNode(subprocess, processNode));
      }
    }
    // generate child nodes for all documents (de: Dokumente)
    if (processRecordObject.documents) {
      for (let document of processRecordObject.documents) {
        children.push(this.getDocumentStructureNode(document, processNode));
      }
    }

    this.nodes.set(processNode.id, processNode);
    return processNode;
  }

  private getDocumentStructureNode(documentRecordObject: DocumentRecordObject, parent: StructureNode): StructureNode {
    const children: StructureNode[] = [];
    const type = parent.type === 'document' || parent.type === 'attachment' ? 'attachment' : 'document';
    const nodeName = type === 'attachment' ? 'Anlage' : 'Dokument';
    const routerLink: string = 'dokument/' + documentRecordObject.id;
    const documentNode: StructureNode = {
      id: documentRecordObject.id,
      title: nodeName + ': ' + documentRecordObject.generalMetadata?.xdomeaID,
      subtitle: documentRecordObject.generalMetadata?.subject,
      xdomeaID: documentRecordObject.xdomeaID,
      type: type,
      routerLink: routerLink,
      parentID: parent.id,
      generalMetadata: documentRecordObject.generalMetadata,
      canBeAppraised: false,
      children: children,
    };
    if (documentRecordObject.attachments) {
      for (let document of documentRecordObject.attachments) {
        children.push(this.getDocumentStructureNode(document, documentNode));
      }
    }
    this.nodes.set(documentNode.id, documentNode);
    return documentNode;
  }
}
