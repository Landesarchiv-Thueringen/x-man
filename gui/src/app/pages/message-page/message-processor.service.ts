import { Injectable } from '@angular/core';
import { Subject, first, firstValueFrom, map, startWith } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import { Config, ConfigService } from '../../services/config.service';
import { Message, MessageType } from '../../services/message.service';
import { SubmissionProcess } from '../../services/process.service';
import {
  DocumentRecord,
  FileRecord,
  GeneralMetadata,
  ProcessRecord,
  Records,
} from '../../services/records.service';
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
  // For records `id` is equal to `recordId`. We need this, so we can expand the
  // active record when loading the page. For meta nodes like record groups or
  // `primaryDocuments`, we `id` is a random UUID.
  id: string;
  type: StructureNodeType;
  title: string;
  subtitle?: string;
  parentId?: string;
  recordId?: string;
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
  /**
   * Whether the user can choose a packaging level for the node.
   */
  canChoosePackaging: boolean;
  /**
   * Whether the node should have a checkbox when the user activates
   * multi-selection.
   */
  selectable: boolean;
  children?: StructureNode[];
}

/**
 * Processes a given message into a tree structure.
 */
@Injectable()
export class MessageProcessorService {
  private config?: Config;
  private readonly nodes = new Map<string, StructureNode>();
  private messageProcessed = new Subject<void>();

  constructor(private configService: ConfigService) {}

  /**
   * Processes the given message into a tree structure and returns the tree's
   * root node.
   *
   * Also save's the generated tree in the service, so subsequent calls to
   * `getNode` will return nodes from the new tree.
   */
  async processMessage(
    process: SubmissionProcess,
    message: Message,
    rootRecords: Records,
  ): Promise<StructureNode> {
    this.config = await firstValueFrom(this.configService.config);
    this.nodes.clear();
    // Create message node
    const messageNode = this.getMessageNode(process, message, rootRecords);
    // Create file-verification node
    if (message.messageType === '0503') {
      messageNode.children!.push(this.getPrimaryDocumentsNode(messageNode.id));
    }
    // Add file nodes
    for (const fileRecordObject of rootRecords.files ?? []) {
      messageNode.children!.push(
        this.getFileStructureNode(fileRecordObject, messageNode, message.messageType),
      );
    }
    // Add process nodes
    for (const processRecordObject of rootRecords.processes ?? []) {
      messageNode.children!.push(
        this.getProcessStructureNode(processRecordObject, messageNode, message.messageType),
      );
    }
    // Add document nodes
    for (const documentRecordObject of rootRecords.documents ?? []) {
      messageNode.children!.push(
        this.getDocumentStructureNode(documentRecordObject, messageNode, message.messageType),
      );
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

  private getMessageNode(
    process: SubmissionProcess,
    message: Message,
    rootRecords: Records,
  ): StructureNode {
    let title: string;
    switch (message.messageType) {
      case '0501':
        title = 'Anbietung';
        break;
      case '0503':
        title = 'Abgabe';
        break;
      default:
        throw new Error('unhandled message type');
    }
    const numberElements =
      (rootRecords.files?.length ?? 0) +
      (rootRecords.documents?.length ?? 0) +
      (rootRecords.processes?.length ?? 0);
    title = `${title} (${numberElements} ${numberElements === 1 ? 'Element' : 'Elemente'})`;
    const messageNode: StructureNode = {
      // We use `id` as track-by function. This makes sure the node gets updated
      // when switching between messages.
      id: message.messageHead.processID + '/' + message.messageType,
      title,
      subtitle: process.agency.name,
      type: 'message',
      routerLink: 'details',
      canBeAppraised: false,
      canChoosePackaging: false,
      selectable: true,
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
      parentId: messageID,
      canBeAppraised: false,
      canChoosePackaging: false,
      selectable: false,
    };
    this.nodes.set(primaryDocumentsNode.id, primaryDocumentsNode);
    return primaryDocumentsNode;
  }

  private getFileStructureNode(
    fileRecord: FileRecord,
    parent: StructureNode,
    messageType: MessageType,
  ): StructureNode {
    const children: StructureNode[] = [];
    const type = parent.type.endsWith('file') ? 'subfile' : 'file';
    const nodeName = type === 'file' ? 'Akte' : 'Teilakte';
    const routerLink: string = 'akte/' + fileRecord.recordId;
    const fileNode: StructureNode = {
      id: fileRecord.recordId,
      title: nodeName + ': ' + fileRecord.generalMetadata?.recordNumber,
      subtitle: fileRecord.generalMetadata?.subject,
      recordId: fileRecord.recordId,
      type,
      routerLink,
      parentId: parent.id,
      generalMetadata: fileRecord.generalMetadata,
      children,
      canBeAppraised: this.canBeAppraised(type, parent, messageType),
      canChoosePackaging: this.canChoosePackaging(type, parent),
      selectable: this.isSelectable(type, parent, messageType),
    };
    // generate child nodes for all subfiles (de: Teilakten)
    if (fileRecord.subfiles) {
      for (let subfile of fileRecord.subfiles) {
        children.push(this.getFileStructureNode(subfile, fileNode, messageType));
      }
    }
    // generate child nodes for all processes (de: Vorgänge)
    if (fileRecord.processes) {
      for (let process of fileRecord.processes) {
        children.push(this.getProcessStructureNode(process, fileNode, messageType));
      }
    }

    this.nodes.set(fileNode.id, fileNode);
    return fileNode;
  }

  private getProcessStructureNode(
    processRecord: ProcessRecord,
    parent: StructureNode,
    messageType: MessageType,
  ): StructureNode {
    const children: StructureNode[] = [];
    const routerLink: string = 'vorgang/' + processRecord.recordId;
    const type = parent.type.endsWith('process') ? 'subprocess' : 'process';
    const nodeName = type === 'process' ? 'Vorgang' : 'Teilvorgang';
    const processNode: StructureNode = {
      id: processRecord.recordId,
      title: nodeName + ': ' + processRecord.generalMetadata?.recordNumber,
      subtitle: processRecord.generalMetadata?.subject,
      recordId: processRecord.recordId,
      type: type,
      routerLink: routerLink,
      parentId: parent.id,
      generalMetadata: processRecord.generalMetadata,
      canBeAppraised: this.canBeAppraised(type, parent, messageType),
      canChoosePackaging: this.canChoosePackaging(type, parent),
      selectable: this.isSelectable(type, parent, messageType),
      children: children,
    };
    // generate child nodes for all subprocesses (de: Teilvorgänge)
    if (processRecord.subprocesses) {
      for (let subprocess of processRecord.subprocesses) {
        children.push(this.getProcessStructureNode(subprocess, processNode, messageType));
      }
    }
    // generate child nodes for all documents (de: Dokumente)
    if (processRecord.documents) {
      for (let document of processRecord.documents) {
        children.push(this.getDocumentStructureNode(document, processNode, messageType));
      }
    }

    this.nodes.set(processNode.id, processNode);
    return processNode;
  }

  private getDocumentStructureNode(
    documentRecord: DocumentRecord,
    parent: StructureNode,
    messageType: MessageType,
  ): StructureNode {
    const children: StructureNode[] = [];
    const type =
      parent.type === 'document' || parent.type === 'attachment' ? 'attachment' : 'document';
    const nodeName = type === 'attachment' ? 'Anlage' : 'Dokument';
    const routerLink: string = 'dokument/' + documentRecord.recordId;
    const documentNode: StructureNode = {
      id: documentRecord.recordId,
      title: nodeName + ': ' + documentRecord.generalMetadata?.recordNumber,
      subtitle: documentRecord.generalMetadata?.subject,
      recordId: documentRecord.recordId,
      type: type,
      routerLink: routerLink,
      parentId: parent.id,
      generalMetadata: documentRecord.generalMetadata,
      canBeAppraised: false,
      canChoosePackaging: this.canChoosePackaging(type, parent),
      selectable: this.isSelectable(type, parent, messageType),
      children: children,
    };
    if (documentRecord.attachments) {
      for (let document of documentRecord.attachments) {
        children.push(this.getDocumentStructureNode(document, documentNode, messageType));
      }
    }
    this.nodes.set(documentNode.id, documentNode);
    return documentNode;
  }

  private canBeAppraised(
    type: StructureNodeType,
    parent: StructureNode,
    messageType: MessageType,
  ): boolean {
    if (messageType !== '0501') {
      return false;
    }
    if (type !== 'file' && type !== 'subfile' && type !== 'process' && type !== 'subprocess') {
      return false;
    }
    switch (this.config!.appraisalLevel) {
      case 'root':
        return parent.type === 'message';
      case 'all':
        return true;
      default:
        console.error('called canBeAppraised when config was not ready');
        return false;
    }
  }

  /**
   * Whether the user can choose a packaging level for the node.
   *
   * We currently support packaging only for the root level. Processes cannot be
   * sub-packaged.
   */
  private canChoosePackaging(type: StructureNodeType, parent: StructureNode): boolean {
    return type === 'file' && parent.type === 'message';
  }

  /**
   * Returns whether the node should have a checkbox in multi-selection mode.
   *
   * The result depends on the message type since the action for selected nodes
   * differs depending on the message type.
   */
  private isSelectable(
    type: StructureNodeType,
    parent: StructureNode,
    messageType: MessageType,
  ): boolean {
    if (type === 'message') {
      return true;
    }
    switch (messageType) {
      case '0501':
        // Multi-selection is used for appraisal.
        return this.canBeAppraised(type, parent, messageType);
      case '0503':
        // Multi-selection is used for packaging.
        return this.canChoosePackaging(type, parent);
      default:
        console.error('unexpected message type: ' + type);
        return false;
    }
  }
}
