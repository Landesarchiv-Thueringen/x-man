import { Injectable, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import {
  BehaviorSubject,
  Observable,
  distinctUntilChanged,
  filter,
  first,
  firstValueFrom,
  map,
  of,
  switchMap,
  take,
} from 'rxjs';
import { Appraisal, AppraisalDecision, AppraisalService } from '../../services/appraisal.service';
import { Message, MessageService } from '../../services/message.service';
import { Process, ProcessService } from '../../services/process.service';
import { notNull } from '../../utils/predicates';

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
  private process = new BehaviorSubject<Process | null>(null);
  /**
   * The message references by the page URL.
   *
   * We fetch the message once when the relevant URL parameter changes.
   */
  private message = new BehaviorSubject<Message | null>(null);
  /**
   * All appraisals for the current process.
   *
   * Appraisals can be updated by the user at any time.
   */
  private appraisals = new BehaviorSubject<Appraisal[] | null>(null);

  readonly showSelection = signal(false);

  constructor(
    private route: ActivatedRoute,
    private processService: ProcessService,
    private messageService: MessageService,
    private appraisalService: AppraisalService,
  ) {
    this.registerProcessAndAppraisals();
    this.registerMessage();
  }

  /** Regularly fetches the process and updates `this.process`.  */
  private async registerProcessAndAppraisals() {
    this.route.params.pipe(take(1)).subscribe((params) => {
      const processId = params['processId'];
      // Fetch appraisals once, will be updated when changed by other functions.
      this.appraisalService.getAppraisals(processId).subscribe((appraisals) => this.appraisals.next(appraisals));
      // Observe process until destroyed.
      this.processService
        .observeProcess(processId)
        .pipe(takeUntilDestroyed())
        .subscribe((process) => this.process.next(process));
    });
  }

  /**
   * Fetches the message and updates `this.message`.
   */
  private registerMessage() {
    this.route.params
      .pipe(
        map((params) => params['messageCode']),
        filter((messageCode) => messageCode != ''),
        distinctUntilChanged(),
        switchMap((messageCode) =>
          this.getProcess().pipe(
            map((process) => {
              switch (messageCode) {
                case '0501':
                  return process.message0501Id;
                case '0503':
                  return process.message0503Id;
                default:
                  return null;
              }
            }),
          ),
        ),
        switchMap((messageId) => {
          if (messageId) {
            return this.messageService.getMessage(messageId);
          } else {
            return of(null);
          }
        }),
      )
      .subscribe((message) => this.message.next(message));
  }

  getProcess(): Observable<Process> {
    return this.process.pipe(first(notNull));
  }

  observeProcess(): Observable<Process> {
    return this.process.pipe(filter(notNull));
  }

  observeMessage(): Observable<Message> {
    return this.message.pipe(filter(notNull));
  }

  observeAppraisal(recordObjectId: string): Observable<Appraisal | null> {
    return this.observeAppraisals().pipe(
      map((appraisals) => appraisals.find((a) => a.recordObjectID === recordObjectId) ?? null),
    );
  }

  observeAppraisals(): Observable<Appraisal[]> {
    return this.appraisals.pipe(filter(notNull));
  }

  observeAppraisalComplete(): Observable<boolean> {
    return this.observeProcess().pipe(
      map((process) => process.processState.appraisal.complete),
      distinctUntilChanged(),
    );
  }

  async setAppraisalDecision(recordObjectId: string, decision: AppraisalDecision): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(this.appraisalService.setDecision(process.id, recordObjectId, decision));
    this.appraisals.next(appraisals);
  }

  async setAppraisalInternalNote(recordObjectId: string, internalNote: string): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(
      this.appraisalService.setInternalNote(process.id, recordObjectId, internalNote),
    );
    this.appraisals.next(appraisals);
  }

  async setAppraisals(recordObjectIds: string[], decision: AppraisalDecision, internalNote: string): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(
      this.appraisalService.setAppraisals(process.id, recordObjectIds, decision, internalNote),
    );
    this.appraisals.next(appraisals);
  }

  async finalizeAppraisals(): Promise<void> {
    await firstValueFrom(this.messageService.finalizeMessageAppraisal(this.message.value!.id));
    this.updateAppraisals();
    // FIXME: We should rather do a genuine update of the process object.
    this.process.value!.processState.appraisal.complete = true;
  }

  private async updateAppraisals(): Promise<void> {
    const process = await firstValueFrom(this.getProcess());
    const appraisals = await firstValueFrom(this.appraisalService.getAppraisals(process.id));
    this.appraisals.next(appraisals);
  }
}
