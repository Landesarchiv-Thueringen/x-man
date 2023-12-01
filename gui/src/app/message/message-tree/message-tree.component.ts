// angular
import {
  AfterViewInit,
  Component,
  Inject,
  OnDestroy,
  ViewChild,
} from '@angular/core';
import { Clipboard } from '@angular/cdk/clipboard';
import { DOCUMENT } from '@angular/common';
import { ActivatedRoute, Params } from '@angular/router';

// material
import { FlatTreeControl } from '@angular/cdk/tree';
import {
  MatTree,
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
import { Process, ProcessService } from 'src/app/process/process.service';

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
  process?: Process;
  message?: Message;
  showAppraisal: boolean;
  messageTreeInit: boolean;

  @ViewChild('messageTree')
  messageTree?: MatTree<StructureNode>;

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
    private clipboard: Clipboard,
    @Inject(DOCUMENT) private document: Document,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
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
            this.message = message;
            return this.processService.getProcessByXdomeaID(message.messageHead.processID);
          }),
          switchMap((process: Process) => {
            this.process = process;
            this.initTree();
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
            const nodeElement: HTMLElement | null =
              this.document.getElementById(nodeID);
            if (nodeElement) {
              nodeElement.scrollIntoView({block: 'center'});
            }
          }
        });
    } else {
      this.urlParameterSubscription = this.route.params
        .pipe(
          switchMap((params: Params) => {
            return this.messageService.getMessage(params['id']);
          }),
          switchMap((message: Message) => {
            this.message = message;
            return this.processService.getProcessByXdomeaID(message.messageHead.processID);
          }),
        )
        .subscribe((process: Process) => {
          this.process = process;
          this.initTree();
        });
    }
  }

  initTree(): void {
    if (this.message) {
      this.showAppraisal = this.message.messageType.code === '0501';
      const treeData: StructureNode[] = [];
      const messageNode = this.messageService.processMessage(this.message);
      treeData.push(messageNode);
      this.dataSource.data = treeData;
      this.expandNode(messageNode.id);
      // updating the whole tree loses all informationen on expanded nodes
      // this.messageService
      //   .watchStructureNodes()
      //   .subscribe((nodes: StructureNode[]) => {
      //     this.dataSource.data = nodes;
      //   });
      this.messageService
        .watchNodeChanges()
        .subscribe((changedNode: StructureNode) => {
          // TODO: find better solution than manipulating the flat nodes directly
          this.updateFlatNodeInTreeControlRec(changedNode);
          // initialize next nodes with root nodes of tree
          // update the changed node doesn't trigger updates of the corresponding flat node
          //this.updateNodeInDataSource(changedNode);
        });
    }
  }

  // updates flat node and all children
  updateFlatNodeInTreeControlRec(changedNode: StructureNode): void {
    const nextNodes: StructureNode[] = [changedNode];
    while (nextNodes.length !== 0) {
      // shift is breadth-first, pop is depth-first
      const currentNode: StructureNode = nextNodes.shift()!;
      this.updateFlatNodeInTreeControl(currentNode);
      if (currentNode.children) {
        nextNodes.push(...currentNode.children);
      }
    }
  }

  updateFlatNodeInTreeControl(changedNode: StructureNode): void {
    const flatNode: FlatNode = this.treeControl.dataNodes.find(
      (n: FlatNode) => n.id === changedNode.id
    )!;
    flatNode.appraisal = changedNode.appraisal;
    flatNode.displayText = changedNode.displayText;
    flatNode.routerLink = changedNode.routerLink;
    flatNode.type = changedNode.type;
  }

  updateNodeInDataSource(changedNode: StructureNode): void {
    const oldNode: StructureNode = this.findNodeInDataSource(changedNode.id)!;
    if (oldNode.parentID) {
      const parentNode: StructureNode = this.findNodeInDataSource(
        oldNode.parentID
      )!;
      const oldNodeIndex: number = parentNode.children!.findIndex(
        (node: StructureNode) => node.id === oldNode.id
      )!;
      parentNode.children![oldNodeIndex] = changedNode;
    } else {
      // element must be root element
      const oldNodeIndex: number = this.dataSource.data.findIndex(
        (node: StructureNode) => node.id === oldNode.id
      )!;
      this.dataSource.data[oldNodeIndex] = changedNode;
    }
  }

  findNodeInDataSource(targetID: string): StructureNode | undefined {
    // initialize next nodes with root nodes of tree
    const nextNodes: StructureNode[] = [...this.dataSource.data];
    while (nextNodes.length !== 0) {
      // shift is breadth-first search, pop is depth-first search
      const currentNode: StructureNode = nextNodes.shift()!;
      if (currentNode.id === targetID) {
        return currentNode;
      }
      if (currentNode.children) {
        nextNodes.push(...currentNode.children);
      }
    }
    return undefined;
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
          this.message!.appraisalComplete = true;
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

  copyMessageUrl() {
    this.clipboard.copy(this.document.location.toString());
    this.notificationService.show(
      'Nachrichten-Link in Zwischenspeicher kopiert'
    );
  }

  archive0503Message() {
    if (this.message) {
      this.messageService.archive0503Message(this.message.id).subscribe({
        error: (error: any) => {
          console.error(error);
        },
        next: () => {
          console.log('yeah');
        }
      });
    }
  }
}
