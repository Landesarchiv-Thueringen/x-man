// angular
import { AfterViewInit, Component, Inject, OnDestroy } from '@angular/core';
import { DOCUMENT } from '@angular/common';
import { ActivatedRoute, Params } from '@angular/router';

// material
import { FlatTreeControl } from '@angular/cdk/tree';
import {
  MatTreeFlatDataSource,
  MatTreeFlattener,
} from '@angular/material/tree';

// project
import {
  DisplayText,
  Message,
  MessageService,
  StructureNode,
  StructureNodeType,
} from '../message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';

export interface FlatNode {
  id: string;
  parentID?: string;
  expandable: boolean;
  level: number;
  displayText: DisplayText;
  type: StructureNodeType;
  routerLink: string;
  appraisal?: string;
}

@Component({
  selector: 'app-message-tree',
  templateUrl: './message-tree.component.html',
  styleUrls: ['./message-tree.component.scss'],
})
export class MessageTreeComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  message?: Message;
  showAppraisal: boolean;
  messageTreeInit: boolean;

  treeControl: FlatTreeControl<FlatNode>;
  treeFlattener: MatTreeFlattener<StructureNode, FlatNode>;
  dataSource: MatTreeFlatDataSource<StructureNode, FlatNode>;

  private _transformer = (node: StructureNode, level: number): FlatNode => {
    return {
      id: node.id,
      parentID: node.parentID,
      expandable: !!node.children && node.children.length > 0,
      level: level,
      displayText: node.displayText,
      type: node.type,
      routerLink: node.routerLink,
      appraisal: node.appraisal,
    };
  };

  constructor(
    @Inject(DOCUMENT) private document: Document,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute
  ) {
    this.messageTreeInit = true;
    this.showAppraisal = true;
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

  hasChild = (_: number, node: FlatNode) => node.expandable;

  ngAfterViewInit(): void {
    this.urlParameterSubscription?.unsubscribe();
    if (this.route.firstChild) {
      this.urlParameterSubscription = this.route.params
        .pipe(
          switchMap((params: Params) => {
            return this.messageService.getMessage(params['id']);
          }),
          switchMap((message: Message) => {
            this.initTree(message);
            return this.route.firstChild!.params;
          })
        )
        .subscribe((params: Params) => {
          const nodeID: string = params['id'];
          // expand node children and scroll selected node into view if message tree is initialized
          // when opening a link to a node, it gets scrolled into view and expanded
          if (nodeID && this.messageTreeInit) {
            this.messageTreeInit = false;
            this.expandNode(nodeID);
            this.document.getElementById(nodeID)?.scrollIntoView();
          }
        });
    } else {
      this.urlParameterSubscription = this.route.params
        .pipe(
          switchMap((params: Params) => {
            return this.messageService.getMessage(params['id']);
          })
        )
        .subscribe((message: Message) => {
          this.initTree(message);
        });
    }
  }

  initTree(message: Message): void {
    this.message = message;
    this.showAppraisal = this.message.messageType.code === '0501';
    const treeData: StructureNode[] = [];
    const messageNode = this.messageService.processMessage(message);
    treeData.push(messageNode);
    this.dataSource.data = treeData;
    console.log(this.treeControl.dataNodes);
    // this.treeControl.dataNodes = treeData;
    // this.treeControl.expand(messageNode);
    this.expandNode(messageNode.id);
  }

  sendAppraisalMessage(): void {
    if (this.message) {
      this.messageService.finalizeMessageAppraisal(this.message.id).subscribe({
        error: (error) => {
          console.error(error);
        },
        next: () => {
          this.notificationService.show(
            'Bewertungsnachricht wurde erfolgreich versandt'
          );
        },
      });
    }
  }

  expandNode(id: string): void {
    const node: FlatNode | undefined = this.treeControl.dataNodes.find(
      (n: FlatNode) => n.id === id
    );
    if (node) {
      this.treeControl.expand(node);
      if (node.parentID) {
        this.expandNode(node.parentID);
      }
    }
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
