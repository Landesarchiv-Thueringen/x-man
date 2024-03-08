import { Injectable } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute } from '@angular/router';
import { BehaviorSubject, Observable, distinctUntilChanged, filter, first, map, of, switchMap, take } from 'rxjs';
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

  constructor(
    route: ActivatedRoute,
    private processService: ProcessService,
    private messageService: MessageService,
  ) {
    // Regularly fetch the process and update `this.process`.
    route.params.pipe(take(1)).subscribe((params) => {
      const processId = params['processId'];
      this.processService
        .observeProcess(processId)
        .pipe(takeUntilDestroyed())
        .subscribe((process) => this.process.next(process));
    });
    // Fetch the message and update `this.message`.
    route.params
      .pipe(
        map((params) => params['messageCode']),
        filter((messageCode) => messageCode != ''),
        distinctUntilChanged(),
        switchMap((messageCode) =>
          this.process.pipe(
            filter(notNull),
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
        distinctUntilChanged(),
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
    return this.observeProcess().pipe(first());
  }

  observeProcess(): Observable<Process> {
    return this.process.pipe(filter(notNull));
  }

  observeMessage(): Observable<Message> {
    return this.message.pipe(filter(notNull));
  }
}
