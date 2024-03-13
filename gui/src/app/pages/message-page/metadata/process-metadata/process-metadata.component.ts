import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
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
  MessageService,
  ProcessRecordObject,
  RecordObjectAppraisal,
  StructureNode,
} from '../../../../services/message.service';
import { NotificationService } from '../../../../services/notification.service';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatExpansionModule, MatFormFieldModule, MatInputModule, MatSelectModule],
})
export class ProcessMetadataComponent implements AfterViewInit, OnDestroy {
  urlParameterSubscription?: Subscription;
  messageAppraisalComplete?: boolean;
  processRecordObject?: ProcessRecordObject;
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
      processType: new FormControl<string | null>(null),
      lifeStart: new FormControl<string | null>(null),
      lifeEnd: new FormControl<string | null>(null),
      appraisal: new FormControl<string | null>(null),
      appraisalRecomm: new FormControl<string | null>(null),
      appraisalNote: new FormControl<string | null>(null),
      confidentiality: new FormControl<string | null>(null),
      medium: new FormControl<string | null>(null),
    });
    // fetch metadata of record object every time the object ID changes
    this.urlParameterSubscription = combineLatest([
      this.route.params.pipe(
        switchMap((params: Params) => this.messageService.getProcessRecordObject(params['id'])),
        tap((processRecordObject) => (this.processRecordObject = processRecordObject)),
      ),
      this.messageService.getAppraisalCodelist().pipe(tap((appraisals) => (this.recordObjectAppraisals = appraisals))),
      this.messagePage
        .observeMessage()
        .pipe(tap((message) => (this.messageAppraisalComplete = message.appraisalComplete))),
    ])
      .pipe(takeUntilDestroyed())
      .subscribe(() => this.setMetadata());
    this.registerAppraisalNoteChanges();
    // update metadata if record object changes
    this.messageService
      .watchNodeChanges()
      .pipe(
        filter((changedNode: StructureNode) => {
          return changedNode.id === this.processRecordObject?.id;
        }),
        switchMap((changedNode: StructureNode) => {
          return this.messageService.getProcessRecordObject(changedNode.id);
        }),
      )
      .subscribe((processRecordObject: ProcessRecordObject) => {
        this.processRecordObject = processRecordObject;
        this.setMetadata();
      });
  }

  ngAfterViewInit(): void {}

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
    this.urlParameterSubscription?.unsubscribe;
  }

  setMetadata(): void {
    if (this.processRecordObject && this.recordObjectAppraisals) {
      let appraisal: string | undefined;
      const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
        this.processRecordObject.archiveMetadata?.appraisalRecommCode,
        this.recordObjectAppraisals,
      )?.shortDesc;
      if (this.messageAppraisalComplete) {
        appraisal = this.messageService.getRecordObjectAppraisalByCode(
          this.processRecordObject.archiveMetadata?.appraisalCode,
          this.recordObjectAppraisals,
        )?.shortDesc;
      } else {
        appraisal = this.processRecordObject.archiveMetadata?.appraisalCode;
      }
      this.form.patchValue({
        recordPlanId: this.processRecordObject.generalMetadata?.filePlan?.xdomeaID,
        fileId: this.processRecordObject.generalMetadata?.xdomeaID,
        subject: this.processRecordObject.generalMetadata?.subject,
        processType: this.processRecordObject.type,
        lifeStart: this.messageService.getDateText(this.processRecordObject.lifetime?.start),
        lifeEnd: this.messageService.getDateText(this.processRecordObject.lifetime?.end),
        appraisal: appraisal,
        appraisalRecomm: appraisalRecomm,
        appraisalNote: this.processRecordObject.archiveMetadata?.internalAppraisalNote,
        confidentiality: this.processRecordObject.generalMetadata?.confidentialityLevel?.shortDesc,
        medium: this.processRecordObject.generalMetadata?.medium?.shortDesc,
      });
    }
  }

  setAppraisal(event: any): void {
    if (this.processRecordObject) {
      this.messageService.setProcessRecordObjectAppraisal(this.processRecordObject.id, event.value).subscribe({
        error: (error) => {
          console.error(error);
        },
        next: (processRecordObject: ProcessRecordObject) => {
          this.messageService.updateStructureNodeForRecordObject(processRecordObject);
          this.notificationService.show('Bewertung erfolgreich gespeichert');
        },
      });
    }
  }

  setAppraisalNote(note: string | null): void {
    if (this.processRecordObject) {
      this.messageService.setProcessRecordObjectAppraisalNote(this.processRecordObject.id, note).subscribe({
        error: (error: any) => {
          console.error(error);
        },
      });
    }
  }
}
