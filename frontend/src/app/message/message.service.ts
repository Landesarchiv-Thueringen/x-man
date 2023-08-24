// angular
import { Injectable } from '@angular/core';

// utility
import { v4 as uuidv4 } from 'uuid';

type StructureNodeType =
  | 'message'
  | 'messageHead'
  | 'recordObjectList'
  | 'file'
  | 'process'
  | 'document';

export interface StructureNode {
  displayText: string;
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

  constructor() {
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
    displayText: string,
    type: StructureNodeType,
    xmlNode: Node,
    children?: StructureNode[]
  ): StructureNode {
    const nodeId = this.getNodeId(type, xmlNode);
    const node: StructureNode = {
      displayText: displayText,
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

  getXmlNodes(xpath: string, node?: Node): XPathResult {
    if (!this.messageDom) {
      throw new Error('message dom not initialized');
    }
    return this.messageDom!.evaluate(
      xpath,
      node ? node : this.messageDom,
      (namespace) => {
        return 'urn:xoev-de:xdomea:schema:2.3.0';
      },
      XPathResult.ORDERED_NODE_SNAPSHOT_TYPE,
      null
    );
  }
}
