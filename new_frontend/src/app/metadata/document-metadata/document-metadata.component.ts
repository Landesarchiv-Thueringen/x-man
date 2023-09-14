// angular
import { AfterViewInit, Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { DocumentRecordObject, MessageService } from '../../message/message.service';

// utility
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-document-metadata',
  templateUrl: './document-metadata.component.html',
  styleUrls: ['./document-metadata.component.scss']
})
export class DocumentMetadataComponent implements AfterViewInit {
  urlParameterSubscription?: Subscription;
  documentRecordObject?: DocumentRecordObject;
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
      appraisal: new FormControl<number | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.subscribe((params) => {
      this.messageService.getDocumentRecordObject(+params['id']).subscribe(
        (documentRecordObject: DocumentRecordObject) => {
          console.log(documentRecordObject);
          this.documentRecordObject = documentRecordObject;
          this.form.patchValue({
            recordPlanId: documentRecordObject.generalMetadata?.filePlan?.xdomeaID,
            fileId: documentRecordObject.generalMetadata?.xdomeaID,
            subject: documentRecordObject.generalMetadata?.subject,
            documentType: documentRecordObject.type,
          });
        }
      )
    })   
  }
}
