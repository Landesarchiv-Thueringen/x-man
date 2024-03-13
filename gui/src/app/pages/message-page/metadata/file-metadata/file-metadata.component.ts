import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, OnDestroy, Query } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ActivatedRoute, Params } from '@angular/router';
import { Subscription, combineLatest, switchMap } from 'rxjs';
import { debounceTime, filter, skip, tap } from 'rxjs/operators';
import {
  FileRecordObject,
  MessageService,
  RecordObjectAppraisal,
  StructureNode,
} from '../../../../services/message.service';
import { NotificationService } from '../../../../services/notification.service';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatExpansionModule, MatFormFieldModule, MatInputModule, MatSelectModule],
})
export class FileMetadataComponent implements AfterViewInit, OnDestroy {
  metadataQuery?: Query;
  metadataSubscription?: Subscription;
  messageAppraisalComplete?: boolean;
  fileRecordObject?: FileRecordObject;
  recordObjectAppraisals?: RecordObjectAppraisal[];
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute,
    private messagePage: MessagePageService,
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
      medium: new FormControl<string | null>(null),
    });
    this.messagePage
      .observeMessage()
      .pipe(takeUntilDestroyed())
      .subscribe((message) => {
        this.messageAppraisalComplete = message.appraisalComplete;
      });
    this.registerAppraisalNoteChanges();
  }

  ngAfterViewInit(): void {
    // fetch metadata of record object every time the object ID changes
    this.metadataSubscription = combineLatest([
      this.route.params.pipe(
        switchMap((params: Params) => this.messageService.getFileRecordObject(params['id'])),
        tap((fileRecordObject) => (this.fileRecordObject = fileRecordObject)),
      ),
      this.messageService.getAppraisalCodelist().pipe(tap((appraisals) => (this.recordObjectAppraisals = appraisals))),
    ]).subscribe(() => this.setMetadata());

    // update metadata if record object changes
    this.messageService
      .watchNodeChanges()
      .pipe(
        filter((changedNode: StructureNode) => {
          return changedNode.id === this.fileRecordObject?.id;
        }),
        switchMap((changedNode: StructureNode) => {
          return this.messageService.getFileRecordObject(changedNode.id);
        }),
      )
      .subscribe((fileRecordObject: FileRecordObject) => {
        this.fileRecordObject = fileRecordObject;
        this.setMetadata();
      });
  }

  registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value: string | null) => {
        if (this.messageAppraisalComplete === false) {
          this.setAppraisalNote(value);
        }
      });
  }

  ngOnDestroy(): void {
    this.metadataSubscription?.unsubscribe;
  }

  setMetadata(): void {
    if (this.fileRecordObject && this.recordObjectAppraisals) {
      let appraisal: string | undefined;
      const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
        this.fileRecordObject.archiveMetadata?.appraisalRecommCode,
        this.recordObjectAppraisals,
      )?.shortDesc;
      if (this.messageAppraisalComplete) {
        appraisal = this.messageService.getRecordObjectAppraisalByCode(
          this.fileRecordObject.archiveMetadata?.appraisalCode,
          this.recordObjectAppraisals,
        )?.shortDesc;
      } else {
        appraisal = this.fileRecordObject.archiveMetadata?.appraisalCode;
      }
      this.form.patchValue({
        recordPlanId: this.fileRecordObject.generalMetadata?.filePlan?.xdomeaID,
        fileId: this.fileRecordObject.generalMetadata?.xdomeaID,
        subject: this.fileRecordObject.generalMetadata?.subject,
        fileType: this.fileRecordObject.type,
        lifeStart: this.messageService.getDateText(this.fileRecordObject.lifetime?.start),
        lifeEnd: this.messageService.getDateText(this.fileRecordObject.lifetime?.end),
        appraisal: appraisal,
        appraisalRecomm: appraisalRecomm,
        appraisalNote: this.fileRecordObject.archiveMetadata?.internalAppraisalNote,
        confidentiality: this.fileRecordObject.generalMetadata?.confidentialityLevel?.shortDesc,
        medium: this.fileRecordObject.generalMetadata?.medium?.shortDesc,
      });
    }
  }

  setAppraisal(event: any): void {
    if (this.fileRecordObject) {
      this.messageService.setFileRecordObjectAppraisal(this.fileRecordObject.id, event.value).subscribe({
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
      this.messageService.setFileRecordObjectAppraisalNote(this.fileRecordObject.id, note).subscribe({
        error: (error: any) => {
          console.error(error);
        },
      });
    }
  }
}
