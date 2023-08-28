// angular
import { Component, Input } from '@angular/core';

// angular material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource, MatTree } from '@angular/material/tree';

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
  treeExtractionInProgress: boolean;

  constructor(private messageService: MessageService) {
    this.treeExtractionInProgress = false;
    this.treeControl = new NestedTreeControl<StructureNode>(
      (node) => node.children
    );
    this.dataSource = new MatTreeNestedDataSource<StructureNode>();
  }

  hasChild = (_: number, node: StructureNode) =>
    !!node.children && node.children.length > 0;

  @Input() set messageText(message: string | undefined) {
    if (message) {
      this.treeExtractionInProgress = true;
      
      this.clearTreeData();
      const treeData: StructureNode[] = [];
      const messageNode = this.getMessageNode(message);
      treeData.push(messageNode);
      this.dataSource.data = treeData;
      this.treeControl.dataNodes = treeData;
      this.treeControl.expand(messageNode);
      this.treeControl.expand(
        messageNode.children!.find((node) => node.type === 'recordObjectList')!
      );
      this.treeExtractionInProgress = false;
    }
  }

  private clearTreeData(): void {
    this.dataSource.data = [];
    this.treeControl.dataNodes = [];
  }

  getMessageNode(message: string) {
    const messageDom = this.messageService.parseMessage(message);
    const messageXmlNode: Node = messageDom.firstChild!;
    const messageHeadNode = this.getMessageHeadNode(messageXmlNode);
    const recordObjectNode = this.getRecordObjectsNode(messageXmlNode);
    const node = this.messageService.addNode(
      'message',
      messageXmlNode,
      [messageHeadNode, recordObjectNode]
    );
    return node;
  }

  getMessageHeadNode(messageXmlNode: Node): StructureNode {
    const messageHeadXmlNodes = this.messageService.getXmlNodes(
      '//xdomea:Kopf',
      messageXmlNode
    );
    const messageHeadXmlNode: Node = messageHeadXmlNodes.snapshotItem(0)!;
    const node = this.messageService.addNode(
      'messageHead',
      messageHeadXmlNode
    );
    return node;
  }

  getRecordObjectsNode(messageXmlNode: Node): StructureNode {
    const node = this.messageService.addNode(
      'recordObjectList',
      messageXmlNode,
      this.getFileObjectNodes(messageXmlNode)
    );
    return node;
  }

  getFileObjectNodes(messageXmlNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const fileXmlNodes = this.messageService.getXmlNodes(
      '//xdomea:Schriftgutobjekt/xdomea:Akte',
      messageXmlNode
    );
    for (let index = 0; index < fileXmlNodes.snapshotLength; ++index) {
      const fileXmlNode: Node = fileXmlNodes.snapshotItem(index)!;
      const node = this.messageService.addNode(
        'file',
        fileXmlNode,
        this.getProcessObjectNodes(fileXmlNode)
      );
      nodes.push(node);
    }
    return nodes;
  }

  getProcessObjectNodes(fileXmlNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const processXmlNodes = this.messageService.getXmlNodes(
      'xdomea:Akteninhalt/xdomea:Vorgang',
      fileXmlNode
    );
    for (let index = 0; index < processXmlNodes.snapshotLength; ++index) {
      const processXmlNode: Node = processXmlNodes.snapshotItem(index)!;
      const node = this.messageService.addNode(
        'process',
        processXmlNode,
        this.getDocumentObjectNodes(processXmlNode)
      );
      nodes.push(node);
    }
    return nodes;
  }

  getDocumentObjectNodes(processXmlNode: Node): StructureNode[] {
    const nodes: StructureNode[] = [];
    const documentXmlNodes = this.messageService.getXmlNodes(
      'xdomea:Dokument',
      processXmlNode
    );
    for (let index = 0; index < documentXmlNodes.snapshotLength; ++index) {
      const documentXmlNode: Node = documentXmlNodes.snapshotItem(index)!;
      const node = this.messageService.addNode(
        'document',
        documentXmlNode
      );
      nodes.push(node);
    }
    return nodes;
  }
}