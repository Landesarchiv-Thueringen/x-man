import { CollectionViewer, DataSource } from '@angular/cdk/collections';
import { FlatTreeControl } from '@angular/cdk/tree';
import { MatTreeFlatDataSource, MatTreeFlattener } from '@angular/material/tree';
import { BehaviorSubject, Observable } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import { notNull } from '../../../utils/predicates';
import { StructureNode, StructureNodeType } from '../message-processor';

const GROUP_SIZE = 100;

export type GroupedStructureNodeType =
  | StructureNodeType
  | 'filtered'
  | 'file-group'
  | 'process-group'
  | 'document-group';

interface GroupedStructureNode extends Omit<StructureNode, 'type' | 'children'> {
  type: GroupedStructureNodeType;
  children?: GroupedStructureNode[];
  /** Whether to show a checkbox for the item when selection is active. */
  selectable: boolean;
}

export interface FlatNode extends Omit<GroupedStructureNode, 'children'> {
  expandable: boolean;
  level: number;
}

export type FilterPredicate = (node: StructureNode) => FilterResult;

/**
 * Values have the following meaning:
 * - 'show': The node will be included. Its children will be tested separately.
 * - 'show-recursive': The node and all its children will be included.
 * - 'hide': The node will be omitted by default, but its children will be
 *   tested separately and if any child is included, the node will also be
 *   included.
 * - 'hide-recursive': The node and all its children will be omitted.
 * - 'propagate-recursive': The node will assume the filter result of its parent
 *   and apply it to all of its children.
 */
export type FilterResult =
  | 'show'
  | 'show-recursive'
  | 'hide'
  | 'hide-recursive'
  | 'propagate-recursive';

/**
 * MessageTreeDataSource has three data layers
 * - `data` is the source data, that is passed into MessageTreeDataSource.
 * - `displayData` is the filtered and grouped data, which is always updated
 *   when `data` changes, but can also change while `data` remains unchanged.
 * - `flattenedData`, which is handled by `flatTreeDataSource`. `flattenedData`
 *   is always updated when `displayData` changes and additionally when nodes
 *   are expanded / collapsed. `connect()` returns an Observable for
 *   `flattenedData`.
 */
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

  private readonly displayData = new BehaviorSubject<GroupedStructureNode | null>(null);
  private readonly flatTreeDataSource = new MatTreeFlatDataSource(
    this.treeControl,
    this.treeFlattener,
  );
  private nodesMap = new Map<string, GroupedStructureNode>();

  private _data?: StructureNode;
  /** The original tree as obtained from MessageProcessorService. */
  set data(data: StructureNode | undefined) {
    this._data = data;
    this.updateDisplayData();
  }
  get data() {
    return this._data;
  }

  private _filters?: FilterPredicate[];
  set filters(filters: FilterPredicate[]) {
    this._filters = filters;
    this.updateDisplayData();
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

  observeDisplayData() {
    return this.displayData.asObservable();
  }

  /**
   * Updates the structure nodes, that are the base for this data source's data.
   *
   * We feed structure nodes as input to `flatTreeDataSource`. We update
   * structure nodes when the source data is updated with `init` and when
   * filters change.
   */
  private updateDisplayData(): void {
    if (!this._data) {
      return;
    }
    let nodes = this._data;
    nodes = this.filterNodes(nodes);
    let groupedNodes = this.groupNodes(nodes);
    this.nodesMap.clear();
    this.addToNodesMap(groupedNodes);
    this.flatTreeDataSource.data = [groupedNodes];
    this.displayData.next(groupedNodes);
  }

  private addToNodesMap(node: GroupedStructureNode): void {
    this.nodesMap.set(node.id, node);
    if (node.children) {
      for (const child of node.children) {
        this.addToNodesMap(child);
      }
    }
  }

  private filterNodes(rootNode: StructureNode): StructureNode {
    if (!this._filters?.length) {
      return rootNode;
    } else {
      const children = this.getFilteredChildren(
        rootNode,
        Array(this._filters!.length).fill('show'),
      );
      return { ...rootNode, children };
    }
  }

  private filterNode(node: StructureNode, parentResults: FilterResult[]): StructureNode | null {
    const filterResults = this.getFilterResultsForNode(node, parentResults);
    if (filterResults.every((r) => r === 'show-recursive')) {
      return node;
    } else if (filterResults.some((r) => r === 'hide-recursive')) {
      return null;
    } else if (filterResults.some((r) => r === 'hide') && !node.children) {
      return null;
    } else {
      const children = this.getFilteredChildren(node, filterResults);
      const hasRealChildren = children?.some(
        (child) => child.type !== ('filtered' as StructureNodeType),
      );
      if (hasRealChildren || filterResults.every((r) => r === 'show')) {
        return { ...node, children };
      } else {
        return null;
      }
    }
  }

  private getFilteredChildren(
    node: StructureNode,
    filterResults: FilterResult[],
  ): StructureNode[] | undefined {
    const children = node.children?.map((child) => this.filterNode(child, filterResults));
    const filteredChildren = children?.filter(notNull);
    const numberFiltered = children?.filter((child) => child == null).length ?? 0;
    if (numberFiltered > 0) {
      if (node.type === 'message' && filteredChildren!.length === 0) {
        filteredChildren!.push({
          id: uuidv4(),
          title: 'Keine übereinstimmenden Elemente',
          canBeAppraised: false,
          canChoosePackaging: false,
          selectable: false,
          // Not a valid type for StructureNode, but will be valid for GroupedStructureNode
          type: 'filtered' as StructureNodeType,
        });
      } else {
        filteredChildren!.push({
          id: uuidv4(),
          title: `${numberFiltered} ${numberFiltered > 1 ? 'Elemente' : 'Element'} gefiltert`,
          canBeAppraised: false,
          canChoosePackaging: false,
          selectable: false,
          type: 'filtered' as StructureNodeType,
        });
      }
    }
    return filteredChildren;
  }

  private getFilterResultsForNode(
    node: StructureNode,
    parentResults: FilterResult[],
  ): FilterResult[] {
    const results = Array<FilterResult>(parentResults.length);
    for (const [i, parentResult] of parentResults.entries()) {
      switch (parentResult) {
        case 'show-recursive':
        case 'hide-recursive':
          results[i] = parentResult;
          break;
        case 'show':
        case 'hide':
          const result = this._filters![i](node);
          if (result === 'propagate-recursive') {
            results[i] = (parentResult + '-recursive') as FilterResult;
          } else {
            results[i] = result;
          }
          break;
        default:
          throw new Error('unhandled filter result: ' + parentResult);
      }
    }
    return results;
  }

  private groupNodes(nodes: StructureNode): GroupedStructureNode {
    const { children, ...node } = nodes;
    const selectable = node.selectable;
    const shouldGroupChildren = (children?.length ?? 0) > GROUP_SIZE;
    if (shouldGroupChildren) {
      let groupedChildren: GroupedStructureNode[] = [];
      for (const type of [
        'file',
        'subfile',
        'process',
        'subprocess',
        'document',
        'attachment',
      ] as const) {
        groupedChildren = [...groupedChildren, ...this.getGroups(children!, type)];
      }
      return { ...node, selectable, children: groupedChildren };
    } else {
      return { ...node, selectable, children: children?.map((child) => this.groupNodes(child)) };
    }
  }

  private getGroups(nodes: StructureNode[], type: StructureNodeType): GroupedStructureNode[] {
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
          parentId: node.parentId,
          children: [],
          canBeAppraised: false,
          canChoosePackaging: false,
          selectable: relevantNodes.some((n) => n.selectable),
        };
        result.push(currentGroup);
      }
      currentGroup!.children!.push({
        ...node,
        parentId: currentGroup!.id,
        children: node.children?.map((child) => this.groupNodes(child)),
      });
    }
    return result;
  }
}
