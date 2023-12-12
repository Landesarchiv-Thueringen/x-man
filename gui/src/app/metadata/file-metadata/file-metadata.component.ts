// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import {
  FileRecordObject,
  MessageService,
  RecordObjectConfidentiality,
  RecordObjectAppraisal,
} from '../../message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';
import { debounceTime, skip } from 'rxjs/operators';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
})
export class FileMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  messageAppraisalComplete?: boolean;
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
      appraisalNote: new FormControl<string | null>(null),
      confidentiality: new FormControl<string | null>(null),
    });
    this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value: string) => {
        this.setAppraisalNote(value);
      });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params
      .pipe(
        switchMap((params: Params) => {
          return this.messageService.getFileRecordObject(params['id']);
        }),
        switchMap((fileRecordObject: FileRecordObject) => {
          this.fileRecordObject = fileRecordObject;
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
            this.fileRecordObject!,
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
    recordObjectAppraisals: RecordObjectAppraisal[],
    recordObjectConfidentialities: RecordObjectConfidentiality[]
  ): void {
    let appraisal: string | undefined;
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      fileRecordObject.archiveMetadata?.appraisalRecommCode,
      recordObjectAppraisals
    )?.shortDesc;
    if (this.messageAppraisalComplete) {
      appraisal = this.messageService.getRecordObjectAppraisalByCode(
        fileRecordObject.archiveMetadata?.appraisalCode,
        recordObjectAppraisals
      )?.shortDesc;
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
      appraisalNote: fileRecordObject.archiveMetadata?.internalAppraisalNote,
      confidentiality: recordObjectConfidentialities.find(
        (c: RecordObjectConfidentiality) =>
          c.code === this.fileRecordObject?.generalMetadata?.confidentialityCode
      )?.shortDesc,
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

  setAppraisalNote(note: string): void {
    if (this.fileRecordObject) {
      this.messageService
        .setFileRecordObjectAppraisalNote(this.fileRecordObject.id, note)
        .subscribe({
          error: (error: any) => {
            console.error(error);
          }
        });
    }
  }
}
