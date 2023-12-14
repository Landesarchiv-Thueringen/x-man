// angular
import { AfterViewInit, Component, OnDestroy, Query } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params } from '@angular/router';

// project
import {
  FileRecordObject,
  MessageService,
  RecordObjectConfidentiality,
  RecordObjectAppraisal,
  StructureNode,
} from '../../message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';
import { debounceTime, filter, skip } from 'rxjs/operators';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
})
export class FileMetadataComponent implements AfterViewInit, OnDestroy {
  metadataQuery?: Query;
  metadataSubscription?: Subscription;
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
  }

  ngAfterViewInit(): void {
    // fetch metadata of record object everytime the object ID changes
    this.metadataSubscription = this.route.params
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
          this.saveAppraisalNoteChanges();
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
          this.setMetadata();
        },
      });
    // update metadata if record object changes
    this.messageService
      .watchNodeChanges()
      .pipe(
        filter((changedNode: StructureNode) => {
          return changedNode.id === this.fileRecordObject?.id;
        }),
        switchMap((changedNode: StructureNode) => {
          return this.messageService.getFileRecordObject(changedNode.id);
        })
      )
      .subscribe((fileRecordObject: FileRecordObject) => {
        this.fileRecordObject = fileRecordObject;
        this.setMetadata();
      });
  }

  saveAppraisalNoteChanges(): void {
    if (!this.messageAppraisalComplete) {
      this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value: string | null) => {
        this.setAppraisalNote(value);
      });
    }
  }

  ngOnDestroy(): void {
    this.metadataSubscription?.unsubscribe;
  }

  setMetadata(): void {
    if (
      this.fileRecordObject &&
      this.recordObjectAppraisals &&
      this.recordObjectConfidentialities
    ) {
      let appraisal: string | undefined;
      const appraisalRecomm =
        this.messageService.getRecordObjectAppraisalByCode(
          this.fileRecordObject.archiveMetadata?.appraisalRecommCode,
          this.recordObjectAppraisals
        )?.shortDesc;
      if (this.messageAppraisalComplete) {
        appraisal = this.messageService.getRecordObjectAppraisalByCode(
          this.fileRecordObject.archiveMetadata?.appraisalCode,
          this.recordObjectAppraisals
        )?.shortDesc;
      } else {
        appraisal = this.fileRecordObject.archiveMetadata?.appraisalCode;
      }
      this.form.patchValue({
        recordPlanId: this.fileRecordObject.generalMetadata?.filePlan?.xdomeaID,
        fileId: this.fileRecordObject.generalMetadata?.xdomeaID,
        subject: this.fileRecordObject.generalMetadata?.subject,
        fileType: this.fileRecordObject.type,
        lifeStart: this.messageService.getDateText(
          this.fileRecordObject.lifetime?.start
        ),
        lifeEnd: this.messageService.getDateText(
          this.fileRecordObject.lifetime?.end
        ),
        appraisal: appraisal,
        appraisalRecomm: appraisalRecomm,
        appraisalNote:
          this.fileRecordObject.archiveMetadata?.internalAppraisalNote,
        confidentiality: this.recordObjectConfidentialities.find(
          (c: RecordObjectConfidentiality) =>
            c.code ===
            this.fileRecordObject?.generalMetadata?.confidentialityCode
        )?.shortDesc,
      });
    }
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
            this.messageService.updateStructureNodeForRecordObject(fileRecordObject);
            this.notificationService.show('Bewertung erfolgreich gespeichert');
          },
        });
    }
  }

  setAppraisalNote(note: string | null): void {
    if (this.fileRecordObject) {
      this.messageService
        .setFileRecordObjectAppraisalNote(this.fileRecordObject.id, note)
        .subscribe({
          error: (error: any) => {
            console.error(error);
          },
        });
    }
  }
}
