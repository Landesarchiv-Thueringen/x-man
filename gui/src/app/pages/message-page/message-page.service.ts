import { Injectable, Signal, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed, toObservable, toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import { Map } from 'immutable';
import { filter, firstValueFrom, switchMap, tap } from 'rxjs';
import { Appraisal, AppraisalCode, AppraisalService } from '../../services/appraisal.service';
import { ProcessingError } from '../../services/clearing.service';
import { ConfigService } from '../../services/config.service';
import { Message, MessageService } from '../../services/message.service';
import {
  PackagingChoice,
  PackagingData,
  PackagingDecision,
  PackagingService,
  PackagingStats,
} from '../../services/packaging.service';
import { ProcessService, SubmissionProcess, Warning } from '../../services/process.service';
import {
  DocumentRecord,
  FileRecord,
  ProcessRecord,
  Records,
  RecordsService,
} from '../../services/records.service';
import { notEmpty } from '../../utils/predicates';
import { MessageProcessor, StructureNode } from './message-processor';

/**
 * A Service to provide data the the message page and its child components.
 *
 * It's lifetime is linked to that of the message-page component.
 */
@Injectable()
export class MessagePageService {
  private appraisalService = inject(AppraisalService);
  private configService = inject(ConfigService);
  private messageService = inject(MessageService);
  private processService = inject(ProcessService);
  private recordsService = inject(RecordsService);
  private packagingService = inject(PackagingService);
  private route = inject(ActivatedRoute);

  private readonly params = toSignal(this.route.params, { requireSync: true });

  /** The submission process's ID. Constant for the lifetime of the message page.  */
  readonly processId: string = this.params()['processId'];
  /**
   * The process references by the page URL.
   *
   * We update the process regularly by refetching it from the backend, however,
   * the process ID should not change for the lifetime of the message page.
   */
  readonly process = signal<SubmissionProcess | undefined>(undefined);
  readonly warnings = signal<Warning[]>([]);
  readonly processingErrors = signal<ProcessingError[]>([]);
  readonly hasUnresolvedError = computed(() => !!this.process()?.unresolvedErrors);
  readonly appraisalComplete = computed(
    () =>
      (this.process()?.processState.appraisal.complete ?? false) ||
      (this.process()?.processState.receive0503.complete ?? false),
  );

  /** The message being currently displayed. Controlled by an URL parameter. */
  readonly messageType = computed<'0501' | '0503' | ''>(() => this.params()['messageType']);
  /**
   * The message references by the page URL.
   *
   * We fetch the message once when messageType changes.
   */
  readonly message = signal<Message | undefined>(undefined);

  /**
   * The message's root records.
   *
   * We fetch the root records when the message changes.
   */
  readonly rootRecords = signal<Records | undefined>(undefined);
  /**
   * Record maps of root and all nested records.
   *
   * We update all record maps when the root records change.
   */
  readonly fileRecords = signal(Map<string, FileRecord>());
  readonly processRecords = signal(Map<string, ProcessRecord>());
  readonly documentsRecords = signal(Map<string, DocumentRecord>());

  /**
   * All appraisals for the current process.
   *
   * Appraisals can be updated by the user at any time.
   */
  readonly appraisals = signal<Map<string, Appraisal>>(Map());

  readonly packagingChoices = signal<{ [recordId in string]?: PackagingChoice }>({});
  readonly packagingDecisions = signal<{ [recordId in string]?: PackagingDecision }>({});
  readonly packagingStats = signal<{ [recordId in string]?: PackagingStats }>({});

  readonly selectionActive = signal(false);

  /** A tree structure that reflects the current message. */
  readonly treeRoot: Signal<StructureNode | undefined>;
  readonly treeNodes: Signal<Map<string, StructureNode>>;

  constructor() {
    this.registerMessage(this.processId);
    // Fetch appraisals and record options once, will be updated when changed
    // by other functions.
    this.appraisalService
      .getAppraisals(this.processId)
      .subscribe((appraisals) => this._setAppraisals(appraisals ?? []));
    // Observe process until destroyed and update `this.process`.
    this.processService
      .observeProcessData(this.processId)
      .pipe(takeUntilDestroyed())
      .subscribe((data) => {
        this.process.set(data.process);
        this.warnings.set(data.warnings);
        this.processingErrors.set(data.processingErrors);
      });
    // Register tree data.
    //
    // Compute `agencyName` here, so `initTree` won't be triggered each time the
    // process state changes.
    const agencyName = computed(() => this.process()?.agency.name);
    const treeData = computed(() => {
      if (agencyName() && this.message() && this.rootRecords() && this.configService.config()) {
        const processor = new MessageProcessor(this.configService.config()!);
        return processor.processMessage(agencyName()!, this.message()!, this.rootRecords()!);
      } else {
        return undefined;
      }
    });
    this.treeRoot = computed(() => treeData()?.root);
    this.treeNodes = computed(() => Map(treeData()?.map));
  }

  /**
   * Fetches the message and updates `this.message`.
   */
  private registerMessage(processId: string) {
    const messageType = toObservable(this.messageType).pipe(filter(notEmpty));
    messageType
      .pipe(switchMap((messageType) => this.messageService.getMessage(processId, messageType)))
      .subscribe((message) => this.message.set(message));
    messageType
      .pipe(
        tap(() => this.updateRecords(undefined)),
        switchMap((messageType) => this.recordsService.getRootRecords(processId, messageType)),
      )
      .subscribe((rootRecords) => {
        this.updateRecords(rootRecords);
      });
    // Fetch packaging data when first displaying the 0503 message.
    messageType
      .pipe(
        filter(
          (messageType) =>
            messageType === '0503' && Object.keys(this.packagingDecisions()).length === 0,
        ),
        switchMap(() => this.packagingService.getPackaging(processId)),
      )
      .subscribe((data) => this.setPackagingData(data));
  }

  /**
   * Clears the record maps and replaces their content with the given root
   * records.
   */
  private updateRecords(rootRecords?: Records) {
    let fileRecords = this.fileRecords().clear();
    let processRecords = this.processRecords().clear();
    let documentsRecords = this.documentsRecords().clear();
    const processDocument = (document: DocumentRecord) => {
      documentsRecords = documentsRecords.set(document.recordId, document);
      document.attachments?.forEach(processDocument);
    };
    const processProcess = (process: ProcessRecord) => {
      processRecords = processRecords.set(process.recordId, process);
      process.subprocesses?.forEach(processProcess);
      process.documents?.forEach(processDocument);
    };
    const processFile = (file: FileRecord) => {
      fileRecords = fileRecords.set(file.recordId, file);
      file.subfiles?.forEach(processFile);
      file.processes?.forEach(processProcess);
      file.documents?.forEach(processDocument);
    };
    rootRecords?.files?.forEach(processFile);
    rootRecords?.processes?.forEach(processProcess);
    rootRecords?.documents?.forEach(processDocument);
    this.rootRecords.set(rootRecords);
    this.fileRecords.set(fileRecords);
    this.processRecords.set(processRecords);
    this.documentsRecords.set(documentsRecords);
  }

  async setAppraisalDecision(recordId: string, decision: AppraisalCode): Promise<void> {
    const appraisals = await firstValueFrom(
      this.appraisalService.setDecision(this.processId, recordId, decision),
    );
    this._setAppraisals(appraisals);
  }

  async setPackaging(recordIds: string[], packaging: PackagingChoice): Promise<void> {
    const data = await firstValueFrom(
      this.packagingService.setPackagingChoice(this.processId, recordIds, packaging),
    );
    this.setPackagingData(data);
  }

  async setAppraisalInternalNote(recordObjectId: string, internalNote: string): Promise<void> {
    const appraisals = await firstValueFrom(
      this.appraisalService.setInternalNote(this.processId, recordObjectId, internalNote),
    );
    this._setAppraisals(appraisals);
  }

  async setAppraisals(
    recordObjectIds: string[],
    decision: AppraisalCode,
    internalNote: string,
  ): Promise<void> {
    const appraisals = await firstValueFrom(
      this.appraisalService.setAppraisals(this.processId, recordObjectIds, decision, internalNote),
    );
    this._setAppraisals(appraisals);
  }

  async finalizeAppraisals(): Promise<void> {
    await firstValueFrom(
      this.messageService.finalizeMessageAppraisal(this.message()!.messageHead.processID),
    );
    this.updateAppraisals();
  }

  private async updateAppraisals(): Promise<void> {
    const appraisals = await firstValueFrom(this.appraisalService.getAppraisals(this.processId));
    this._setAppraisals(appraisals);
  }

  private _setAppraisals(appraisals: Appraisal[]): void {
    let map = Map<string, Appraisal>();
    for (const appraisal of appraisals) {
      map = map.set(appraisal.recordId, appraisal);
    }
    this.appraisals.set(map);
  }

  private setPackagingData(data: PackagingData): void {
    this.packagingChoices.set(data.choices);
    this.packagingDecisions.set(data.decisions);
    this.packagingStats.set(data.stats);
  }
}
