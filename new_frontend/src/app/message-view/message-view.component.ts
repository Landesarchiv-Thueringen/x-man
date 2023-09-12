// angular
import { AfterViewInit, Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import { FileRecordObject, Message, MessageService } from '../message/message.service';

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
export class MessageViewComponent implements AfterViewInit{
  treeControl: NestedTreeControl<StructureNode>;
  dataSource: MatTreeNestedDataSource<StructureNode>;

  constructor(
    private route: ActivatedRoute,
    private messageService: MessageService,
  ) {
    this.treeControl = new NestedTreeControl<StructureNode>(
      (node) => node.children
    );
    this.dataSource = new MatTreeNestedDataSource<StructureNode>();
  }

  hasChild = (_: number, node: StructureNode) =>
    !!node.children && node.children.length > 0;

  ngAfterViewInit(): void {
    this.route.params.subscribe((params) => {
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
    const routerLink: string = '';
    const messageNode: StructureNode = {
      displayText: displayText,
      type: 'message',
      routerLink: routerLink,
      children: children,
    }
    return messageNode;
  }

  getFileStructureNode(fileRecordObject: FileRecordObject): StructureNode {
    const displayText: DisplayText = {
      title: 'Akte: ' + fileRecordObject.generalMetadata.xdomeaID,
      subtitle: fileRecordObject.generalMetadata.subject,
    }
    const routerLink: string = 'akte/' + fileRecordObject.id;
    fileRecordObject.generalMetadata.xdomeaID
    const node: StructureNode = {
      displayText: displayText,
      type: 'file',
      routerLink: routerLink,
    };
    return node;
  }
}
