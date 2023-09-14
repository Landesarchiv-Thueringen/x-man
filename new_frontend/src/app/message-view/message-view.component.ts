// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import { DocumentRecordObject, FileRecordObject, Message, MessageService, ProcessRecordObject } from '../message/message.service';

// utility
import { Subscription } from 'rxjs';

type StructureNodeType =
  | 'message'
  | 'file'
  | 'process'
  | 'document';

export interface DisplayText {
  title: string;
  subtitle?: string;
}

export interface StructureNode {
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  children?: StructureNode[];
}

@Component({
  selector: 'app-message-view',
  templateUrl: './message-view.component.html',
  styleUrls: ['./message-view.component.scss']
})
export class MessageViewComponent implements AfterViewInit, OnDestroy {
  treeControl: NestedTreeControl<StructureNode>;
  dataSource: MatTreeNestedDataSource<StructureNode>;
  urlParameterSubscription?: Subscription;

  constructor(
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    this.treeControl = new NestedTreeControl<StructureNode>(
      (node) => node.children
    );
    this.dataSource = new MatTreeNestedDataSource<StructureNode>();
  }

  hasChild = (_: number, node: StructureNode) =>
    !!node.children && node.children.length > 0;

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.subscribe((params) => {
      this.messageService.getMessage(+params['id']).subscribe(
        (message: Message) => {
          const treeData: StructureNode[] = [];
          const messageNode = this.processMessage(message);
          treeData.push(messageNode);
          this.dataSource.data = treeData;
          this.treeControl.dataNodes = treeData;
          this.treeControl.expand(messageNode);
        }
      )
    })
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }

  processMessage(message: Message): StructureNode {
    const children: StructureNode[] = [];
    for (let recordObject of message.recordObjects) {
      if (recordObject.fileRecordObject) {
        children.push(this.getFileStructureNode(recordObject.fileRecordObject));
      }
    }
    let displayText: DisplayText;
    switch (message.messageType.code) {
      case '0501':
        displayText = {
          title: 'Anbietung',
        };
        break;
      case '0503':
        displayText = {
          title: 'Abgabe',
        };
        break;
      default:
        throw new Error('unhandled message type');
    }
    const routerLink: string = 'details';
    const messageNode: StructureNode = {
      displayText: displayText,
      type: 'message',
      routerLink: routerLink,
      children: children,
    }
    return messageNode;
  }

  getFileStructureNode(fileRecordObject: FileRecordObject): StructureNode {
    const children: StructureNode[] = [];
    for (let process of fileRecordObject.processes) {
      children.push(this.getProcessStructureNode(process));
    }
    const displayText: DisplayText = {
      title: 'Akte: ' + fileRecordObject.generalMetadata?.xdomeaID,
      subtitle: fileRecordObject.generalMetadata?.subject,
    }
    const routerLink: string = 'akte/' + fileRecordObject.id;
    const node: StructureNode = {
      displayText: displayText,
      type: 'file',
      routerLink: routerLink,
      children: children,
    };
    return node;
  }

  getProcessStructureNode(processRecordObject: ProcessRecordObject): StructureNode {
    const children: StructureNode[] = [];
    for (let document of processRecordObject.documents) {
      children.push(this.getDocumentStructureNode(document));
    }
    const displayText: DisplayText = {
      title: 'Vorgang: ' + processRecordObject.generalMetadata?.xdomeaID,
      subtitle: processRecordObject.generalMetadata?.subject,
    }
    const routerLink: string = 'vorgang/' + processRecordObject.id;
    const node: StructureNode = {
      displayText: displayText,
      type: 'process',
      routerLink: routerLink,
      children: children,
    };
    return node;
  }

  getDocumentStructureNode(documentRecordObject: DocumentRecordObject): StructureNode {
    const displayText: DisplayText = {
      title: 'Dokument: ' + documentRecordObject.generalMetadata?.xdomeaID,
      subtitle: documentRecordObject.generalMetadata?.subject,
    }
    const routerLink: string = 'dokument/' + documentRecordObject.id;
    const node: StructureNode = {
      displayText: displayText,
      type: 'document',
      routerLink: routerLink,
    };
    return node;
  }
}
