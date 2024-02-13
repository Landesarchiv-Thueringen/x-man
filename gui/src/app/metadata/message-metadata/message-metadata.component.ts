import { DatePipe } from '@angular/common';
import { Component, TemplateRef, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { ActivatedRoute, Router } from '@angular/router';
import { Observable, map, of, switchMap, tap } from 'rxjs';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { Message, MessageService } from '../../message/message.service';
import { Process, ProcessService } from '../../process/process.service';
import { AuthService } from '../../utility/authorization/auth.service';
import { ConfigService } from '../../utility/config.service';

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
        switchMap((message) => this.processService.observeProcessByXdomeaID(message.messageHead.processID)),
        tap((process) => {
          this.process = process;
          this.form.patchValue({ note: process.note });
        }),
        // Get and handle config
        switchMap((process) => this.getProcessDeleteTime(process)),
        tap((processDeleteTime) => (this.processDeleteTime = processDeleteTime)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  saveNote(): void {
    const value = this.form.get('note')?.value ?? '';
    if (this.process!.note !== value) {
      this.processService.setNote(this.process!.xdomeaID, value).subscribe(() => {
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
          this.processService.deleteProcess(this.process!.xdomeaID).subscribe(() => {
            this.notification.show('Aussonderung gelöscht');
            this.router.navigate(['/']);
          });
        }
      });
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
