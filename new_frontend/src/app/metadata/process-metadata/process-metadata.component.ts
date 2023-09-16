// angular
import { AfterViewInit, Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { ProcessRecordObject, MessageService } from '../../message/message.service';

// utility
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss']
})
export class ProcessMetadataComponent {
  urlParameterSubscription?: Subscription;
  processRecordObject?: ProcessRecordObject;
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
      processType: new FormControl<string | null>(null),
      lifeStart: new FormControl<string | null>(null),
      lifeEnd: new FormControl<string | null>(null),
      appraisal: new FormControl<string | null>(null),
      appraisalRecomm: new FormControl<string | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.subscribe((params) => {
      this.messageService.getProcessRecordObject(+params['id']).subscribe(
        (processRecordObject: ProcessRecordObject) => {
          console.log(processRecordObject);
          this.processRecordObject = processRecordObject;
          this.form.patchValue({
            recordPlanId: processRecordObject.generalMetadata?.filePlan?.xdomeaID,
            fileId: processRecordObject.generalMetadata?.xdomeaID,
            subject: processRecordObject.generalMetadata?.subject,
            processType: processRecordObject.type,
            lifeStart: this.messageService.getDateText(processRecordObject.lifetime?.start),
            lifeEnd: this.messageService.getDateText(processRecordObject.lifetime?.end),
            appraisal: processRecordObject.archiveMetadata?.appraisalCode,
            appraisalRecomm: processRecordObject.archiveMetadata?.appraisalRecommCode,
          });
        }
      )
    })   
  }
}
