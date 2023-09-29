// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import {
  DocumentRecordObject,
  MessageService,
  RecordObjectConfidentiality,
} from '../../message/message.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-document-metadata',
  templateUrl: './document-metadata.component.html',
  styleUrls: ['./document-metadata.component.scss'],
})
export class DocumentMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  documentRecordObject?: DocumentRecordObject;
  recordObjectConfidentialities?: RecordObjectConfidentiality[];
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private route: ActivatedRoute
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
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.pipe(
      switchMap((params: Params) => {
        return this.messageService.getDocumentRecordObject(params['id'])
      }),
      switchMap((document: DocumentRecordObject) => {
        console.log(document);
        this.documentRecordObject = document;
        return this.messageService.getRecordObjectConfidentialities();
      }),
    ).subscribe(
      (confidentialities: RecordObjectConfidentiality[]) => {
        this.recordObjectConfidentialities = confidentialities;
        this.form.patchValue({
          recordPlanId:
          this.documentRecordObject!.generalMetadata?.filePlan?.xdomeaID,
          fileId: this.documentRecordObject!.generalMetadata?.xdomeaID,
          subject: this.documentRecordObject!.generalMetadata?.subject,
          documentType: this.documentRecordObject!.type,
          incomingDate: this.messageService.getDateText(
            this.documentRecordObject!.incomingDate
          ),
          outgoingDate: this.messageService.getDateText(
            this.documentRecordObject!.outgoingDate
          ),
          documentDate: this.messageService.getDateText(
            this.documentRecordObject!.documentDate
          ),
          confidentiality: this.recordObjectConfidentialities.find(
            (c: RecordObjectConfidentiality) =>
              c.code === this.documentRecordObject?.generalMetadata?.confidentialityCode
          )?.desc,
        });
      }
    );
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
