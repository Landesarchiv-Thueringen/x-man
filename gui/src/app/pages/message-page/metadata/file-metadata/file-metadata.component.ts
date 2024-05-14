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
import {
  Appraisal,
  AppraisalCode,
  AppraisalService,
  appraisalDescriptions,
} from '../../../../services/appraisal.service';
import { MessageService } from '../../../../services/message.service';
import { FileRecord } from '../../../../services/records.service';
import { MessagePageService } from '../../message-page.service';
import { MessageProcessorService, StructureNode } from '../../message-processor.service';
import { confidentialityLevels } from '../confidentiality-level.pipe';
import { media } from '../medium.pipe';

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
  record?: FileRecord;
  appraisal?: Appraisal | null;
  appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({ code, ...d }));
  appraisalComplete?: boolean;
  form: FormGroup;
  canBeAppraised = false;

  constructor(
    private appraisalService: AppraisalService,
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
    const record = this.route.params.pipe(
      switchMap((params: Params) => this.messagePage.getFileRecord(params['id'])),
      shareReplay(1),
    );
    const structureNode = record.pipe(switchMap((record) => this.messageProcessor.getNodeWhenReady(record.recordId)));
    const appraisal = record.pipe(switchMap((record) => this.messagePage.observeAppraisal(record.recordId)));
    // Update the form and local properties on changes.
    combineLatest([record, structureNode, appraisal, this.messagePage.observeAppraisalComplete()])
      .pipe(takeUntilDestroyed())
      .subscribe(([recordObject, structureNode, appraisal, appraisalComplete]) =>
        this.setMetadata(recordObject, structureNode, appraisal, appraisalComplete),
      );
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
    this.form.controls['appraisalNote'].valueChanges.pipe(skip(1), debounceTime(400)).subscribe((value) => {
      if (value !== this.appraisal?.note && this.appraisalComplete === false) {
        this.setAppraisalNote(value);
      }
    });
  }

  setMetadata(
    record: FileRecord,
    structureNode: StructureNode,
    appraisal: Appraisal | null,
    appraisalComplete: boolean,
  ): void {
    this.record = record;
    this.canBeAppraised = structureNode.canBeAppraised;
    this.appraisal = appraisal;
    this.appraisalComplete = appraisalComplete;
    const appraisalRecomm = this.record.archiveMetadata?.appraisalRecommCode;
    let confidentiality: string | undefined;
    if (record.generalMetadata?.confidentialityLevel) {
      confidentiality = confidentialityLevels[record.generalMetadata?.confidentialityLevel].shortDesc;
    }
    let medium: string | undefined;
    if (record.generalMetadata?.medium) {
      medium = media[record.generalMetadata?.medium].shortDesc;
    }
    this.form.patchValue({
      recordPlanId: this.record.generalMetadata?.filePlan?.filePlanNumber,
      fileId: this.record.generalMetadata?.recordNumber,
      subject: this.record.generalMetadata?.subject,
      fileType: this.record.type,
      lifeStart: this.messageService.getDateText(this.record.lifetime?.start),
      lifeEnd: this.messageService.getDateText(this.record.lifetime?.end),
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
    this.messagePage.setAppraisalDecision(this.record!.recordId, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.record!.recordId, note);
  }
}
