import { FlatTreeControl } from '@angular/cdk/tree';
import { CommonModule } from '@angular/common';
import { Component, effect, input, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatTreeFlatDataSource, MatTreeFlattener, MatTreeModule } from '@angular/material/tree';
import { MessageService } from '../../../../services/message.service';
import { DocumentRecord } from '../../../../services/records.service';
import { MessagePageService } from '../../message-page.service';

export type NodeType = 'version' | 'format';

export interface Node {
  text: string;
  type: NodeType;
  filename?: string;
  children?: Node[];
}

export interface FlatNode {
  expandable: boolean;
  level: number;
  text: string;
  filename?: string;
  type: NodeType;
}

@Component({
  selector: 'app-document-version-metadata',
  templateUrl: './document-version-metadata.component.html',
  styleUrls: ['./document-version-metadata.component.scss'],
  imports: [CommonModule, MatButtonModule, MatExpansionModule, MatTreeModule, MatIconModule],
})
export class DocumentVersionMetadataComponent {
  private messageService = inject(MessageService);
  private messagePageService = inject(MessagePageService);

  readonly document = input<DocumentRecord>();

  treeControl: FlatTreeControl<FlatNode>;
  treeFlattener: MatTreeFlattener<Node, FlatNode>;
  dataSource: MatTreeFlatDataSource<Node, FlatNode>;
  message = this.messagePageService.message;

  constructor() {
    this.treeControl = new FlatTreeControl<FlatNode>(
      (node) => node.level,
      (node) => node.expandable,
    );
    this.treeFlattener = new MatTreeFlattener(
      this._transformer,
      (node) => node.level,
      (node) => node.expandable,
      (node) => node.children,
    );
    this.dataSource = new MatTreeFlatDataSource(this.treeControl, this.treeFlattener);
    effect(() => this.initTree(this.document()));
  }

  private _transformer = (node: Node, level: number): FlatNode => {
    return {
      expandable: !!node.children && node.children.length > 0,
      level: level,
      text: node.text,
      type: node.type,
      filename: node.filename,
    };
  };

  hasChild = (_: number, node: FlatNode) => node.expandable;

  initTree(document?: DocumentRecord): void {
    if (document && document.versions) {
      const treeData: Node[] = [];
      for (let version of document.versions) {
        const formatNodes: Node[] = [];
        for (let format of version.formats) {
          const formatNode: Node = {
            text: format.primaryDocument.filenameOriginal
              ? format.primaryDocument.filenameOriginal
              : format.primaryDocument.filename,
            type: 'format',
            filename: format.primaryDocument.filenameOriginal
              ? format.primaryDocument.filenameOriginal
              : format.primaryDocument.filename,
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

  downloadPrimaryFile(node: FlatNode): void {
    if (this.document()) {
      this.messageService
        .getPrimaryDocument(this.message()!.messageHead.processID, node.filename!)
        .subscribe((file) => {
          const a = document.createElement('a');
          document.body.appendChild(a);
          a.download = node.filename!;
          a.href = window.URL.createObjectURL(file);
          a.click();
          document.body.removeChild(a);
        });
    }
  }
}
