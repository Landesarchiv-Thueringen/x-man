import { Clipboard } from '@angular/cdk/clipboard';
import { FlatTreeControl } from '@angular/cdk/tree';
import { DOCUMENT } from '@angular/common';
import { AfterViewInit, Component, Inject, OnDestroy, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatTree, MatTreeFlatDataSource, MatTreeFlattener } from '@angular/material/tree';
import { ActivatedRoute, Params } from '@angular/router';
import { Subscription, filter, switchMap } from 'rxjs';
import { Process, ProcessService } from 'src/app/process/process.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { AppraisalFormComponent } from '../appraisal-form/appraisal-form.component';
import { FinalizeAppraisalDialogComponent } from '../finalize-appraisal-dialog/finalize-appraisal-dialog.component';
import {
  DisplayText,
  Message,
  MessageService,
  MultiAppraisalResponse,
  StructureNode,
  StructureNodeType,
} from '../message.service';
import { StartArchivingDialogComponent } from '../start-archiving-dialog/start-archiving-dialog.component';

export interface FlatNode {
  id: string;
  selected: boolean;
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
  showSelection: boolean;
  selectedNodes: string[];

  @ViewChild('messageTree')
  messageTree?: MatTree<StructureNode>;

  treeControl: FlatTreeControl<FlatNode>;
  treeFlattener: MatTreeFlattener<StructureNode, FlatNode>;
  dataSource: MatTreeFlatDataSource<StructureNode, FlatNode>;

  private _transformer = (node: StructureNode, level: number): FlatNode => {
    return {
      id: node.id,
      selected: node.selected,
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
    private dialog: MatDialog,
    @Inject(DOCUMENT) private document: Document,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
  ) {
    this.messageTreeInit = true;
    this.showAppraisal = false;
    this.showSelection = false;
    this.selectedNodes = [];
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
          }),
        )
        .subscribe((params: Params) => {
          const nodeID: string = params['id'];
          // expand node children and scroll selected node into view if message tree is initialized
          // when opening a link to a node, it gets scrolled into view and expanded
          if (nodeID && this.messageTreeInit) {
            this.messageTreeInit = false;
            this.expandNode(nodeID);
            const nodeElement: HTMLElement | null = this.document.getElementById(nodeID);
            if (nodeElement) {
              nodeElement.scrollIntoView({ block: 'center' });
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

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }

  initTree(): void {
    if (this.message && this.process) {
      this.showAppraisal = this.message.messageType.code === '0501';
      const treeData: StructureNode[] = [];
      const messageNode = this.messageService.processMessage(this.process, this.message);
      treeData.push(messageNode);
      this.dataSource.data = treeData;
      this.expandNode(messageNode.id);
      // updating the whole tree loses all informationen on expanded nodes
      // this.messageService
      //   .watchStructureNodes()
      //   .subscribe((nodes: StructureNode[]) => {
      //     this.dataSource.data = nodes;
      //   });
      this.messageService.watchNodeChanges().subscribe((changedNode: StructureNode) => {
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
    const flatNode: FlatNode = this.treeControl.dataNodes.find((n: FlatNode) => n.id === changedNode.id)!;
    flatNode.selected = changedNode.selected;
    flatNode.appraisal = changedNode.appraisal;
    flatNode.displayText = changedNode.displayText;
    flatNode.routerLink = changedNode.routerLink;
    flatNode.type = changedNode.type;
  }

  // The following functions don't work as expected

  // updateNodeInDataSource(changedNode: StructureNode): void {
  //   const oldNode: StructureNode = this.findNodeInDataSource(changedNode.id)!;
  //   if (oldNode.parentID) {
  //     const parentNode: StructureNode = this.findNodeInDataSource(
  //       oldNode.parentID
  //     )!;
  //     const oldNodeIndex: number = parentNode.children!.findIndex(
  //       (node: StructureNode) => node.id === oldNode.id
  //     )!;
  //     parentNode.children![oldNodeIndex] = changedNode;
  //   } else {
  //     // element must be root element
  //     const oldNodeIndex: number = this.dataSource.data.findIndex(
  //       (node: StructureNode) => node.id === oldNode.id
  //     )!;
  //     this.dataSource.data[oldNodeIndex] = changedNode;
  //   }
  // }

  // findNodeInDataSource(targetID: string): StructureNode | undefined {
  //   // initialize next nodes with root nodes of tree
  //   const nextNodes: StructureNode[] = [...this.dataSource.data];
  //   while (nextNodes.length !== 0) {
  //     // shift is breadth-first search, pop is depth-first search
  //     const currentNode: StructureNode = nextNodes.shift()!;
  //     if (currentNode.id === targetID) {
  //       return currentNode;
  //     }
  //     if (currentNode.children) {
  //       nextNodes.push(...currentNode.children);
  //     }
  //   }
  //   return undefined;
  // }

  expandNode(id: string): void {
    const node: FlatNode | undefined = this.treeControl.dataNodes.find((n: FlatNode) => n.id === id);
    if (node) {
      this.treeControl.expand(node);
      if (node.parentID) {
        this.expandNode(node.parentID);
      }
    }
  }

  enableSelection(): void {
    this.showSelection = true;
  }

  disableSelection(): void {
    this.selectedNodes.forEach((nodeID: string) => {
      const node: StructureNode | undefined = this.messageService.getStructureNode(nodeID);
      if (node) {
        node.selected = false;
        this.messageService.updateStructureNode(node);
      }
    });
    this.selectedNodes = [];
    this.showSelection = false;
  }

  selectItem(selected: boolean, id: string): void {
    if (selected) {
      this.selectedNodes.push(id);
    } else {
      this.selectedNodes = this.selectedNodes.filter((nodeID) => nodeID !== id);
    }
    const node: StructureNode | undefined = this.messageService.getStructureNode(id);
    if (node) {
      node.selected = selected;
      node.children?.forEach((nodeChild: StructureNode) => {
        if (
          nodeChild.type === 'file' ||
          nodeChild.type === 'subfile' ||
          nodeChild.type === 'process' ||
          nodeChild.type === 'subprocess'
        ) {
          this.selectItem(selected, nodeChild.id);
        }
      });
      this.messageService.updateStructureNode(node);
    }
  }

  copyMessageUrl() {
    this.clipboard.copy(this.document.location.toString());
    this.notificationService.show('Objekt-Link in Zwischenspeicher kopiert');
  }

  setAppraisalForMultipleRecorcObjects(): void {
    this.dialog
      .open(AppraisalFormComponent, { autoFocus: false })
      .afterClosed()
      .pipe(
        filter((formResult) => !!formResult),
        switchMap((formResult: any) => {
          return this.messageService.setAppraisalForMultipleRecorcObjects(
            this.selectedNodes,
            formResult.appraisalCode,
            formResult.appraisalNote,
          );
        }),
      )
      .subscribe({
        error: (error: any) => {
          console.error(error);
          this.notificationService.show('Bewertung konnte nicht gespeichert werden');
          this.disableSelection();
        },
        next: (response: MultiAppraisalResponse) => {
          for (let fileRecordObject of response.updatedFileRecordObjects) {
            this.messageService.updateStructureNodeForRecordObject(fileRecordObject);
          }
          for (let processRecordObject of response.updatedProcessRecordObjects) {
            this.messageService.updateStructureNodeForRecordObject(processRecordObject);
          }
          this.notificationService.show('Bewertung erfolgreich gespeichert');
          this.disableSelection();
        },
      });
  }

  sendAppraisalMessage(): void {
    if (this.message) {
      this.dialog
        .open(FinalizeAppraisalDialogComponent, {
          autoFocus: false,
          data: { messageID: this.message.id },
        })
        .afterClosed()
        .pipe(
          filter((formResult) => !!formResult),
          switchMap(() => {
            return this.messageService.finalizeMessageAppraisal(this.message!.id);
          }),
        )
        .subscribe({
          error: (error: any) => {
            console.error(error);
          },
          next: () => {
            this.notificationService.show('Bewertungsnachricht wurde erfolgreich versandt');
            this.message!.appraisalComplete = true;
          },
        });
    }
  }

  archive0503Message() {
    if (this.message) {
      this.dialog
        .open(StartArchivingDialogComponent, {
          autoFocus: false,
        })
        .afterClosed()
        .pipe(
          filter((formResult) => !!formResult),
          switchMap(() => {
            return this.messageService.archive0503Message(this.message!.id);
          }),
        )
        .subscribe({
          error: (error: any) => {
            this.notificationService.show('Archivierung fehlgeschlagen');
            console.error(error);
          },
          next: () => {
            this.notificationService.show('Archivierung erfolgreich abgeschlossen');
          },
        });
    }
  }
}
