import { DatePipe } from '@angular/common';
import { Component, TemplateRef, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { MatExpansionPanel } from '@angular/material/expansion';
import { ActivatedRoute, Router } from '@angular/router';
import { Observable, debounceTime, distinctUntilChanged, filter, map, of, skip, switchMap, take, tap } from 'rxjs';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { ClearingDetailsComponent } from '../../clearing/clearing-details.component';
import { Message, MessageService } from '../../message/message.service';
import { Process, ProcessService, ProcessStep } from '../../process/process.service';
import { AuthService } from '../../utility/authorization/auth.service';
import { ConfigService } from '../../utility/config.service';

interface StateItem {
  icon: string;
  title: string;
  date: string;
  message?: string;
  onClick?: () => void;
}

@Component({
  selector: 'app-message-metadata',
  templateUrl: './message-metadata.component.html',
  styleUrls: ['./message-metadata.component.scss'],
})
export class MessageMetadataComponent {
  @ViewChild('deleteDialog') deleteDialogTemplate!: TemplateRef<unknown>;

  form = this.formBuilder.group({
    processID: new FormControl<string | null>(null),
    creationTime: new FormControl<string | null>(null),
    xdomeaVersion: new FormControl<string | null>(null),
    note: new FormControl<string | null>(null),
  });
  message?: Message;
  process?: Process;
  stateItems: StateItem[] = [];
  processDeleteTime: Date | null = null;
  isAdmin = this.auth.isAdmin();

  constructor(
    private auth: AuthService,
    private configService: ConfigService,
    private datePipe: DatePipe,
    private dialog: MatDialog,
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notification: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
    private router: Router,
  ) {
    this.route.parent?.params
      .pipe(
        // Get and handle message
        switchMap((params) => this.messageService.getMessage(params['id'])),
        tap((message) => {
          this.message = message;
          this.form.patchValue({
            processID: message.messageHead.processID,
            creationTime: this.datePipe.transform(new Date(message.messageHead.creationTime), 'short'),
            xdomeaVersion: message.xdomeaVersion,
          });
        }),
        // Get and handle process
        switchMap((message) => this.processService.observeProcess(message.messageHead.processID)),
        tap((process) => {
          const isFirstValue = !this.process;
          this.process = process;
          this.stateItems = this.getStateItems();
          if (isFirstValue) {
            this.form.patchValue({ note: process.note });
          }
        }),
        // Get and handle config
        switchMap((process) => this.getProcessDeleteTime(process)),
        tap((processDeleteTime) => (this.processDeleteTime = processDeleteTime)),
        takeUntilDestroyed(),
      )
      .subscribe();
    // Save note when typing (after a debounce)
    this.form.valueChanges
      .pipe(
        map((changes) => changes.note),
        filter((note) => note != null),
        distinctUntilChanged(),
        skip(1), // Skip initial value
        debounceTime(3000),
      )
      .subscribe(() => this.saveNote());
  }

  saveNote(): void {
    const value = this.form.get('note')?.value ?? '';
    if (this.process!.note !== value) {
      this.processService.setNote(this.process!.id, value).subscribe(() => {
        this.process!.note = value;
        this.notification.show('Notiz gespeichert');
      });
    }
  }

  deleteProcess() {
    this.dialog
      .open(this.deleteDialogTemplate)
      .afterClosed()
      .subscribe((confirmed) => {
        if (confirmed) {
          this.processService.deleteProcess(this.process!.id).subscribe(() => {
            this.notification.show('Aussonderung gelöscht');
            this.router.navigate(['/']);
          });
        }
      });
  }

  numberOfUnresolvedErrors(): number {
    return this.process?.processingErrors.filter((processingError) => !processingError.resolved).length ?? 0;
  }

  hasUnresolvedError(): boolean {
    return this.numberOfUnresolvedErrors() > 0;
  }

  isStepRunning(processStep: ProcessStep): boolean {
    return processStep.tasks.some((task) => task.state === 'running');
  }

  scrollToBottom(panel: MatExpansionPanel): void {
    let expanded = false;
    const scrollParent = document.getElementsByTagName('mat-sidenav-content').item(0)!;
    function scroll() {
      scrollParent.scroll({ top: 1000000 });
      if (!expanded) window.requestAnimationFrame(scroll);
    }
    panel.afterExpand.pipe(take(1)).subscribe(() => (expanded = true));
    window.requestAnimationFrame(scroll);
  }

  private getStateItems(): StateItem[] {
    if (!this.process) {
      return [];
    }
    const state = this.process.processState;
    let items: StateItem[] = [];
    if (state.receive0501.complete) {
      items.push({ title: 'Anbietung erhalten', icon: 'check', date: state.receive0501.completionTime! });
    }
    if (state.appraisal.complete) {
      items.push({ title: 'Bewertung abgeschlossen', icon: 'check', date: state.appraisal.completionTime! });
    }
    if (state.receive0505.complete) {
      items.push({ title: 'Bewertung in VIS importiert', icon: 'check', date: state.receive0505.completionTime! });
    }
    if (state.receive0503.complete) {
      items.push({ title: 'Abgabe erhalten', icon: 'check', date: state.receive0503.completionTime! });
    }
    if (state.formatVerification.complete) {
      items.push({
        title: 'Formatverifikation abgeschlossen',
        icon: 'check',
        date: state.formatVerification.completionTime!,
      });
    } else if (this.isStepRunning(state.formatVerification)) {
      const task = state.formatVerification.tasks.find((task) => task.state === 'running')!;
      items.push({
        title: 'Formatverifikation läuft...',
        icon: 'spinner',
        date: task.createdAt,
        message: `${task.itemCompletedCount} / ${task.itemCount}`,
      });
    }
    if (state.archiving.complete) {
      items.push({
        title: 'Abgabe archiviert',
        icon: 'check',
        date: state.archiving.completionTime!,
      });
    } else if (this.isStepRunning(state.archiving)) {
      const task = state.archiving.tasks.find((task) => task.state === 'running')!;
      items.push({
        title: 'Archivierung läuft...',
        icon: 'spinner',
        date: task.createdAt,
      });
    }
    for (const processingError of this.process.processingErrors) {
      let onClick;
      if (this.auth.isAdmin()) {
        onClick = () => this.dialog.open(ClearingDetailsComponent, { data: processingError });
      }
      if (processingError.resolved) {
        items.push({
          title: 'Gelöst: ' + processingError.description,
          icon: 'check_circle',
          date: processingError.detectedAt,
          onClick,
        });
      } else {
        items.push({
          title: processingError.description,
          icon: 'error',
          date: processingError.detectedAt,
          onClick,
        });
      }
    }
    return items.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
  }

  private getProcessDeleteTime(process: Process): Observable<Date | null> {
    if (process.processState.archiving.complete) {
      return this.configService.config.pipe(
        map((config) => {
          let date = new Date(process.processState.archiving.completionTime!);
          date.setDate(date.getDate() + config.deleteArchivedProcessesAfterDays);
          return date;
        }),
      );
    } else {
      return of(null);
    }
  }
}
