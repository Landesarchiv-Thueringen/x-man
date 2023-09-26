// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import {
  FileRecordObject,
  Message,
  MessageService,
  RecordObjectConfidentiality,
  RecordObjectAppraisal,
} from '../../message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
})
export class FileMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  message?: Message;
  fileRecordObject?: FileRecordObject;
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
      fileType: new FormControl<string | null>(null),
      lifeStart: new FormControl<string | null>(null),
      lifeEnd: new FormControl<string | null>(null),
      appraisal: new FormControl<string | null>(null),
      appraisalRecomm: new FormControl<string | null>(null),
      confidentiality: new FormControl<string | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.route.parent?.params.subscribe((params) => {
      this.messageService
        .getMessage(params['id'])
        .subscribe((message: Message) => {
          this.message = message;
        });
    });
    this.urlParameterSubscription = this.route.params
      .pipe(
        switchMap((params: Params) => {
          return this.messageService.getFileRecordObject(params['id']);
        }),
        switchMap((fileRecordObject: FileRecordObject) => {
          console.log(fileRecordObject);
          this.fileRecordObject = fileRecordObject;
          return this.messageService.getMessage(
            this.route.parent!.snapshot.params['id']
          );
        }),
        switchMap((message: Message) => {
          this.message = message;
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
            this.fileRecordObject!,
            this.message!,
            this.recordObjectAppraisals!,
            this.recordObjectConfidentialities
          );
        },
      });
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }

  setMetadata(
    fileRecordObject: FileRecordObject,
    message: Message,
    recordObjectAppraisals: RecordObjectAppraisal[],
    recordObjectConfidentialities: RecordObjectConfidentiality[],
  ): void {
    let appraisal: string | undefined;
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      fileRecordObject.archiveMetadata?.appraisalRecommCode,
      recordObjectAppraisals
    )?.desc;
    if (message.appraisalComplete) {
      appraisal = this.messageService.getRecordObjectAppraisalByCode(
        fileRecordObject.archiveMetadata?.appraisalCode,
        recordObjectAppraisals
      )?.desc;
    } else {
      appraisal = fileRecordObject.archiveMetadata?.appraisalCode;
    }
    this.form.patchValue({
      recordPlanId: fileRecordObject.generalMetadata?.filePlan?.xdomeaID,
      fileId: fileRecordObject.generalMetadata?.xdomeaID,
      subject: fileRecordObject.generalMetadata?.subject,
      fileType: fileRecordObject.type,
      lifeStart: this.messageService.getDateText(
        fileRecordObject.lifetime?.start
      ),
      lifeEnd: this.messageService.getDateText(fileRecordObject.lifetime?.end),
      appraisal: appraisal,
      appraisalRecomm: appraisalRecomm,
      confidentiality: recordObjectConfidentialities.find(
        (c: RecordObjectConfidentiality) =>
          c.code === this.fileRecordObject?.generalMetadata?.confidentialityCode
      )?.desc,
    });
  }

  setAppraisal(event: any): void {
    if (this.fileRecordObject) {
      this.messageService
        .setFileRecordObjectAppraisal(this.fileRecordObject.id, event.value)
        .subscribe({
          error: (error) => {
            console.error(error);
          },
          next: (fileRecordObject: FileRecordObject) => {
            this.messageService.updateStructureNode(fileRecordObject);
            this.notificationService.show('Bewertung erfolgreich gespeichert');
          },
        });
    }
  }
}
