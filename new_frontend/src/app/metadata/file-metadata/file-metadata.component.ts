// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { FileRecordObject, MessageService, RecordObjectAppraisal } from '../../message/message.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss']
})
export class FileMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  fileRecordObject?: FileRecordObject;
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
      fileType: new FormControl<string | null>(null),
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
            return this.messageService.getFileRecordObject(+params['id']);
          }
        )
      ).subscribe(
        (fileRecordObject: FileRecordObject) => {
          console.log(fileRecordObject);
          this.fileRecordObject = fileRecordObject;
          this.form.patchValue({
            recordPlanId: fileRecordObject.generalMetadata?.filePlan?.xdomeaID,
            fileId: fileRecordObject.generalMetadata?.xdomeaID,
            subject: fileRecordObject.generalMetadata?.subject,
            fileType: fileRecordObject.type,
            lifeStart: this.messageService.getDateText(fileRecordObject.lifetime?.start),
            lifeEnd: this.messageService.getDateText(fileRecordObject.lifetime?.end),
            appraisal: this.messageService.getRecordObjectAppraisalByCode(
              fileRecordObject.archiveMetadata?.appraisalCode, this.recordObjectAppraisals!,
            )?.desc,
            appraisalRecomm: this.messageService.getRecordObjectAppraisalByCode(
              fileRecordObject.archiveMetadata?.appraisalRecommCode, this.recordObjectAppraisals!,
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
