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
  nodes: Map<string, StructureNode>;

  constructor() {
    this.nodes = new Map<string, StructureNode>();
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
    const nodeId = uuidv4();
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

  private getRouterLink(nodeType: StructureNodeType, nodeId: string): string {
    let routerLink = '';
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
    return routerLink;
  }
}
