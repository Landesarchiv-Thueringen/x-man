// angular
import { Injectable } from '@angular/core';
import { DatePipe } from '@angular/common';

// utility
import { v4 as uuidv4 } from 'uuid';

type StructureNodeType =
  | 'message'
  | 'messageHead'
  | 'recordObjectList'
  | 'file'
  | 'process'
  | 'document';

export interface DisplayText {
  title: string;
  subtitle?: string;
}

export interface StructureNode {
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  xmlNode: Node;
  children?: StructureNode[];
}

@Injectable({
  providedIn: 'root',
})
export class MessageService {
  parser: DOMParser;
  messageDom?: Document;
  nodes: Map<string, StructureNode>;

  constructor(private datePipe: DatePipe,) {
    this.parser = new DOMParser();
    this.nodes = new Map<string, StructureNode>();
  }

  parseMessage(message: string): Document {
    this.messageDom = this.parser.parseFromString(message, 'application/xml');
    return this.messageDom;
  }

  getMessageDom(): Document {
    if (!this.messageDom) {
      throw new Error('message dom not initialized');
    }
    return this.messageDom;
  }

  /**
   * Structure node is stored in map service and tree component for fast access. There are no
   * storage concerns because in the node in the component is a shallow copy.
   */
  addNode(
    type: StructureNodeType,
    xmlNode: Node,
    children?: StructureNode[]
  ): StructureNode {
    const nodeId = this.getNodeId(type, xmlNode);
    const node: StructureNode = {
      displayText: this.getDisplayText(type, xmlNode),
      type: type,
      xmlNode: xmlNode,
      children: children,
      routerLink: this.getRouterLink(type, nodeId),
    };

    this.nodes.set(nodeId, node);
    return node;
  }

  getNode(id: string): StructureNode | undefined {
    return this.nodes.get(id);
  }

  getNodeId(type: StructureNodeType, xmlNode: Node): string {
    if (type === 'file' || type === 'process' || type === 'document') {
      const idXmlNode: Node = this.getXmlNodes(
        'xdomea:Identifikation/xdomea:ID',
        xmlNode
      ).snapshotItem(0)!;
      return idXmlNode.textContent!;
    }
    return uuidv4();
  }

  private getDisplayText(type: StructureNodeType, xmlNode: Node): DisplayText {
    switch (type) {
      case 'message':
        return { title: 'Anbietungsverzeichnis' };
      case 'recordObjectList':
        return { title: 'Schriftgutobjekte' };
      case 'messageHead':
        return { title: 'Nachrichtenkopf' };
      case 'file':
        return this.getRecorcObjectDisplayText(type, xmlNode);
      case 'process':
        return this.getRecorcObjectDisplayText(type, xmlNode);
      case 'document':
        return this.getRecorcObjectDisplayText(type, xmlNode);
    }
  }

  private getRecorcObjectDisplayText(
    type: StructureNodeType,
    xmlNode: Node
  ): DisplayText {
    if (type === 'file' || type === 'process' || type === 'document') {
      const recordNumberXmlNode: Node = this.getXmlNodes(
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        xmlNode
      ).snapshotItem(0)!;
      const subjectXmlNode: Node | null = this.getXmlNodes(
        'xdomea:AllgemeineMetadaten/xdomea:Betreff',
        xmlNode
      ).snapshotItem(0);
      const subtitle: string | undefined = subjectXmlNode?.textContent
        ? subjectXmlNode.textContent
        : undefined;
      let recordObjectTitle = recordNumberXmlNode.textContent;
      switch (type) {
        case 'file':
          return { title: 'Akte: ' + recordObjectTitle, subtitle: subtitle };
        case 'process':
          return { title: 'Vorgang: ' + recordObjectTitle, subtitle: subtitle };
        case 'document':
          return { title: 'Dokument: ' + recordObjectTitle, subtitle: subtitle };
      }
    }
    throw new Error('no record object');
  }

  private getRouterLink(nodeType: StructureNodeType, nodeId: string): string {
    switch (nodeType) {
      case 'message':
        return 'nachricht/' + nodeId;
      case 'recordObjectList':
        return 'schriftgutobjekte' + nodeId;
      case 'messageHead':
        return 'nachrichtenkopf/' + nodeId;
      case 'file':
        return 'akte/' + nodeId;
      case 'process':
        return 'vorgang/' + nodeId;
      case 'document':
        return 'dokument/' + nodeId;
    }
  }

  getXmlNodes(xpath: string, xmlNode?: Node): XPathResult {
    if (!this.messageDom) {
      throw new Error('message dom not initialized');
    }
    return this.messageDom!.evaluate(
      xpath,
      xmlNode ? xmlNode : this.messageDom,
      (namespace) => {
        return 'urn:xoev-de:xdomea:schema:2.3.0';
      },
      XPathResult.ORDERED_NODE_SNAPSHOT_TYPE,
      null
    );
  }

  /** 
   * Returns null if the xml node or its text contents are null, because that means the date was not
   * provided in the message. Returns the text content of the xml node if the text content is no 
   * parsable date to show the malformed date in the ui. Returns formatted date string if text 
   * content is parsable.
   */
  getDateText(xmlNode: Node|null): string|null {
    if (xmlNode?.textContent) {
      const timestamp: number = Date.parse(xmlNode?.textContent);
      if (Number.isNaN(timestamp)) {
        return xmlNode?.textContent;
      } else {
        const date: Date = new Date(timestamp);
        return this.datePipe.transform(date);
      }
    }
    return null;
  }
}
