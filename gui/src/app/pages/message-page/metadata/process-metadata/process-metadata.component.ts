import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ActivatedRoute, Params } from '@angular/router';
import { combineLatest, switchMap } from 'rxjs';
import { debounceTime, shareReplay, skip } from 'rxjs/operators';
import { Appraisal, AppraisalDecision } from '../../../../services/appraisal.service';
import { AppraisalCode, MessageService, ProcessRecordObject } from '../../../../services/message.service';
import { MessagePageService } from '../../message-page.service';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatExpansionModule, MatFormFieldModule, MatInputModule, MatSelectModule],
})
export class ProcessMetadataComponent {
  /** The pages process record object. Might update on page changes. */
  recordObject?: ProcessRecordObject;
  appraisal?: Appraisal | null;
  appraisalCodes: AppraisalCode[] = [];
  appraisalComplete?: boolean;
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
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
      appraisalNote: new FormControl<string>('', { nonNullable: true }),
      confidentiality: new FormControl<string | null>(null),
      medium: new FormControl<string | null>(null),
    });
    const recordObject = this.route.params.pipe(
      switchMap((params: Params) => this.messageService.getProcessRecordObject(params['id'])),
      shareReplay(1),
    );
    const appraisal = recordObject.pipe(
      switchMap((recordObject) => this.messagePage.observeAppraisal(recordObject.xdomeaID)),
    );
    // Update the form and local properties on changes.
    combineLatest([
      recordObject,
      appraisal,
      this.messageService.getAppraisalCodelist(),
      this.messagePage.observeAppraisalComplete(),
    ])
      .pipe(takeUntilDestroyed())
      .subscribe(([recordObject, appraisal, appraisalCodes, appraisalComplete]) =>
        this.setMetadata(recordObject, appraisal, appraisalCodes, appraisalComplete),
      );
    // Send the appraisal note to the backend when the value of the form field changes.
    this.registerAppraisalNoteChanges();
  }

  registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges.pipe(skip(1), debounceTime(400)).subscribe((value) => {
      if (value !== this.appraisal?.internalNote && this.appraisalComplete === false) {
        this.setAppraisalNote(value);
      }
    });
  }

  setMetadata(
    recordObject: ProcessRecordObject,
    appraisal: Appraisal | null,
    appraisalCodes: AppraisalCode[],
    appraisalComplete: boolean,
  ): void {
    this.recordObject = recordObject;
    this.appraisal = appraisal;
    this.appraisalCodes = appraisalCodes;
    this.appraisalComplete = appraisalComplete;
    const appraisalDecision = this.messageService.getRecordObjectAppraisalByCode(appraisal?.decision, appraisalCodes);
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      recordObject.archiveMetadata?.appraisalRecommCode,
      appraisalCodes,
    )?.shortDesc;
    this.form.patchValue({
      recordPlanId: recordObject.generalMetadata?.filePlan?.xdomeaID,
      fileId: recordObject.generalMetadata?.xdomeaID,
      subject: recordObject.generalMetadata?.subject,
      processType: recordObject.type,
      lifeStart: this.messageService.getDateText(recordObject.lifetime?.start),
      lifeEnd: this.messageService.getDateText(recordObject.lifetime?.end),
      appraisal: this.appraisalComplete ? appraisalDecision?.shortDesc : appraisalDecision?.code,
      appraisalRecomm: appraisalRecomm,
      appraisalNote: appraisal?.internalNote,
      confidentiality: recordObject.generalMetadata?.confidentialityLevel?.shortDesc,
      medium: recordObject.generalMetadata?.medium?.shortDesc,
    });
  }

  setAppraisal(decision: AppraisalDecision): void {
    this.messagePage.setAppraisalDecision(this.recordObject!.xdomeaID, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.recordObject!.xdomeaID, note);
  }
}
