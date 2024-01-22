import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';
import { Subscription, switchMap } from 'rxjs';
import { DocumentRecordObject, MessageService } from '../../message/message.service';

@Component({
  selector: 'app-document-metadata',
  templateUrl: './document-metadata.component.html',
  styleUrls: ['./document-metadata.component.scss'],
})
export class DocumentMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  documentRecordObject?: DocumentRecordObject;
  messageTypeCode?: string;
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    this.form = this.formBuilder.group({
      recordPlanId: new FormControl<string | null>(null),
      fileId: new FormControl<string | null>(null),
      subject: new FormControl<string | null>(null),
      documentType: new FormControl<string | null>(null),
      incomingDate: new FormControl<string | null>(null),
      outgoingDate: new FormControl<string | null>(null),
      documentDate: new FormControl<string | null>(null),
      appraisal: new FormControl<number | null>(null),
      confidentiality: new FormControl<string | null>(null),
      medium: new FormControl<string | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params
      .pipe(
        switchMap((params: Params) => {
          return this.messageService.getDocumentRecordObject(params['id']);
        }),
        switchMap((document: DocumentRecordObject) => {
          console.log(document);
          this.documentRecordObject = document;
          return this.messageService.getMessageTypeCode(document.messageID);
        }),
      )
      .subscribe((messageTypeCode: string) => {
        this.messageTypeCode = messageTypeCode;
        this.form.patchValue({
          recordPlanId: this.documentRecordObject!.generalMetadata?.filePlan?.xdomeaID,
          fileId: this.documentRecordObject!.generalMetadata?.xdomeaID,
          subject: this.documentRecordObject!.generalMetadata?.subject,
          documentType: this.documentRecordObject!.type,
          incomingDate: this.messageService.getDateText(this.documentRecordObject!.incomingDate),
          outgoingDate: this.messageService.getDateText(this.documentRecordObject!.outgoingDate),
          documentDate: this.messageService.getDateText(this.documentRecordObject!.documentDate),
          confidentiality: this.documentRecordObject?.generalMetadata?.confidentialityLevel?.shortDesc,
          medium: this.documentRecordObject?.generalMetadata?.medium?.shortDesc,
        });
      });
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
