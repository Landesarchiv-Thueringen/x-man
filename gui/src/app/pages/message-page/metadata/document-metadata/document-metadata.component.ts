import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ActivatedRoute, Params } from '@angular/router';
import { Subscription, switchMap } from 'rxjs';
import { MessageService } from '../../../../services/message.service';
import { DocumentRecord } from '../../../../services/records.service';
import { MessagePageService } from '../../message-page.service';
import { confidentialityLevels } from '../confidentiality-level.pipe';
import { DocumentVersionMetadataComponent } from '../document-version-metadata/document-version-metadata.component';
import { media } from '../medium.pipe';

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
export class DocumentMetadataComponent {
  urlParameterSubscription?: Subscription;
  documentRecordObject?: DocumentRecord;
  messageTypeCode?: string;
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
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
    this.route.params
      .pipe(switchMap((params: Params) => this.messagePage.getDocumentRecord(params['id'])))
      .subscribe((record) => this.setRecord(record));
  }

  private setRecord(record: DocumentRecord): void {
    this.documentRecordObject = record;
    let confidentiality: string | undefined;
    if (record.generalMetadata?.confidentialityLevel) {
      confidentiality = confidentialityLevels[record.generalMetadata?.confidentialityLevel].shortDesc;
    }
    let medium: string | undefined;
    if (record.generalMetadata?.medium) {
      medium = media[record.generalMetadata?.medium].shortDesc;
    }
    this.form.patchValue({
      recordPlanId: this.documentRecordObject!.generalMetadata?.filePlan?.filePlanNumber,
      fileId: this.documentRecordObject!.generalMetadata?.recordNumber,
      subject: this.documentRecordObject!.generalMetadata?.subject,
      documentType: this.documentRecordObject!.type,
      incomingDate: this.messageService.getDateText(this.documentRecordObject!.incomingDate),
      outgoingDate: this.messageService.getDateText(this.documentRecordObject!.outgoingDate),
      documentDate: this.messageService.getDateText(this.documentRecordObject!.documentDate),
      confidentiality,
      medium,
    });
  }
}
