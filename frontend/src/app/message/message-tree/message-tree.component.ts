// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import {
  DocumentRecordObject,
  FileRecordObject,
  Message,
  MessageService,
  ProcessRecordObject,
} from '../message.service';

// utility
import { Subscription } from 'rxjs';

type StructureNodeType = 'message' | 'file' | 'process' | 'document';

export interface DisplayText {
  title: string;
  subtitle?: string;
}

export interface StructureNode {
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  appraisal?: string;
  children?: StructureNode[];
}

@Component({
  selector: 'app-message-tree',
  templateUrl: './message-tree.component.html',
  styleUrls: ['./message-tree.component.scss'],
})
export class MessageTreeComponent implements AfterViewInit, OnDestroy {
  treeControl: NestedTreeControl<StructureNode>;
  dataSource: MatTreeNestedDataSource<StructureNode>;
  urlParameterSubscription?: Subscription;
  message?: Message;

  constructor(
    private messageService: MessageService,
    private route: ActivatedRoute
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
      this.messageService
        .getMessage(params['id'])
        .subscribe((message: Message) => {
          this.message = message;
          const treeData: StructureNode[] = [];
          const messageNode = this.messageService.processMessage(message);
          treeData.push(messageNode);
          this.dataSource.data = treeData;
          this.treeControl.dataNodes = treeData;
          this.treeControl.expand(messageNode);
        });
    });
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
