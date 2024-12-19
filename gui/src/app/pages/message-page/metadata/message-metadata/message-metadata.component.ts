import { CommonModule, DatePipe } from '@angular/common';
import { Component, computed, effect, inject, Signal, TemplateRef, viewChild } from '@angular/core';
import { FormBuilder, FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule, MatExpansionPanel } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { Router } from '@angular/router';
import { debounceTime, distinctUntilChanged, filter, map, skip, take } from 'rxjs';
import { AuthService } from '../../../../services/auth.service';
import { Config, ConfigService } from '../../../../services/config.service';
import { MessageService } from '../../../../services/message.service';
import { NotificationService } from '../../../../services/notification.service';
import { ProcessService, SubmissionProcess } from '../../../../services/process.service';
import { ItemProgress, TaskState } from '../../../../services/tasks.service';
import { TaskStateIconComponent } from '../../../../shared/task-state-icon.component';
import { TaskDetailsComponent } from '../../../admin-page/tasks/task-details.component';
import { ClearingDetailsComponent } from '../../../clearing-page/clearing-details.component';
import { MessagePageService } from '../../message-page.service';
import { InstitutMetadataComponent } from '../institution-metadata/institution-metadata.component';
import { ProcessStepProgressPipe } from './process-step-progress.pipe';

interface StateItem {
  icon: string;
  title: string;
  date: string;
  class?: string;
  taskState?: TaskState;
  progress?: ItemProgress;
  onClick?: () => void;
}

@Component({
  selector: 'app-message-metadata',
  templateUrl: './message-metadata.component.html',
  styleUrls: ['./message-metadata.component.scss'],
  imports: [
    CommonModule,
    InstitutMetadataComponent,
    MatButtonModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatListModule,
    MatProgressSpinnerModule,
    TaskStateIconComponent,
    ProcessStepProgressPipe,
    ReactiveFormsModule,
  ],
})
export class MessageMetadataComponent {
  private auth = inject(AuthService);
  private configService = inject(ConfigService);
  private datePipe = inject(DatePipe);
  private dialog = inject(MatDialog);
  private formBuilder = inject(FormBuilder);
  private notification = inject(NotificationService);
  private processService = inject(ProcessService);
  private router = inject(Router);
  private messagePage = inject(MessagePageService);
  private messageService = inject(MessageService);

  readonly reimportMessageDialogTemplate =
    viewChild.required<TemplateRef<unknown>>('reimportMessageDialog');
  readonly deleteMessageDialogTemplate =
    viewChild.required<TemplateRef<unknown>>('deleteMessageDialog');
  readonly deleteSubmissionProcessDialogTemplate = viewChild.required<TemplateRef<unknown>>(
    'deleteSubmissionProcessDialog',
  );

  readonly process = this.messagePage.process;
  readonly warnings = this.messagePage.warnings;
  readonly processingErrors = this.messagePage.processingErrors;
  readonly message = this.messagePage.message;
  readonly hasUnresolvedError = this.messagePage.hasUnresolvedError;
  readonly isAdmin = this.auth.isAdmin();
  readonly processDeleteTime: Signal<Date | undefined>;

  readonly form = this.formBuilder.group({
    processID: new FormControl<string | null>(null),
    creationTime: new FormControl<string | null>(null),
    xdomeaVersion: new FormControl<string | null>(null),
    note: new FormControl<string | null>(null),
  });

  stateItems: StateItem[] = [];

  constructor() {
    this.processDeleteTime = computed(() =>
      this.getProcessDeleteTime(this.configService.config(), this.process()),
    );
    let initialized = false;
    effect(() => {
      const process = this.process();
      if (process) {
        this.stateItems = this.getStateItems();
        if (!initialized) {
          this.form.patchValue({ note: process.note });
        }
        initialized = true;
      }
    });
    effect(() => {
      const message = this.message();
      if (message) {
        this.form.patchValue({
          processID: message.messageHead.processID,
          creationTime: this.datePipe.transform(
            new Date(message.messageHead.creationTime),
            'short',
          ),
          xdomeaVersion: message.xdomeaVersion,
        });
      }
    });
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
    if (this.process()!.note !== value) {
      this.processService.setNote(this.process()!.processId, value).subscribe(() => {
        this.process()!.note = value;
        this.notification.show('Notiz gespeichert');
      });
    }
  }

  reimportMessage() {
    this.dialog
      .open(this.reimportMessageDialogTemplate())
      .afterClosed()
      .subscribe((confirmed) => {
        if (confirmed) {
          this.messageService
            .reimportMessage(this.process()!.processId, this.message()!.messageType)
            .subscribe(() => {
              this.notification.show('Nachricht wird neu eingelesen...');
            });
        }
      });
  }

  deleteMessage() {
    this.dialog
      .open(this.deleteMessageDialogTemplate())
      .afterClosed()
      .subscribe((confirmed) => {
        if (confirmed) {
          this.messageService
            .deleteMessage(this.process()!.processId, this.message()!.messageType)
            .subscribe(() => {
              this.notification.show('Nachricht gelöscht');
              const processHasOtherMessage =
                this.process()?.processState?.receive0501?.complete &&
                this.process()?.processState?.receive0503?.complete;
              if (!processHasOtherMessage) {
                this.router.navigate(['/']);
              }
            });
        }
      });
  }

  deleteProcess() {
    this.dialog
      .open(this.deleteSubmissionProcessDialogTemplate())
      .afterClosed()
      .subscribe((confirmed) => {
        if (confirmed) {
          this.processService.deleteProcess(this.process()!.processId).subscribe(() => {
            this.notification.show('Aussonderung gelöscht');
            this.router.navigate(['/']);
          });
        }
      });
  }

  numberOfUnresolvedErrors(): number {
    return this.process()?.unresolvedErrors ?? 0;
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
    if (!this.process()) {
      return [];
    }
    const state = this.process()!.processState;
    let items: StateItem[] = [];
    if (state.receive0501.complete) {
      items.push({
        title: 'Anbietung erhalten',
        icon: 'check',
        date: state.receive0501.completedAt!,
      });
    }
    if (state.appraisal.complete) {
      items.push({
        title: 'Bewertung abgeschlossen',
        icon: 'check',
        date: state.appraisal.completedAt!,
      });
    } else if (state.appraisal.progress) {
      items.push({
        title: 'Bewertung',
        icon: 'edit_note',
        progress: state.appraisal.progress,
        date: state.appraisal.updatedAt!,
      });
    }
    if (state.receive0505.complete) {
      items.push({
        title: 'Bewertung in DMS importiert',
        icon: 'check',
        date: state.receive0505.completedAt!,
      });
    }
    if (state.receive0503.complete) {
      items.push({ title: 'Abgabe erhalten', icon: 'check', date: state.receive0503.completedAt! });
    }
    let onClick = () =>
      this.dialog.open(TaskDetailsComponent, {
        data: state.formatVerification.taskId,
        width: '1000px',
        maxWidth: '80vw',
      });
    if (state.formatVerification.complete) {
      items.push({
        title: 'Formatverifikation abgeschlossen',
        icon: 'check',
        date: state.formatVerification.completedAt!,
        onClick,
      });
    } else if (
      state.formatVerification.taskState &&
      state.formatVerification.taskState !== 'failed'
      // The failed state is already displayed by a processing error.
    ) {
      items.push({
        title: 'Formatverifikation läuft...',
        icon: this.getTaskIcon(state.formatVerification.taskState),
        date: state.formatVerification.updatedAt,
        taskState: state.formatVerification.taskState,
        progress: state.formatVerification.progress,
        onClick,
      });
    }
    onClick = () =>
      this.dialog.open(TaskDetailsComponent, {
        data: state.archiving.taskId,
        width: '1000px',
        maxWidth: '80vw',
      });
    if (state.archiving.complete) {
      items.push({
        title: 'Abgabe archiviert',
        icon: 'check',
        date: state.archiving.completedAt!,
        onClick,
      });
    } else if (state.archiving.taskState && state.archiving.taskState !== 'failed') {
      items.push({
        title: 'Archivierung läuft...',
        icon: this.getTaskIcon(state.archiving.taskState),
        date: state.archiving.updatedAt,
        taskState: state.archiving.taskState,
        progress: state.archiving.progress,
        onClick,
      });
    }
    for (const warning of this.warnings()) {
      items.push({
        title: warning.title,
        icon: 'warning',
        class: 'warning',
        date: warning.createdAt,
      });
    }
    for (const processingError of this.processingErrors()) {
      let onClick;
      if (this.auth.isAdmin()) {
        onClick = () =>
          this.dialog.open(ClearingDetailsComponent, {
            maxWidth: '80vw',
            data: processingError,
          });
      }
      if (processingError.resolved) {
        items.push({
          title: 'Gelöst: ' + processingError.title,
          icon: 'check_circle',
          class: 'solved',
          date: processingError.createdAt,
          onClick,
        });
      } else {
        items.push({
          title: processingError.title,
          icon: 'error',
          date: processingError.createdAt,
          onClick,
        });
      }
    }
    return items.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
  }

  private getProcessDeleteTime(config?: Config, process?: SubmissionProcess): Date | undefined {
    if (config && process?.processState.archiving.complete) {
      let date = new Date(process.processState.archiving.completedAt!);
      date.setDate(date.getDate() + config.deleteArchivedProcessesAfterDays);
      return date;
    } else {
      return undefined;
    }
  }

  private getTaskIcon(state: TaskState): string {
    switch (state) {
      case 'pending':
        return 'schedule';
      case 'running':
      case 'pausing':
        return 'spinner';
      case 'paused':
        return 'spinner-stopped';
      default:
        return '';
    }
  }
}
