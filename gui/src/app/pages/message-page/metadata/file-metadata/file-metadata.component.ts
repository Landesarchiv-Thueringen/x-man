import { CommonModule } from '@angular/common';
import { Component, Query, effect } from '@angular/core';
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
import { AppraisalCode, FileRecordObject, MessageService } from '../../../../services/message.service';
import { MessagePageService } from '../../message-page.service';
import { MessageProcessorService, StructureNode } from '../../message-processor.service';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatExpansionModule, MatFormFieldModule, MatInputModule, MatSelectModule],
})
export class FileMetadataComponent {
  metadataQuery?: Query;
  /** The pages file record object. Might update on page changes. */
  recordObject?: FileRecordObject;
  appraisal?: Appraisal | null;
  appraisalCodes?: AppraisalCode[];
  appraisalComplete?: boolean;
  form: FormGroup;
  canBeAppraised = false;

  constructor(
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
    private messageService: MessageService,
    private messageProcessor: MessageProcessorService,
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
      appraisalNote: new FormControl<string | null>(null),
      confidentiality: new FormControl<string | null>(null),
      medium: new FormControl<string | null>(null),
    });
    const recordObject = this.route.params.pipe(
      switchMap((params: Params) => this.messageService.getFileRecordObject(params['id'])),
      shareReplay(1),
    );
    const structureNode = recordObject.pipe(
      switchMap((recordObject) => this.messageProcessor.getNodeWhenReady(recordObject.id)),
    );
    const appraisal = recordObject.pipe(
      switchMap((recordObject) => this.messagePage.observeAppraisal(recordObject.xdomeaID)),
    );
    // Update the form and local properties on changes.
    combineLatest([
      recordObject,
      structureNode,
      appraisal,
      this.messageService.getAppraisalCodelist(),
      this.messagePage.observeAppraisalComplete(),
    ])
      .pipe(takeUntilDestroyed())
      .subscribe(([recordObject, structureNode, appraisal, appraisalCodes, appraisalComplete]) =>
        this.setMetadata(recordObject, structureNode, appraisal, appraisalCodes, appraisalComplete),
      );
    this.registerAppraisalNoteChanges();
    // Disable individual appraisal controls while selection is active.
    effect(() => {
      if (this.messagePage.showSelection()) {
        this.form.get('appraisal')?.disable();
        this.form.get('appraisalNote')?.disable();
      } else {
        this.form.get('appraisal')?.enable();
        this.form.get('appraisalNote')?.enable();
      }
    });
  }

  registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges.pipe(skip(1), debounceTime(400)).subscribe((value) => {
      if (value !== this.appraisal?.internalNote && this.appraisalComplete === false) {
        this.setAppraisalNote(value);
      }
    });
  }

  setMetadata(
    recordObject: FileRecordObject,
    structureNode: StructureNode,
    appraisal: Appraisal | null,
    appraisalCodes: AppraisalCode[],
    appraisalComplete: boolean,
  ): void {
    this.recordObject = recordObject;
    this.canBeAppraised = structureNode.canBeAppraised;
    this.appraisal = appraisal;
    this.appraisalCodes = appraisalCodes;
    this.appraisalComplete = appraisalComplete;
    const appraisalDecision = this.messageService.getRecordObjectAppraisalByCode(appraisal?.decision, appraisalCodes);
    const appraisalRecomm = this.messageService.getRecordObjectAppraisalByCode(
      this.recordObject.archiveMetadata?.appraisalRecommCode,
      this.appraisalCodes,
    )?.shortDesc;
    this.form.patchValue({
      recordPlanId: this.recordObject.generalMetadata?.filePlan?.xdomeaID,
      fileId: this.recordObject.generalMetadata?.xdomeaID,
      subject: this.recordObject.generalMetadata?.subject,
      fileType: this.recordObject.type,
      lifeStart: this.messageService.getDateText(this.recordObject.lifetime?.start),
      lifeEnd: this.messageService.getDateText(this.recordObject.lifetime?.end),
      appraisal: this.appraisalComplete ? appraisalDecision?.shortDesc : appraisalDecision?.code,
      appraisalRecomm: appraisalRecomm,
      appraisalNote: appraisal?.internalNote,
      confidentiality: this.recordObject.generalMetadata?.confidentialityLevel?.shortDesc,
      medium: this.recordObject.generalMetadata?.medium?.shortDesc,
    });
  }

  setAppraisal(decision: AppraisalDecision): void {
    this.messagePage.setAppraisalDecision(this.recordObject!.xdomeaID, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.recordObject!.xdomeaID, note);
  }
}
