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
      const messageNode = this.getMessageNode(message);
      treeData.push(messageNode);
      this.dataSource.data = treeData;
    }
  }

  getMessageNode(message: string) {
    const parser = new DOMParser();
    const doc: Document = parser.parseFromString(message, 'application/xml');
    const messageXmlNode: Node = doc.firstChild!;
    const messageHeadNode = this.getMessageHeadNode(doc);
    const recordObjectNode = this.getRecordObjectsNode(doc, messageXmlNode);
    const node = this.messageService.addNode(
      'Anbietungsverzeichnis',
      'message',
      messageXmlNode,
      [messageHeadNode, recordObjectNode]
    );
    return node;
  }

  getMessageHeadNode(doc: Document): StructureNode {
    const messageHeadXmlNodes = this.getXmlNodes(doc, '//xdomea:Kopf');
    if (messageHeadXmlNodes.snapshotLength !== 1) {
      console.error('alarm');
    }
    const messageHeadXmlNode: Node = messageHeadXmlNodes.snapshotItem(0)!;
    const node = this.messageService.addNode(
      'Nachrichtenkopf',
      'messageHead',
      messageHeadXmlNode,
    )
    return node;
  }

  getRecordObjectsNode(doc: Document, messageXmlNode: Node): StructureNode {
    const node = this.messageService.addNode(
      'Schriftgutobjekte',
      'recordObjectList',
      messageXmlNode,
      this.getFileObjectNodes(doc),
    )
    return node;
  }

  getFileObjectNodes(doc: Document): StructureNode[] {
    const nodes: StructureNode[] = [];
    const fileXmlNodes = this.getXmlNodes(
      doc,
      '//xdomea:Schriftgutobjekt/xdomea:Akte'
    );
    for (let index = 0; index < fileXmlNodes.snapshotLength; ++index) {
      const fileXmlNode: Node = fileXmlNodes.snapshotItem(index)!;
      const recordNumberXmlNode = this.getXmlNodes(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        fileXmlNode
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Akte: ' + recordNumberXmlNode!.textContent,
        'file',
        fileXmlNode,
        this.getProcessObjectNodes(doc, fileXmlNode),
      )
      nodes.push(node);
    }
    return nodes;
  }

  getProcessObjectNodes(doc: Document, fileXmlNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const processXmlNodes = this.getXmlNodes(
      doc,
      'xdomea:Akteninhalt/xdomea:Vorgang',
      fileXmlNode
    );
    for (let index = 0; index < processXmlNodes.snapshotLength; ++index) {
      const processXmlNode: Node = processXmlNodes.snapshotItem(index)!;
      const recordNumberXmlNode = this.getXmlNodes(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        processXmlNode
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Vorgang: ' + recordNumberXmlNode!.textContent,
        'process',
        processXmlNode,
        this.getDocumentObjectNodes(doc, processXmlNode),
      )
      nodes.push(node);
    }
    return nodes;
  }

  getDocumentObjectNodes(doc: Document, processXmlNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const documentXmlNodes = this.getXmlNodes(
      doc,
      'xdomea:Dokument',
      processXmlNode
    );
    for (let index = 0; index < documentXmlNodes.snapshotLength; ++index) {
      const documentXmlNode: Node = documentXmlNodes.snapshotItem(index)!;
      const recordNumberXmlNode = this.getXmlNodes(
        doc,
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        documentXmlNode
      ).snapshotItem(0);
      const node = this.messageService.addNode(
        'Dokument: ' + recordNumberXmlNode!.textContent,
        'document',
        documentXmlNode,
      )
      nodes.push(node);
    }
    return nodes;
  }

  getXmlNodes(doc: Document, xpath: string, node?: Node): XPathResult {
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
