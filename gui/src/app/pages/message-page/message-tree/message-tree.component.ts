import { Clipboard } from '@angular/cdk/clipboard';
import { FlatTreeControl } from '@angular/cdk/tree';
import { CommonModule, DOCUMENT } from '@angular/common';
import {
  AfterViewInit,
  Component,
  Inject,
  QueryList,
  ViewChild,
  ViewChildren,
  computed,
} from '@angular/core';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipEditedEvent, MatChipRow, MatChipsModule } from '@angular/material/chips';
import { MatRippleModule } from '@angular/material/core';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTree, MatTreeModule } from '@angular/material/tree';
import { ActivatedRoute, ChildActivationEnd, Router, RouterModule } from '@angular/router';
import { ReplaySubject, Subject, combineLatest, concat, delay, filter, switchMap } from 'rxjs';
import { Appraisal } from '../../../services/appraisal.service';
import { AuthService } from '../../../services/auth.service';
import { ConfigService } from '../../../services/config.service';
import { Message, MessageService } from '../../../services/message.service';
import { NotificationService } from '../../../services/notification.service';
import { ProcessService, SubmissionProcess } from '../../../services/process.service';
import { Records } from '../../../services/records.service';
import { notNull } from '../../../utils/predicates';
import { MessagePageService } from '../message-page.service';
import { MessageProcessorService, StructureNode } from '../message-processor.service';
import { RecordAppraisalPipe } from '../metadata/record-appraisal-pipe';
import { AppraisalFormComponent } from './appraisal-form/appraisal-form.component';
import { FinalizeAppraisalDialogComponent } from './finalize-appraisal-dialog/finalize-appraisal-dialog.component';
import { FilterResult, FlatNode, MessageTreeDataSource } from './message-tree-data-source';
import { StartArchivingDialogComponent } from './start-archiving-dialog/start-archiving-dialog.component';

export interface Filter {
  /** A unique string to identify the filter. */
  type: string;
  /** A label shown to the user. */
  label: string;
  /** An optional filter value to be entered by the user and passed to the predicate. */
  value?: string;
  /** An optional condition for when to show the filter in the menu. */
  showIf?: () => boolean;
  /** The filter predicate that decides whether to include a node in results. */
  predicate: (node: StructureNode, value?: string) => FilterResult;
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
    MatChipsModule,
    MatIconModule,
    MatMenuModule,
    MatRippleModule,
    MatToolbarModule,
    MatTreeModule,
    RecordAppraisalPipe,
    RouterModule,
  ],
})
export class MessageTreeComponent implements AfterViewInit {
  @ViewChild('messageTree') messageTree?: MatTree<StructureNode>;
  @ViewChildren(MatChipRow) matChipRow?: QueryList<MatChipRow>;

  process?: SubmissionProcess;
  message?: Message;
  rootRecords?: Records;
  showAppraisal = false;
  showSelection = this.messagePage.showSelection;
  hasUnresolvedError = this.messagePage.hasUnresolvedError;
  isDisabled = computed(() => this.hasUnresolvedError() && !this.authService.isAdmin());
  selectedNodes = new Set<string>();
  intermediateNodes = new Set<string>();
  treeControl = new FlatTreeControl<FlatNode>(
    (node) => node.level,
    (node) => node.expandable,
  );

  dataSource = new MessageTreeDataSource(this.treeControl);
  viewInitialized = new ReplaySubject<void>(1);
  appraisals: { [recordId: string]: Appraisal } = {};
  readonly availableFilters: Filter[] = [
    {
      type: 'not-appraised',
      label: 'Noch nicht bewertet',
      showIf: () =>
        !this.process?.processState.appraisal.complete &&
        !this.process?.processState.receive0503.complete,
      predicate: (node) => {
        if (!node.canBeAppraised) {
          return 'propagate-recursive';
        } else if (
          !this.appraisals[node.recordId!]?.decision ||
          this.appraisals[node.recordId!].decision === 'B'
        ) {
          return 'show';
        } else {
          return 'hide';
        }
      },
    },
    {
      type: 'record-plan-id',
      label: 'Aktenplanschlüssel',
      value: '',
      predicate: (node, value) =>
        node.generalMetadata?.filePlan?.filePlanNumber?.toString() === value
          ? 'show-recursive'
          : 'hide-recursive',
    },
  ];
  activeFilters: Filter[] = [];
  filtersHint: string | null = null;
  currentRecordId?: string;
  config = toSignal(this.configService.config);

  constructor(
    @Inject(DOCUMENT) private document: Document,
    private clipboard: Clipboard,
    private configService: ConfigService,
    private dialog: MatDialog,
    private messagePage: MessagePageService,
    private messageProcessor: MessageProcessorService,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
    private router: Router,
    private authService: AuthService,
  ) {
    this.registerAppraisals();
    const processReady = new Subject<void>();
    this.messagePage.observeProcessData().subscribe(({ process }) => {
      this.process = process;
      processReady.complete();
    });
    // Update currentRecordId with the record ID in the URL.
    this.router.events
      .pipe(
        takeUntilDestroyed(),
        filter((e) => e instanceof ChildActivationEnd),
        switchMap(() => this.route.firstChild!.params),
      )
      .subscribe((params) => {
        this.currentRecordId = params['id'];
      });
    // Update the tree when `message` changes.
    concat(
      processReady,
      combineLatest([this.messagePage.observeMessage(), this.messagePage.observeRootRecords()]),
    )
      .pipe(filter(notNull))
      .subscribe(async ([message, rootRecords]) => {
        this.message = message;
        this.rootRecords = rootRecords;
        await this.initTree();
      });
    // Expand the current node when display data is updated.
    this.dataSource
      .observeDisplayData()
      .pipe(filter(notNull), delay(0))
      .subscribe(() => {
        // Expand message node.
        this.treeControl.expand(this.treeControl.dataNodes?.[0]);
        // Expand current node, if any and visible.
        if (this.currentRecordId) {
          this.expandNode(this.currentRecordId);
          this.document.getElementById(this.currentRecordId)?.scrollIntoView({ block: 'center' });
        }
      });
  }

  ngAfterViewInit() {
    this.viewInitialized.next(void 0);
  }

  hasChild = (_: number, node: FlatNode) => node.expandable;

  trackTree(index: number, element: FlatNode): string {
    return element.id;
  }

  addFilter(filter: Filter): void {
    if (filter.value == null) {
      this.activeFilters.push(filter);
      this.applyFilters();
    } else {
      // If the filter has a value, insert the chip in editing mode.
      this.activeFilters.push({ ...filter, value: '' });
      // Start editing the chip value.
      setTimeout(() => {
        const chipRow = this.matChipRow!.last;
        chipRow._elementRef.nativeElement.dispatchEvent(new Event('dblclick'));
        // Force a space after the colon.
        setTimeout(() => {
          const editInput = chipRow.defaultEditInput!;
          editInput.getNativeElement().innerHTML = `${filter.label}:&nbsp;`;
          editInput['_moveCursorToEndOfInput']();
        });
      });
      this.filtersHint = `Geben Sie einen Wert ein, um nach ${filter.label} zu filtern, und bestätigen Sie Ihre Eingabe mit Enter.`;
    }
  }

  onFilterEdited(event: MatChipEditedEvent, filter: Filter): void {
    const value = event.value.replace(new RegExp(filter.label + ':?'), '').trim();
    if (value) {
      filter.value = value;
      setTimeout(() => {
        this.applyFilters();
      });
    } else {
      this.removeFilter(filter);
    }
    this.filtersHint = null;
  }

  filterHasValue(filter: Filter): boolean {
    return typeof filter.value === 'string';
  }

  hasFilter(filter: Filter): boolean {
    return this.activeFilters.some((f) => f.type === filter.type);
  }

  removeFilter(filter: Filter): void {
    this.activeFilters = this.activeFilters.filter((f) => f != filter);
    this.applyFilters();
  }

  private applyFilters(): void {
    this.dataSource.filters = this.activeFilters.map(
      (filter) => (node) => filter.predicate(node, filter.value),
    );
  }

  async initTree(): Promise<void> {
    if (this.message && this.process && this.rootRecords) {
      this.showAppraisal = this.message.messageType === '0501';
      const rootNode = await this.messageProcessor.processMessage(
        this.process,
        this.message,
        this.rootRecords,
      );
      this.dataSource.data = rootNode;
    }
  }

  expandNode(id: string): void {
    const node = this.treeControl.dataNodes.find((n: FlatNode) => n.id === id);
    if (node) {
      this.treeControl.expand(node);
      if (node.parentId) {
        this.expandNode(node.parentId);
      }
    }
  }

  enableSelection(): void {
    this.showSelection.set(true);
  }

  disableSelection(): void {
    this.selectedNodes.clear();
    this.intermediateNodes.clear();
    this.showSelection.set(false);
  }

  selectItem(selected: boolean, id: string, propagating: 'down' | 'up' | null = null): void {
    if (propagating !== 'up') {
      this.intermediateNodes.delete(id);
    }
    if (selected) {
      this.selectedNodes.add(id);
    } else {
      this.selectedNodes.delete(id);
    }
    const node = this.dataSource.getNode(id);
    // Propagate the selection down to the node's children if the selection is
    // not already the result of a selection of one of its children.
    if (node.children && propagating !== 'up') {
      for (const child of node.children) {
        // Note that we set the selection state even for nodes that we don't
        // allow to be appraised directly in the UI in order to send a complete
        // list to the backend for the multi-appraisal request.
        if (
          child.type === 'file-group' ||
          child.type === 'file' ||
          child.type === 'subfile' ||
          child.type === 'process-group' ||
          child.type === 'process' ||
          child.type === 'subprocess'
        ) {
          this.selectItem(selected, child.id, 'down');
        } else if (selected && child.type === 'filtered') {
          // If the current node has filtered child nodes, we cannot select it.
          // The filtered child nodes might have conflicting appraisals to what
          // the user is going to choose for selected nodes.
          this.selectedNodes.delete(id);
          this.intermediateNodes.add(id);
        }
      }
    }
    // Propagate the selection up to the node's parent if the selection is not
    // already the result of the selection of its parent.
    if (node.parentId && propagating !== 'down') {
      const parent = this.dataSource.getNode(node.parentId);
      // If all of the parent's child nodes have the same selection state, let
      // the parent assume the same selection state.
      if (
        parent.children!.every((n) => this.selectedNodes.has(n.id) === selected) &&
        !parent.children!.some((n) => this.intermediateNodes.has(n.id))
      ) {
        this.intermediateNodes.delete(parent.id);
        this.selectItem(selected, parent.id, 'up');
      } else {
        // If not, mark the parent deselected and give it an intermediate
        // selection appearance.
        //
        // When sending a request to the backend for multi appraisal, the
        // backend will automatically change the appraisal decision of parent
        // nodes if necessary, so we can safely omit the now deselected parent
        // from the request.
        this.intermediateNodes.add(parent.id);
        this.selectItem(false, parent.id, 'up');
      }
    }
  }

  copyMessageUrl() {
    this.clipboard.copy(this.document.location.toString());
    this.notificationService.show('Objekt-Link in Zwischenspeicher kopiert');
  }

  setAppraisalForMultipleRecordObjects(): void {
    this.dialog
      .open(AppraisalFormComponent, { autoFocus: false })
      .afterClosed()
      .subscribe(async (formResult) => {
        if (formResult) {
          await this.messagePage.setAppraisals(
            [...this.selectedNodes]
              .map((id) => this.dataSource.getNode(id))
              .filter(
                (node) =>
                  node.type === 'file' ||
                  node.type === 'subfile' ||
                  node.type === 'process' ||
                  node.type === 'subprocess',
              )
              .map((node) => node.recordId!),
            formResult.appraisalCode,
            formResult.appraisalNote,
          );
          this.notificationService.show('Bewertung erfolgreich gespeichert');
          this.disableSelection();
        }
      });
  }

  getAppraisal(node: FlatNode): Appraisal | null {
    if (node.recordId) {
      return this.appraisals[node.recordId];
    } else {
      return null;
    }
  }

  private registerAppraisals(): void {
    this.messagePage
      .observeAppraisals()
      .pipe(takeUntilDestroyed())
      .subscribe((appraisals) => {
        for (const appraisal of appraisals) {
          this.appraisals[appraisal.recordId] = appraisal;
        }
      });
  }

  sendAppraisalMessage(): void {
    if (this.message) {
      this.dialog
        .open(FinalizeAppraisalDialogComponent, {
          autoFocus: false,
          data: { processId: this.message.messageHead.processID },
        })
        .afterClosed()
        .pipe(
          filter((formResult) => !!formResult),
          switchMap(() => this.messagePage.finalizeAppraisals()),
        )
        .subscribe({
          error: (error: any) => {
            console.error(error);
          },
          next: () => {
            // Navigate to the tree root so the user sees the new status
            this.goToRootNode();
            this.notificationService.show('Bewertungsnachricht wurde erfolgreich versandt');
          },
        });
    }
  }

  archive0503Message() {
    if (this.message) {
      this.dialog
        .open(StartArchivingDialogComponent, {
          autoFocus: false,
          data: {
            agency: this.process?.agency,
          },
        })
        .afterClosed()
        .pipe(
          filter((formResult) => !!formResult),
          switchMap((formResult) => {
            // Navigate to the tree root so the user sees the new status
            this.goToRootNode();
            return this.messageService.archive0503Message(
              this.message!.messageHead.processID,
              formResult.collectionId,
            );
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
    this.processService.getReport(this.process!.processId).subscribe((report) => {
      const a = document.createElement('a');
      document.body.appendChild(a);
      a.download = `Übernahmebericht ${this.process!.agency.abbreviation} ${this.process!.createdAt}.pdf`;
      a.href = window.URL.createObjectURL(report);
      a.click();
      document.body.removeChild(a);
    });
  }

  private goToRootNode() {
    this.router.navigate([
      'nachricht',
      this.process?.processId,
      this.message?.messageType,
      'details',
    ]);
  }
}
