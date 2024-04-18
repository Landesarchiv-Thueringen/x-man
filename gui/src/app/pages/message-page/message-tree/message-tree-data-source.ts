import { CollectionViewer, DataSource } from '@angular/cdk/collections';
import { FlatTreeControl } from '@angular/cdk/tree';
import { MatTreeFlatDataSource, MatTreeFlattener } from '@angular/material/tree';
import { Observable } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import { StructureNode, StructureNodeType } from '../message-processor.service';

const GROUP_SIZE = 100;

export type GroupedStructureNodeType = StructureNodeType | 'file-group' | 'process-group' | 'document-group';

interface GroupedStructureNode extends Omit<StructureNode, 'type' | 'children'> {
  type: GroupedStructureNodeType;
  children?: GroupedStructureNode[];
}

export interface FlatNode extends Omit<GroupedStructureNode, 'children'> {
  expandable: boolean;
  level: number;
}

export class MessageTreeDataSource extends DataSource<FlatNode> {
  private readonly transformer = (node: GroupedStructureNode, level: number): FlatNode => {
    const { children, ...baseNode } = node;
    return {
      ...baseNode,
      expandable: children != null && children.length > 0,
      level,
    };
  };

  private readonly treeFlattener = new MatTreeFlattener(
    this.transformer,
    (node) => node.level,
    (node) => node.expandable,
    (node) => node.children,
  );

  private readonly flatTreeDataSource = new MatTreeFlatDataSource(this.treeControl, this.treeFlattener);
  private nodesMap = new Map<string, GroupedStructureNode>();

  private _data?: StructureNode;
  /** The original tree as obtained from MessageProcessorService. */
  set data(data: StructureNode) {
    this._data = data;
    let flatTreeData = this.groupNodes(data);
    this.updateFlatTreeData(flatTreeData);
  }

  constructor(private treeControl: FlatTreeControl<FlatNode>) {
    super();
  }

  connect(collectionViewer: CollectionViewer): Observable<readonly FlatNode[]> {
    return this.flatTreeDataSource.connect(collectionViewer);
  }

  disconnect() {
    return this.flatTreeDataSource.disconnect();
  }

  getNode(id: string): GroupedStructureNode {
    const node = this.nodesMap.get(id);
    if (node == null) {
      throw new Error('node not found: ' + id);
    }
    return node;
  }

  /**
   * Updates the structure nodes, that are the base for this data source's data.
   *
   * We feed structure nodes as input to `flatTreeDataSource`. We update
   * structure nodes when the source data is updated with `init` and when
   * filters change.
   */
  private updateFlatTreeData(data: GroupedStructureNode): void {
    this.flatTreeDataSource.data = [data];
    this.nodesMap.clear();
    this.addToNodesMap(data);
  }

  private addToNodesMap(node: GroupedStructureNode): void {
    this.nodesMap.set(node.id, node);
    if (node.children) {
      for (const child of node.children) {
        this.addToNodesMap(child);
      }
    }
  }

  private groupNodes(data: StructureNode): GroupedStructureNode {
    const { children, ...node } = data;
    const shouldGroupChildren = (children?.length ?? 0) > GROUP_SIZE;
    if (shouldGroupChildren) {
      let groupedChildren: GroupedStructureNode[] = [];
      for (const type of ['file', 'subfile', 'process', 'subprocess', 'document', 'attachment'] as const) {
        groupedChildren = [...groupedChildren, ...this.getGroups(children!, type)];
      }
      return { ...node, children: groupedChildren };
    } else {
      return { ...node, children: children?.map((child) => this.groupNodes(child)) };
    }
  }

  private getGroups(nodes: StructureNode[], type: StructureNode['type']): GroupedStructureNode[] {
    const relevantNodes = nodes.filter((node) => node.type === type);
    let currentGroup: GroupedStructureNode;
    let name: string;
    let groupType: GroupedStructureNodeType;
    const result: GroupedStructureNode[] = [];
    switch (type) {
      case 'file':
        name = 'Akten';
        groupType = 'file-group';
        break;
      case 'subfile':
        name = 'Teilakten';
        groupType = 'file-group';
        break;
      case 'process':
        name = 'Vorgänge';
        groupType = 'process-group';
        break;
      case 'subprocess':
        name = 'Teilvorgänge';
        groupType = 'process-group';
        break;
      case 'document':
        name = 'Dokumente';
        groupType = 'document-group';
        break;
      case 'attachment':
        name = 'Anhänge';
        groupType = 'document-group';
        break;
      default:
        throw new Error('unhandled type: ' + type);
    }
    for (const [index, node] of relevantNodes.entries()) {
      if (index % GROUP_SIZE === 0) {
        currentGroup = {
          id: uuidv4(),
          title: `${name} ${index + 1}...${Math.min(index + GROUP_SIZE, relevantNodes.length)}`,
          type: groupType,
          parentID: node.parentID,
          children: [],
          canBeAppraised: relevantNodes.some((n) => n.canBeAppraised),
        };
        result.push(currentGroup);
      }
      currentGroup!.children!.push({
        ...node,
        parentID: currentGroup!.id,
        children: node.children?.map((child) => this.groupNodes(child)),
      });
    }
    return result;
  }
}
