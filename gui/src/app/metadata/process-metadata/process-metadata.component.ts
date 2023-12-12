// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import {
  ProcessRecordObject,
  MessageService,
  RecordObjectAppraisal,
  RecordObjectConfidentiality,
} from '../../message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';
import { debounceTime, skip } from 'rxjs/operators';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss'],
})
export class ProcessMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  messageAppraisalComplete?: boolean;
  processRecordObject?: ProcessRecordObject;
  recordObjectAppraisals?: RecordObjectAppraisal[];
  recordObjectConfidentialities?: RecordObjectConfidentiality[];
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute
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
      appraisalNote: new FormControl<string | null>(null),
      confidentiality: new FormControl<string | null>(null),
    });
    this.form.controls['appraisalNote'].valueChanges
    .pipe(
      skip(1),
      debounceTime(400),
    )
    .subscribe((value: string|null) => {
      this.setAppraisalNote(value);
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params
      .pipe(
        switchMap((params: Params) => {
          return this.messageService.getProcessRecordObject(params['id']);
        }),
        switchMap((processRecordObject: ProcessRecordObject) => {
          this.processRecordObject = processRecordObject;
          return this.messageService.isMessageAppraisalComplete(
            this.route.parent!.snapshot.params['id']
          );
        }),
        switchMap((messageAppraisalComplete: boolean) => {
          this.messageAppraisalComplete = messageAppraisalComplete;
          return this.messageService.getRecordObjectAppraisals();
        }),
        switchMap((appraisals: RecordObjectAppraisal[]) => {
          this.recordObjectAppraisals = appraisals;
          return this.messageService.getRecordObjectConfidentialities();
        })
      )
      .subscribe({
        error: (error: any) => {
          console.error(error);
        },
        next: (confidentialities: RecordObjectConfidentiality[]) => {
          this.recordObjectConfidentialities = confidentialities;
          this.setMetadata(
            this.processRecordObject!,
            this.recordObjectAppraisals!,
            this.recordObjectConfidentialities,
          );
        },
      });
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }

  setMetadata(
    processRecordObject: ProcessRecordObject,
    recordObjectAppraisals: RecordObjectAppraisal[],
    recordObjectConfidentialities: RecordObjectConfidentiality[],
  ): void {
    let appraisal: string | undefined;
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      processRecordObject.archiveMetadata?.appraisalRecommCode,
      recordObjectAppraisals
    )?.shortDesc;
    if (this.messageAppraisalComplete) {
      appraisal = this.messageService.getRecordObjectAppraisalByCode(
        processRecordObject.archiveMetadata?.appraisalCode,
        recordObjectAppraisals
      )?.shortDesc;
    } else {
      appraisal = processRecordObject.archiveMetadata?.appraisalCode;
    }
    this.form.patchValue({
      recordPlanId: processRecordObject.generalMetadata?.filePlan?.xdomeaID,
      fileId: processRecordObject.generalMetadata?.xdomeaID,
      subject: processRecordObject.generalMetadata?.subject,
      processType: processRecordObject.type,
      lifeStart: this.messageService.getDateText(
        processRecordObject.lifetime?.start
      ),
      lifeEnd: this.messageService.getDateText(
        processRecordObject.lifetime?.end
      ),
      appraisal: appraisal,
      appraisalRecomm: appraisalRecomm,
      appraisalNote: processRecordObject.archiveMetadata?.internalAppraisalNote,
      confidentiality: recordObjectConfidentialities.find(
        (c: RecordObjectConfidentiality) =>
          c.code === this.processRecordObject?.generalMetadata?.confidentialityCode
      )?.desc,
    });
  }

  setAppraisal(event: any): void {
    if (this.processRecordObject) {
      this.messageService
        .setProcessRecordObjectAppraisal(
          this.processRecordObject.id,
          event.value
        )
        .subscribe({
          error: (error) => {
            console.error(error);
          },
          next: (processRecordObject: ProcessRecordObject) => {
            this.messageService.updateStructureNode(processRecordObject);
            this.notificationService.show('Bewertung erfolgreich gespeichert');
          },
        });
    }
  }

  setAppraisalNote(note: string|null): void {
    if (this.processRecordObject) {
      this.messageService
        .setProcessRecordObjectAppraisalNote(this.processRecordObject.id, note)
        .subscribe({
          error: (error: any) => {
            console.error(error);
          },
        });
    }
  }
}
