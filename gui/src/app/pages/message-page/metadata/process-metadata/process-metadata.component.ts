import { CommonModule } from '@angular/common';
import { Component, effect } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ActivatedRoute, Params } from '@angular/router';
import { combineLatest, switchMap } from 'rxjs';
import { debounceTime, shareReplay, skip } from 'rxjs/operators';
import {
  Appraisal,
  AppraisalCode,
  AppraisalService,
  appraisalDescriptions,
} from '../../../../services/appraisal.service';
import { MessageService } from '../../../../services/message.service';
import { ProcessRecord } from '../../../../services/records.service';
import { MessagePageService } from '../../message-page.service';
import { MessageProcessorService, StructureNode } from '../../message-processor.service';
import { confidentialityLevels } from '../confidentiality-level.pipe';
import { media } from '../medium.pipe';

@Component({
  selector: 'app-process-metadata',
  templateUrl: './process-metadata.component.html',
  styleUrls: ['./process-metadata.component.scss'],
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
})
export class ProcessMetadataComponent {
  /** The pages process record object. Might update on page changes. */
  recordObject?: ProcessRecord;
  appraisal?: Appraisal | null;
  appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({ code, ...d }));
  appraisalComplete?: boolean;
  form: FormGroup;
  canBeAppraised = false;

  constructor(
    private appraisalService: AppraisalService,
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
    private messageProcessor: MessageProcessorService,
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
      switchMap((params: Params) => this.messagePage.getProcessRecord(params['id'])),
      shareReplay(1),
    );
    const structureNode = recordObject.pipe(
      switchMap((recordObject) => this.messageProcessor.getNodeWhenReady(recordObject.recordId)),
    );
    const appraisal = recordObject.pipe(
      switchMap((recordObject) => this.messagePage.observeAppraisal(recordObject.recordId)),
    );
    // Update the form and local properties on changes.
    combineLatest([
      recordObject,
      structureNode,
      appraisal,
      this.messagePage.observeAppraisalComplete(),
    ])
      .pipe(takeUntilDestroyed())
      .subscribe(([recordObject, structureNode, appraisal, appraisalComplete]) =>
        this.setMetadata(recordObject, structureNode, appraisal, appraisalComplete),
      );
    // Send the appraisal note to the backend when the value of the form field changes.
    this.registerAppraisalNoteChanges();
    // Disable individual appraisal controls while selection is active.
    effect(() => {
      if (this.messagePage.showSelection() || this.messagePage.hasUnresolvedError()) {
        this.form.get('appraisal')?.disable();
        this.form.get('appraisalNote')?.disable();
      } else {
        this.form.get('appraisal')?.enable();
        this.form.get('appraisalNote')?.enable();
      }
    });
  }

  registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value) => {
        if (value !== this.appraisal?.note && this.appraisalComplete === false) {
          this.setAppraisalNote(value);
        }
      });
  }

  setMetadata(
    recordObject: ProcessRecord,
    structureNode: StructureNode,
    appraisal: Appraisal | null,
    appraisalComplete: boolean,
  ): void {
    this.recordObject = recordObject;
    this.canBeAppraised = structureNode.canBeAppraised;
    this.appraisal = appraisal;
    this.appraisalComplete = appraisalComplete;
    const appraisalRecomm = recordObject.archiveMetadata?.appraisalRecommCode;
    let confidentiality: string | undefined;
    if (recordObject.generalMetadata?.confidentialityLevel) {
      confidentiality =
        confidentialityLevels[recordObject.generalMetadata?.confidentialityLevel].shortDesc;
    }
    let medium: string | undefined;
    if (recordObject.generalMetadata?.medium) {
      medium = media[recordObject.generalMetadata?.medium].shortDesc;
    }
    this.form.patchValue({
      recordPlanId: recordObject.generalMetadata?.filePlan?.filePlanNumber,
      fileId: recordObject.generalMetadata?.recordNumber,
      subject: recordObject.generalMetadata?.subject,
      processType: recordObject.type,
      lifeStart: this.messageService.getDateText(recordObject.lifetime?.start),
      lifeEnd: this.messageService.getDateText(recordObject.lifetime?.end),
      appraisal: this.appraisalComplete
        ? this.appraisalService.getAppraisalDescription(appraisal?.decision)?.shortDesc
        : appraisal?.decision,
      appraisalRecomm: this.appraisalService.getAppraisalDescription(appraisalRecomm)?.shortDesc,
      appraisalNote: appraisal?.note,
      confidentiality,
      medium,
    });
  }

  setAppraisal(decision: AppraisalCode): void {
    this.messagePage.setAppraisalDecision(this.recordObject!.recordId, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.recordObject!.recordId, note);
  }
}
