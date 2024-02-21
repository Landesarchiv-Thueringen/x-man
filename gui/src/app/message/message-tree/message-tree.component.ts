import { Clipboard } from '@angular/cdk/clipboard';
import { BreakpointObserver } from '@angular/cdk/layout';
import { FlatTreeControl } from '@angular/cdk/tree';
import { DOCUMENT } from '@angular/common';
import { AfterViewInit, Component, Inject, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatDialog } from '@angular/material/dialog';
import { MatSidenav } from '@angular/material/sidenav';
import { MatTree, MatTreeFlatDataSource, MatTreeFlattener } from '@angular/material/tree';
import { ActivatedRoute, NavigationEnd, Router } from '@angular/router';
import { BehaviorSubject, filter, first, switchMap, tap } from 'rxjs';
import { Process, ProcessService, ProcessStep } from 'src/app/process/process.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { Task } from '../../admin/tasks/tasks.service';
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
export class MessageTreeComponent implements AfterViewInit {
  @ViewChild('messageTree') messageTree?: MatTree<StructureNode>;
  @ViewChild(MatSidenav) sidenav?: MatSidenav;

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
  sidenavMode: 'side' | 'over' = 'side';

  constructor(
    @Inject(DOCUMENT) private document: Document,
    private breakpointObserver: BreakpointObserver,
    private clipboard: Clipboard,
    private dialog: MatDialog,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
    private router: Router,
  ) {
    this.route.params
      .pipe(
        // Get message
        switchMap((params) => this.messageService.getMessage(params['id'])),
        tap((message) => (this.message = message)),
        // Get process, update periodically
        switchMap((message) => this.processService.observeProcessByXdomeaID(message.messageHead.processID)),
        tap((process) => (this.process = process)),
        // Initialize tree
        //
        // Only proceed from here if the tree hasn't already been initialized
        // with the current message.
        filter(() => this.dataSource.data?.[0]?.id !== this.message?.id),
        tap(() => this.initTree()),
        // Wait for view initialized
        switchMap(() => this.viewInitialized.pipe(first((done) => done))),
        switchMap(() => this.route.firstChild!.params),
        // Expand current node and scroll into view
        tap((params) => {
          if (params['id']) {
            this.expandNode(params['id']);
            this.document.getElementById(params['id'])?.scrollIntoView({ block: 'center' });
          }
        }),
        takeUntilDestroyed(),
      )
      .subscribe();

    // Show sidenav as overlay on screens smaller than 1700px.
    this.breakpointObserver
      .observe(['(min-width: 1700px)'])
      .pipe(takeUntilDestroyed())
      .subscribe((result) => {
        if (result.matches) {
          this.sidenavMode = 'side';
          this.sidenav?.open();
        } else {
          this.sidenavMode = 'over';
        }
      });
    // Close the sidenav on navigation when in overlay mode.
    this.router.events
      .pipe(
        takeUntilDestroyed(),
        filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      )
      .subscribe(() => {
        if (this.sidenavMode === 'over') {
          this.sidenav?.close();
        }
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
    this.processService.getReport(this.process!.xdomeaID).subscribe((report) => {
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
    this.router.navigate(['nachricht', this.message?.id, 'details']);
  }
}
