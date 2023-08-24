// angular
import { Component, Input } from '@angular/core';

// angular material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import { MessageService, StructureNode } from '../message/message.service';

@Component({
  selector: 'app-message-tree',
  templateUrl: './message-tree.component.html',
  styleUrls: ['./message-tree.component.scss'],
})
export class MessageTreeComponent {
  treeControl: NestedTreeControl<StructureNode>;
  dataSource: MatTreeNestedDataSource<StructureNode>;

  constructor(private messageService: MessageService) {
    this.treeControl = new NestedTreeControl<StructureNode>(
      (node) => node.children
    );
    this.dataSource = new MatTreeNestedDataSource<StructureNode>();
  }

  hasChild = (_: number, node: StructureNode) =>
    !!node.children && node.children.length > 0;

  @Input() set messageText(message: string | undefined) {
    if (message) {
      console.log(message);
      const treeData: StructureNode[] = [];
      const headNode: StructureNode = this.messageService.addNode(
        'Nachrichtenkopf',
        'messageHead',
      );
      const recordObjectListNode: StructureNode = this.messageService.addNode(
        'Schriftgutobjekte',
        'recordObject',
        this.getRecordObjectNodes(message)
      );
      const messageNode: StructureNode = this.messageService.addNode(
        'Anbietungsverzeichnis',
        'message',
        [headNode, recordObjectListNode]
      );
      treeData.push(messageNode);
      this.dataSource.data = treeData;
    }
  }

  getRecordObjectNodes(message: string): StructureNode[] {
    const parser = new DOMParser();
    const doc: Document = parser.parseFromString(message, 'application/xml');
    const recordObjectNodes: StructureNode[] = this.getFileObjectNodes(doc);
    return recordObjectNodes;
  }

  getFileObjectNodes(doc: Document): StructureNode[] {
    const nodes: StructureNode[] = [];
    const fileObjects = this.getObjects(
      doc,
      '//xdomea:Schriftgutobjekt/xdomea:Akte'
    );
    for (let index = 0; index < fileObjects.snapshotLength; ++index) {
      const fileEl: Node = fileObjects.snapshotItem(index)!;
      const recordNumberEl = this.getObjects(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        fileEl
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Akte: ' + recordNumberEl!.textContent,
        'file',
        this.getProcessObjectNodes(doc, fileEl),
      )
      nodes.push(node);
    }
    return nodes;
  }

  getProcessObjectNodes(doc: Document, fileNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const processObjects = this.getObjects(
      doc,
      'xdomea:Akteninhalt/xdomea:Vorgang',
      fileNode
    );
    for (let index = 0; index < processObjects.snapshotLength; ++index) {
      const processEl: Node = processObjects.snapshotItem(index)!;
      const recordNumberEl = this.getObjects(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        processEl
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Vorgang: ' + recordNumberEl!.textContent,
        'process',
        this.getDocumentObjectNodes(doc, processEl),
      )
      nodes.push(node);
    }
    return nodes;
  }

  getDocumentObjectNodes(doc: Document, processNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const documentObjects = this.getObjects(
      doc,
      'xdomea:Dokument',
      processNode
    );
    for (let index = 0; index < documentObjects.snapshotLength; ++index) {
      const documentEl: Node = documentObjects.snapshotItem(index)!;
      const recordNumberEl = this.getObjects(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        documentEl
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Dokument: ' + recordNumberEl!.textContent,
        'document',
      )
      nodes.push(node);
    }
    return nodes;
  }

  getObjects(doc: Document, xpath: string, node?: Node): XPathResult {
    return doc.evaluate(
      xpath,
      node ? node : doc,
      (namespace) => {
        return 'urn:xoev-de:xdomea:schema:2.3.0';
      },
      XPathResult.ORDERED_NODE_SNAPSHOT_TYPE,
      null
    );
  }
}
