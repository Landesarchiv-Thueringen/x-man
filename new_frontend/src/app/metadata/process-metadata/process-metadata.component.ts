// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import { ProcessRecordObject, Message, MessageService, RecordObjectAppraisal } from '../../message/message.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss']
})
export class ProcessMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  message?: Message;
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
    this.route.parent?.params.subscribe((params) => {
      this.messageService.getMessage(+params['id']).subscribe(
        (message: Message) => {
          this.message = message;
        }
      );
    });
    this.urlParameterSubscription = this.route.params.pipe(
      switchMap(
        (params: Params) => {
          return this.messageService.getProcessRecordObject(+params['id']);
        }
      ),
      switchMap(
        (processRecordObject: ProcessRecordObject) => {
          console.log(processRecordObject);
          this.processRecordObject = processRecordObject;
          return this.messageService.getMessage(this.route.parent!.snapshot.params['id']);
        }
      ),
      switchMap(
        (message: Message) => {
          this.message = message;
          return this.messageService.getRecordObjectAppraisals();
        }
      ),
    ).subscribe({
      error: (error: any) => {
        console.error(error)
      },
      next: (appraisals: RecordObjectAppraisal[]) => {
        this.recordObjectAppraisals = appraisals;
        this.setMetadata(
          this.processRecordObject!, 
          this.message!, 
          this.recordObjectAppraisals,
        );
      }
    });
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }

  setMetadata(
    processRecordObject: ProcessRecordObject, 
    message: Message, 
    recordObjectAppraisals: RecordObjectAppraisal[],
  ): void {
    let appraisal: string | undefined;
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      processRecordObject.archiveMetadata?.appraisalRecommCode, recordObjectAppraisals,
    )?.desc;
    if (message.appraisalComplete) {
      appraisal = this.messageService.getRecordObjectAppraisalByCode(
        processRecordObject.archiveMetadata?.appraisalCode, recordObjectAppraisals,
      )?.desc;
    } else {
      appraisal = processRecordObject.archiveMetadata?.appraisalCode;
    }
    this.form.patchValue({
      recordPlanId: processRecordObject.generalMetadata?.filePlan?.xdomeaID,
      fileId: processRecordObject.generalMetadata?.xdomeaID,
      subject: processRecordObject.generalMetadata?.subject,
      processType: processRecordObject.type,
      lifeStart: this.messageService.getDateText(processRecordObject.lifetime?.start),
      lifeEnd: this.messageService.getDateText(processRecordObject.lifetime?.end),
      appraisal: appraisal,
      appraisalRecomm: appraisalRecomm,
    });
  }
}
