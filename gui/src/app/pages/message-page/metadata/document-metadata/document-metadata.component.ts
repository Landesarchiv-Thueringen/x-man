import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ActivatedRoute, Params } from '@angular/router';
import { Subscription, switchMap } from 'rxjs';
import { DocumentRecordObject, MessageService } from '../../../../services/message.service';
import { DocumentVersionMetadataComponent } from '../document-version-metadata/document-version-metadata.component';

@Component({
  selector: 'app-document-metadata',
  templateUrl: './document-metadata.component.html',
  styleUrls: ['./document-metadata.component.scss'],
  standalone: true,
  imports: [
    CommonModule,
    DocumentVersionMetadataComponent,
    MatExpansionModule,
    MatInputModule,
    MatFormFieldModule,
    ReactiveFormsModule,
  ],
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
      )
      .subscribe((document: DocumentRecordObject) => {
        this.documentRecordObject = document;
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
