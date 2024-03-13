import { Injectable } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import { BehaviorSubject, Observable, distinctUntilChanged, filter, first, map, of, switchMap, take } from 'rxjs';
import { Appraisal, AppraisalService } from '../../services/appraisal.service';
import { Message, MessageService } from '../../services/message.service';
import { Process, ProcessService } from '../../services/process.service';
import { notNull } from '../../utils/predicates';

/**
 * A Service to provide data the the message page and its child components.
 */
@Injectable()
export class MessagePageService {
  /**
   * The process references by the page URL.
   *
   * We update the process regularly by refetching it from the backend.
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

  private processId!: string;

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
  private registerProcessAndAppraisals() {
    this.route.params.pipe(take(1)).subscribe((params) => {
      this.processId = params['processId'];
      this.appraisalService
        .observeAppraisals(this.processId)
        .pipe(takeUntilDestroyed())
        .subscribe((appraisals) => this.appraisals.next(appraisals));
      this.processService
        .observeProcess(this.processId)
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

  getProcessId(): string {
    return this.processId;
  }

  getProcess(): Observable<Process> {
    return this.observeProcess().pipe(first());
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
}
