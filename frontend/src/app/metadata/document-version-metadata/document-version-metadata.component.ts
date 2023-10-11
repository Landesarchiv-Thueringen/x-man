// angular
import { Component, Input } from '@angular/core';

// material
import { FlatTreeControl } from '@angular/cdk/tree';
import {
  MatTreeFlatDataSource,
  MatTreeFlattener,
} from '@angular/material/tree';

// project
import { DocumentRecordObject } from 'src/app/message/message.service';

export type NodeType = 'version' | 'format';

export interface Node {
  text: string;
  type: NodeType;
  fileID?: number;
  children?: Node[];
}

export interface FlatNode {
  expandable: boolean;
  level: number;
  text: string;
  fileID?: number;
  type: NodeType;
}

@Component({
  selector: 'app-document-version-metadata',
  templateUrl: './document-version-metadata.component.html',
  styleUrls: ['./document-version-metadata.component.scss'],
})
export class DocumentVersionMetadataComponent {
  treeControl: FlatTreeControl<FlatNode>;
  treeFlattener: MatTreeFlattener<Node, FlatNode>;
  dataSource: MatTreeFlatDataSource<Node, FlatNode>;

  constructor() {
    this.treeControl = new FlatTreeControl<FlatNode>(
      (node) => node.level,
      (node) => node.expandable
    );
    this.treeFlattener = new MatTreeFlattener(
      this._transformer,
      (node) => node.level,
      (node) => node.expandable,
      (node) => node.children
    );
    this.dataSource = new MatTreeFlatDataSource(
      this.treeControl,
      this.treeFlattener
    );
  }

  private _transformer = (node: Node, level: number): FlatNode => {
    return {
      expandable: !!node.children && node.children.length > 0,
      level: level,
      text: node.text,
      type: node.type,
      fileID: node.fileID,
    };
  };

  hasChild = (_: number, node: FlatNode) => node.expandable;

  d?: DocumentRecordObject;
  @Input() set document(d: DocumentRecordObject | null | undefined) {
    if (!!d) {
      this.d = d;
      this.initTree();
    }
  }

  initTree(): void {
    if (!!this.d && !!this.d.versions) {
      const treeData: Node[] = [];
      for (let version of this.d.versions) {
        const formatNodes: Node[] = [];
        for (let format of version.formats) {
          const formatNode: Node = {
            text: format.primaryDocument.fileNameOriginal
              ? format.primaryDocument.fileNameOriginal
              : format.primaryDocument.fileName,
            type: 'format',
            fileID: format.primaryDocument.id,
          };
          formatNodes.push(formatNode);
        }
        const versionNode: Node = {
          text: 'Version ' + version.versionID,
          type: 'version',
          children: formatNodes,
        };
        treeData.push(versionNode);
      }
      this.dataSource.data = treeData;
      this.treeControl.expandAll();
    }
  }

  downloadPrimaryFile(fileID: number): void {
    console.log(fileID);
  }
}
