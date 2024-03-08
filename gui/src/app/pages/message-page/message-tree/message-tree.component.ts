import { Clipboard } from '@angular/cdk/clipboard';
import { FlatTreeControl } from '@angular/cdk/tree';
import { CommonModule, DOCUMENT } from '@angular/common';
import { AfterViewInit, Component, Inject, ViewChild } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTree, MatTreeFlatDataSource, MatTreeFlattener, MatTreeModule } from '@angular/material/tree';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { BehaviorSubject, filter, first, switchMap } from 'rxjs';
import {
  DisplayText,
  Message,
  MessageService,
  MultiAppraisalResponse,
  StructureNode,
  StructureNodeType,
} from '../../../services/message.service';
import { NotificationService } from '../../../services/notification.service';
import { Process, ProcessService, ProcessStep } from '../../../services/process.service';
import { Task } from '../../../services/tasks.service';
import { MessagePageService } from '../message-page.service';
import { RecordObjectAppraisalPipe } from '../metadata/record-object-appraisal-pipe';
import { AppraisalFormComponent } from './appraisal-form/appraisal-form.component';
import { FinalizeAppraisalDialogComponent } from './finalize-appraisal-dialog/finalize-appraisal-dialog.component';
import { StartArchivingDialogComponent } from './start-archiving-dialog/start-archiving-dialog.component';

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
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatButtonModule,
    MatCheckboxModule,
    MatIconModule,
    MatMenuModule,
    MatToolbarModule,
    MatTreeModule,
    RecordObjectAppraisalPipe,
    RouterModule,
  ],
})
export class MessageTreeComponent implements AfterViewInit {
  @ViewChild('messageTree') messageTree?: MatTree<StructureNode>;

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

  process?: Process;
  message?: Message;
  showAppraisal = false;
  showSelection = false;
  selectedNodes: string[] = [];
  treeControl = new FlatTreeControl<FlatNode>(
    (node) => node.level,
    (node) => node.expandable,
  );
  treeFlattener = new MatTreeFlattener(
    this._transformer,
    (node) => node.level,
    (node) => node.expandable,
    (node) => node.children,
  );
  dataSource = new MatTreeFlatDataSource(this.treeControl, this.treeFlattener);
  viewInitialized = new BehaviorSubject(false);

  constructor(
    @Inject(DOCUMENT) private document: Document,
    private clipboard: Clipboard,
    private dialog: MatDialog,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
    private router: Router,
    private messagePage: MessagePageService,
  ) {
    this.messagePage.observeProcess().subscribe((process) => (this.process = process));
    this.messagePage.observeMessage().subscribe((message) => {
      this.message = message;
      this.initTree();
      this.viewInitialized
        .pipe(first((done) => done))
        .pipe(switchMap(() => this.route.firstChild!.params))
        .subscribe((params) => {
          if (params['id']) {
            this.expandNode(params['id']);
            this.document.getElementById(params['id'])?.scrollIntoView({ block: 'center' });
          }
        });
    });
  }

  ngAfterViewInit() {
    this.viewInitialized.next(true);
  }

  hasChild = (_: number, node: FlatNode) => node.expandable;

  trackTree(index: number, element: FlatNode): string {
    return element.id;
  }

  initTree(): void {
    if (this.message && this.process) {
      this.showAppraisal = this.message.messageType.code === '0501';
      const treeData: StructureNode[] = [];
      const messageNode = this.messageService.processMessage(this.process, this.message);
      treeData.push(messageNode);
      this.dataSource.data = treeData;
      this.expandNode(messageNode.id);
      this.messageService.watchNodeChanges().subscribe((changedNode: StructureNode) => {
        // TODO: find better solution than manipulating the flat nodes directly
        this.updateFlatNodeInTreeControl(changedNode);
      });
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
          return this.messageService.setAppraisalForMultipleRecordObjects(
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
            // Navigate to the tree root so the user sees the new status
            this.goToRootNode();
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
            // marks archiving process as started
            // hides the button to start the process
            this.process?.processState.archiving.tasks.push({
              state: 'running',
            } as Task);
            // Navigate to the tree root so the user sees the new status
            this.goToRootNode();
            return this.messageService.archive0503Message(this.message!.id);
          }),
        )
        .subscribe({
          error: (error: any) => {
            this.notificationService.show('Archivierung fehlgeschlagen');
            console.error(error);
          },
          next: () => {
            this.notificationService.show('Archivierung gestartet...');
          },
        });
    }
  }

  downloadReport() {
    this.processService.getReport(this.process!.id).subscribe((report) => {
      const a = document.createElement('a');
      document.body.appendChild(a);
      a.download = `Übernahmebericht ${this.process!.agency.abbreviation} ${this.process!.receivedAt}.pdf`;
      a.href = window.URL.createObjectURL(report);
      a.click();
      document.body.removeChild(a);
    });
  }

  /**
   * Returns true if the given process step is ready to be started by the user.
   *
   * This only considers the steps state. You have to check separately whether
   * - external conditions for running the step are fulfilled, and
   * - the process does not have any unresolved problems
   */
  canStartStep(processStep: ProcessStep): boolean {
    return processStep.tasks.every((task) => task.state === 'failed');
  }

  private goToRootNode() {
    this.router.navigate(['nachricht', this.process?.id, this.message?.messageType.code, 'details']);
  }
}
