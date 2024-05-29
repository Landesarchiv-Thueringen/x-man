import { CommonModule, DatePipe } from '@angular/common';
import { Component, TemplateRef, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
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
import { Observable, debounceTime, distinctUntilChanged, filter, map, of, skip, take } from 'rxjs';
import { AuthService } from '../../../../services/auth.service';
import { ProcessingError } from '../../../../services/clearing.service';
import { ConfigService } from '../../../../services/config.service';
import { Message } from '../../../../services/message.service';
import { NotificationService } from '../../../../services/notification.service';
import { ProcessService, ProcessStep, SubmissionProcess } from '../../../../services/process.service';
import { ClearingDetailsComponent } from '../../../clearing-page/clearing-details.component';
import { MessagePageService } from '../../message-page.service';
import { InstitutMetadataComponent } from '../institution-metadata/institution-metadata.component';

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
  standalone: true,
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
    ReactiveFormsModule,
  ],
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
  process?: SubmissionProcess;
  processingErrors: ProcessingError[] = [];
  stateItems: StateItem[] = [];
  processDeleteTime: Date | null = null;
  isAdmin = this.auth.isAdmin();
  hasUnresolvedError = this.messagePage.hasUnresolvedError;

  constructor(
    private auth: AuthService,
    private configService: ConfigService,
    private datePipe: DatePipe,
    private dialog: MatDialog,
    private formBuilder: FormBuilder,
    private notification: NotificationService,
    private processService: ProcessService,
    private router: Router,
    private messagePage: MessagePageService,
  ) {
    this.messagePage
      .observeProcessData()
      .pipe(takeUntilDestroyed())
      .subscribe((data) => {
        const isFirstValue = !this.process;
        this.process = data.process ?? undefined;
        this.processingErrors = data.processingErrors;
        this.stateItems = this.getStateItems();
        if (isFirstValue) {
          this.form.patchValue({ note: data.process.note });
        }
        this.getProcessDeleteTime(data.process).subscribe(
          (processDeleteTime) => (this.processDeleteTime = processDeleteTime),
        );
      });
    this.messagePage
      .observeMessage()
      .pipe(takeUntilDestroyed())
      .subscribe((message) => {
        this.message = message;
        this.form.patchValue({
          processID: message.messageHead.processID,
          creationTime: this.datePipe.transform(new Date(message.messageHead.creationTime), 'short'),
          xdomeaVersion: message.xdomeaVersion,
        });
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
    if (this.process!.note !== value) {
      this.processService.setNote(this.process!.processId, value).subscribe(() => {
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
          this.processService.deleteProcess(this.process!.processId).subscribe(() => {
            this.notification.show('Aussonderung gelöscht');
            this.router.navigate(['/']);
          });
        }
      });
  }

  numberOfUnresolvedErrors(): number {
    if (!this.process) {
      return 0;
    }
    return Object.values(this.process.processState).reduce((acc, step: ProcessStep) => step.unresolvedErrors + acc, 0);
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
      items.push({ title: 'Anbietung erhalten', icon: 'check', date: state.receive0501.completedAt! });
    }
    if (state.appraisal.complete) {
      items.push({ title: 'Bewertung abgeschlossen', icon: 'check', date: state.appraisal.completedAt! });
    } else if (state.appraisal.progress) {
      items.push({
        title: 'Bewertung',
        icon: 'edit_note',
        message: state.appraisal.progress,
        date: state.appraisal.updatedAt!,
      });
    }
    if (state.receive0505.complete) {
      items.push({ title: 'Bewertung in DMS importiert', icon: 'check', date: state.receive0505.completedAt! });
    }
    if (state.receive0503.complete) {
      items.push({ title: 'Abgabe erhalten', icon: 'check', date: state.receive0503.completedAt! });
    }
    if (state.formatVerification.complete) {
      items.push({
        title: 'Formatverifikation abgeschlossen',
        icon: 'check',
        date: state.formatVerification.completedAt!,
      });
    } else if (state.formatVerification.running) {
      items.push({
        title: 'Formatverifikation läuft...',
        icon: 'spinner',
        date: state.formatVerification.updatedAt,
        message: state.formatVerification.progress,
      });
    }
    if (state.archiving.complete) {
      items.push({
        title: 'Abgabe archiviert',
        icon: 'check',
        date: state.archiving.completedAt!,
      });
    } else if (state.archiving.running) {
      items.push({
        title: 'Archivierung läuft...',
        icon: 'spinner',
        date: state.archiving.updatedAt,
      });
    }
    for (const processingError of this.processingErrors) {
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

  private getProcessDeleteTime(process: SubmissionProcess): Observable<Date | null> {
    if (process.processState.archiving.complete) {
      return this.configService.config.pipe(
        map((config) => {
          let date = new Date(process.processState.archiving.completedAt!);
          date.setDate(date.getDate() + config.deleteArchivedProcessesAfterDays);
          return date;
        }),
      );
    } else {
      return of(null);
    }
  }
}
