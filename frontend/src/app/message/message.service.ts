// angular
import { Injectable } from '@angular/core';

// utility
import { v4 as uuidv4 } from 'uuid';

type StructureNodeType =
  | 'message'
  | 'messageHead'
  | 'recordObject'
  | 'file'
  | 'process'
  | 'document';

export interface StructureNode {
  displayText: string;
  type: StructureNodeType;
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
    children?: StructureNode[]
  ): StructureNode {
    const node: StructureNode = {
      displayText: displayText,
      type: type,
      children: children,
    };
    const nodeId = uuidv4();
    this.nodes.set(nodeId, node);
    return node;
  }
}
