import { CommonModule } from '@angular/common';
import { Component, computed, effect, Signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ActivatedRoute } from '@angular/router';
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
    imports: [
        CommonModule,
        DocumentVersionMetadataComponent,
        MatExpansionModule,
        MatInputModule,
        MatFormFieldModule,
        ReactiveFormsModule,
    ]
})
export class DocumentMetadataComponent {
  readonly record: Signal<DocumentRecord | undefined>;
  readonly form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    this.form = this.formBuilder.group({
      recordPlanId: new FormControl<string | null>(null),
      recordPlanSubject: new FormControl<string | null>(null),
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
    const params = toSignal(this.route.params, { requireSync: true });
    const recordId = computed<string>(() => params()['id']);
    this.record = computed(() => this.messagePage.documentsRecords().get(recordId()));
    effect(() => this.setRecord(this.record()));
  }

  private setRecord(record?: DocumentRecord): void {
    let confidentiality: string | undefined;
    if (record?.generalMetadata?.confidentialityLevel) {
      confidentiality =
        confidentialityLevels[record?.generalMetadata?.confidentialityLevel].shortDesc;
    }
    let medium: string | undefined;
    if (record?.generalMetadata?.medium) {
      medium = media[record?.generalMetadata?.medium].shortDesc;
    }
    this.form.patchValue({
      recordPlanId: record?.generalMetadata?.filePlan?.filePlanNumber,
      recordPlanSubject: record?.generalMetadata?.filePlan?.subject,
      fileId: record?.generalMetadata?.recordNumber,
      subject: record?.generalMetadata?.subject,
      documentType: record?.type,
      incomingDate: this.messageService.getDateText(record?.incomingDate),
      outgoingDate: this.messageService.getDateText(record?.outgoingDate),
      documentDate: this.messageService.getDateText(record?.documentDate),
      confidentiality,
      medium,
    });
  }
}
