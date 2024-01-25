import { DatePipe } from '@angular/common';
import { Component } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { Message, MessageService } from '../../message/message.service';
import { Process, ProcessService } from '../../process/process.service';

@Component({
  selector: 'app-message-metadata',
  templateUrl: './message-metadata.component.html',
  styleUrls: ['./message-metadata.component.scss'],
})
export class MessageMetadataComponent {
  form = this.formBuilder.group({
    processID: new FormControl<string | null>(null),
    creationTime: new FormControl<string | null>(null),
    xdomeaVersion: new FormControl<string | null>(null),
    note: new FormControl<string | null>(null),
  });
  message?: Message;
  process?: Process;

  constructor(
    private datePipe: DatePipe,
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notification: NotificationService,
    private processService: ProcessService,
    private route: ActivatedRoute,
  ) {
    this.route.parent?.params.pipe(takeUntilDestroyed()).subscribe((params) => {
      this.messageService.getMessage(params['id']).subscribe((message: Message) => {
        this.message = message;
        this.form.patchValue({
          processID: message.messageHead.processID,
          creationTime: this.datePipe.transform(new Date(message.messageHead.creationTime), 'short'),
          xdomeaVersion: message.xdomeaVersion,
        });
        this.processService.getProcessByXdomeaID(message.messageHead.processID).subscribe((process) => {
          this.process = process;
          this.form.patchValue({
            note: process.note,
          });
        });
      });
    });
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
}
