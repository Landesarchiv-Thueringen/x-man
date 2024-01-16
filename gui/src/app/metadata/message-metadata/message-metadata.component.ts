// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { DatePipe } from '@angular/common';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { Message, MessageService } from '../../message/message.service';

// utility
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-message-metadata',
  templateUrl: './message-metadata.component.html',
  styleUrls: ['./message-metadata.component.scss'],
})
export class MessageMetadataComponent implements AfterViewInit, OnDestroy {
  form: FormGroup;
  message?: Message;
  urlParameterSubscription?: Subscription;

  constructor(
    private datePipe: DatePipe,
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    this.form = this.formBuilder.group({
      processID: new FormControl<string | null>(null),
      creationTime: new FormControl<Date | null>(null),
      xdomeaVersion: new FormControl<string | null>(null),
    });
  }

  ngAfterViewInit(): void {
    if (!!this.route.parent) {
      this.urlParameterSubscription = this.route.parent.params.subscribe((params) => {
        this.messageService.getMessage(params['id']).subscribe((message: Message) => {
          this.message = message;
          this.form.patchValue({
            processID: message.messageHead.processID,
            creationTime: this.datePipe.transform(new Date(message.messageHead.creationTime), 'short'),
            xdomeaVersion: message.xdomeaVersion,
          });
        });
      });
    }
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
