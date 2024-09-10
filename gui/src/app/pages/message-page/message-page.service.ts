import { Injectable, WritableSignal, computed, signal } from '@angular/core';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import {
  BehaviorSubject,
  Observable,
  distinctUntilChanged,
  filter,
  first,
  firstValueFrom,
  map,
  switchMap,
  take,
} from 'rxjs';
import { Appraisal, AppraisalCode, AppraisalService } from '../../services/appraisal.service';
import { Message, MessageService, MessageType } from '../../services/message.service';
import {
  PackagingData,
  PackagingDecision,
  PackagingOption,
  PackagingService,
  PackagingStats,
} from '../../services/packaging.service';
import { ProcessData, ProcessService, SubmissionProcess } from '../../services/process.service';
import {
  DocumentRecord,
  FileRecord,
  ProcessRecord,
  Records,
  RecordsService,
} from '../../services/records.service';
import { notEmpty, notNull } from '../../utils/predicates';

/**
 * A Service to provide data the the message page and its child components.
 *
 * It's lifetime is linked to that of the message-page component.
 */
@Injectable()
export class MessagePageService {
  /**
   * The process references by the page URL.
   *
   * We update the process regularly by refetching it from the backend, however,
   * the process ID should not change for the lifetime of the message page.
   */
  private processData = new BehaviorSubject<ProcessData | null>(null);
  process = computed(() => this.processDataSignal()?.process);
  /**
   * The message references by the page URL.
   *
   * We fetch the message once when the relevant URL parameter changes.
   */
  private _message = new BehaviorSubject<Message | undefined>(undefined);
  message = toSignal(this._message);
  /**
   * The message's root records.
   *
   * We fetch the root records when the message changes.
   */
  private _rootRecords = new BehaviorSubject<Records | undefined>(undefined);
  rootRecords = toSignal(this._rootRecords);
  /**
   * Record maps of root and all nested records.
   *
   * We update all record maps when the root records change.
   */
  private fileRecordsMap = new Map<string, FileRecord>();
  private processRecordsMap = new Map<string, ProcessRecord>();
  private documentsRecordsMap = new Map<string, DocumentRecord>();
  /**
   * All appraisals for the current process.
   *
   * Appraisals can be updated by the user at any time.
   */
  private appraisals = new BehaviorSubject<Appraisal[] | null>(null);
  packagingOptions: WritableSignal<{ [recordId in string]?: PackagingOption }> = signal({});
  packagingDecisions: WritableSignal<{ [recordId in string]?: PackagingDecision }> = signal({});
  packagingStats: WritableSignal<{ [recordId in string]?: PackagingStats }> = signal({});

  readonly selectionActive = signal(false);
  private readonly processDataSignal = toSignal(this.processData);
  readonly hasUnresolvedError = computed(() => {
    const data = this.processDataSignal();
    if (!data) {
      return 0;
    }
    return data.process.unresolvedErrors > 0;
  });

  constructor(
    private appraisalService: AppraisalService,
    private messageService: MessageService,
    private processService: ProcessService,
    private recordsService: RecordsService,
    private packagingService: PackagingService,
    private route: ActivatedRoute,
  ) {
    const processId = this.route.params.pipe(
      take(1),
      map((params) => params['processId']),
    );
    processId.subscribe((processId) => {
      this.registerMessage(processId);
      // Fetch appraisals and record options once, will be updated when changed
      // by other functions.
      this.appraisalService
        .getAppraisals(processId)
        .subscribe((appraisals) => this.appraisals.next(appraisals ?? []));
      // Observe process until destroyed and update `this.process`.
      this.processService
        .observeProcessData(processId)
        .pipe(takeUntilDestroyed())
        .subscribe((data) => {
          this.processData.next(data);
        });
    });
  }

  /**
   * Fetches the message and updates `this.message`.
   */
  private registerMessage(processId: string) {
    const messageType: Observable<MessageType> = this.route.params.pipe(
      map((params) => params['messageType']),
      filter(notEmpty),
      distinctUntilChanged(),
    );
    messageType
      .pipe(switchMap((messageType) => this.messageService.getMessage(processId, messageType)))
      .subscribe((message) => this._message.next(message));
    messageType
      .pipe(switchMap((messageType) => this.recordsService.getRootRecords(processId, messageType)))
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
  private updateRecords(rootRecords: Records) {
    this.fileRecordsMap.clear();
    this.processRecordsMap.clear();
    this.documentsRecordsMap.clear();
    const processDocument = (document: DocumentRecord) => {
      this.documentsRecordsMap.set(document.recordId, document);
      document.attachments?.forEach(processDocument);
    };
    const processProcess = (process: ProcessRecord) => {
      this.processRecordsMap.set(process.recordId, process);
      process.subprocesses?.forEach(processProcess);
      process.documents?.forEach(processDocument);
    };
    const processFile = (file: FileRecord) => {
      this.fileRecordsMap.set(file.recordId, file);
      file.subfiles?.forEach(processFile);
      file.processes?.forEach(processProcess);
    };
    rootRecords.files?.forEach(processFile);
    rootRecords.processes?.forEach(processProcess);
    rootRecords.documents?.forEach(processDocument);
    this._rootRecords.next(rootRecords);
  }

  getProcess(): Observable<SubmissionProcess> {
    return this.processData.pipe(
      first(notNull),
      map(({ process }) => process),
    );
  }

  observeProcessData(): Observable<ProcessData> {
    return this.processData.pipe(filter(notNull));
  }

  observeMessage(): Observable<Message> {
    return this._message.pipe(filter(notNull));
  }

  observeRootRecords(): Observable<Records> {
    return this._rootRecords.pipe(filter(notNull));
  }

  getFileRecord(recordId: string): Observable<FileRecord> {
    return this._rootRecords.pipe(
      map(() => this.fileRecordsMap.get(recordId)),
      filter(notNull),
      take(1),
    );
  }

  getProcessRecord(recordId: string): Observable<ProcessRecord> {
    return this._rootRecords.pipe(
      map(() => this.processRecordsMap.get(recordId)),
      filter(notNull),
      take(1),
    );
  }

  getDocumentRecord(recordId: string): Observable<DocumentRecord> {
    return this._rootRecords.pipe(
      map(() => this.documentsRecordsMap.get(recordId)),
      filter(notNull),
      take(1),
    );
  }

  observeAppraisal(recordId: string): Observable<Appraisal | null> {
    return this.observeAppraisals().pipe(
      map((appraisals) => appraisals.find((a) => a.recordId === recordId) ?? null),
    );
  }

  observeAppraisals(): Observable<Appraisal[]> {
    return this.appraisals.pipe(filter(notNull));
  }

  observeAppraisalComplete(): Observable<boolean> {
    return this.observeProcessData().pipe(
      map(
        ({ process }) =>
          process.processState.appraisal.complete || process.processState.receive0503.complete,
      ),
      distinctUntilChanged(),
    );
  }

  async setAppraisalDecision(recordId: string, decision: AppraisalCode): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(
      this.appraisalService.setDecision(process.processId, recordId, decision),
    );
    this.appraisals.next(appraisals);
  }

  async setPackaging(recordIds: string[], packaging: PackagingOption): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const data = await firstValueFrom(
      this.packagingService.setPackaging(process.processId, recordIds, packaging),
    );
    this.setPackagingData(data);
  }

  async setAppraisalInternalNote(recordObjectId: string, internalNote: string): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(
      this.appraisalService.setInternalNote(process.processId, recordObjectId, internalNote),
    );
    this.appraisals.next(appraisals);
  }

  async setAppraisals(
    recordObjectIds: string[],
    decision: AppraisalCode,
    internalNote: string,
  ): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(
      this.appraisalService.setAppraisals(
        process.processId,
        recordObjectIds,
        decision,
        internalNote,
      ),
    );
    this.appraisals.next(appraisals);
  }

  async finalizeAppraisals(): Promise<void> {
    await firstValueFrom(
      this.messageService.finalizeMessageAppraisal(this._message.value!.messageHead.processID),
    );
    this.updateAppraisals();
    // FIXME: We should rather do a genuine update of the process object.
    this.processData.value!.process.processState.appraisal.complete = true;
  }

  private async updateAppraisals(): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(this.appraisalService.getAppraisals(process.processId));
    this.appraisals.next(appraisals);
  }

  private setPackagingData(data: PackagingData): void {
    this.packagingOptions.set(data.packagingOptions);
    this.packagingDecisions.set(data.packagingDecisions);
    this.packagingStats.set(data.packagingStats);
  }
}
