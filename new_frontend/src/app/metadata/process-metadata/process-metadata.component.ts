// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { ProcessRecordObject, MessageService, RecordObjectAppraisal } from '../../message/message.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss']
})
export class ProcessMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  processRecordObject?: ProcessRecordObject;
  recordObjectAppraisals?: RecordObjectAppraisal[];
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
      this.messageService.getRecordObjectAppraisals().pipe(
        switchMap(
          (appraisals: RecordObjectAppraisal[]) => {
            this.recordObjectAppraisals = appraisals;
            console.log(appraisals);
            return this.messageService.getProcessRecordObject(+params['id']);
          }
        )
      ).subscribe(
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
            appraisal: this.messageService.getRecordObjectAppraisalByCode(
              processRecordObject.archiveMetadata?.appraisalCode, this.recordObjectAppraisals!,
            )?.desc,
            appraisalRecomm: this.messageService.getRecordObjectAppraisalByCode(
              processRecordObject.archiveMetadata?.appraisalRecommCode, this.recordObjectAppraisals!,
            )?.desc,
          });
        }
      );
    }) 
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
