import { CommonModule } from '@angular/common';
import { Component, computed, effect, Signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ActivatedRoute } from '@angular/router';
import { debounceTime, skip } from 'rxjs/operators';
import {
  AppraisalCode,
  appraisalDescriptions,
  AppraisalService,
} from '../../../../services/appraisal.service';
import { MessageService } from '../../../../services/message.service';
import { MessagePageService } from '../../message-page.service';
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
  // Signals
  readonly recordId: Signal<string>;
  /** The page's process record. Might update on page changes. */
  readonly record = computed(() => this.messagePage.processRecords().get(this.recordId()));
  readonly appraisal = computed(() => this.messagePage.appraisals().get(this.recordId()));
  readonly appraisalComplete = this.messagePage.appraisalComplete;
  readonly canBeAppraised: Signal<boolean>;
  readonly hasUnresolvedError = this.messagePage.hasUnresolvedError;
  readonly selectionActive = this.messagePage.selectionActive;

  readonly form = this.formBuilder.group({
    recordPlanId: new FormControl<string | null>(null),
    recordPlanSubject: new FormControl<string | null>(null),
    fileId: new FormControl<string | null>(null),
    subject: new FormControl<string | null>(null),
    leadership: new FormControl<string | null>(null),
    fileManager: new FormControl<string | null>(null),
    processType: new FormControl<string | null>(null),
    lifeStart: new FormControl<string | null>(null),
    lifeEnd: new FormControl<string | null>(null),
    appraisal: new FormControl<string | null>(null),
    appraisalRecomm: new FormControl<string | null>(null),
    appraisalNote: new FormControl<string>('', { nonNullable: true }),
    confidentiality: new FormControl<string | null>(null),
    medium: new FormControl<string | null>(null),
  });

  readonly appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({
    code,
    ...d,
  }));

  constructor(
    private appraisalService: AppraisalService,
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    // Define signals.
    const params = toSignal(this.route.params, { requireSync: true });
    this.recordId = computed<string>(() => params()['id']);
    const structureNode = computed(() => this.messagePage.treeNodes().get(this.recordId()));
    this.canBeAppraised = computed(() => structureNode()?.canBeAppraised ?? false);
    // Update the form on changes.
    this.registerRecord();
    this.registerAppraisal();
    // Register inputs.
    this.registerAppraisalNoteChanges();
    // Disable individual appraisal controls while selection is active.
    effect(() => {
      if (this.messagePage.selectionActive() || this.messagePage.hasUnresolvedError()) {
        this.form.get('appraisal')?.disable();
        this.form.get('appraisalNote')?.disable();
      } else {
        this.form.get('appraisal')?.enable();
        this.form.get('appraisalNote')?.enable();
      }
    });
  }

  /** Updates the form when `record` changes. */
  private registerRecord(): void {
    effect(() => {
      const record = this.record();
      const appraisalRecomm = record?.archiveMetadata?.appraisalRecommCode;
      let confidentiality: string | undefined;
      if (record?.generalMetadata?.confidentialityLevel) {
        confidentiality =
          confidentialityLevels[record?.generalMetadata?.confidentialityLevel].shortDesc;
      }
      let medium: string | undefined;
      if (record?.generalMetadata?.medium) {
        medium = media[record?.generalMetadata?.medium].shortDesc;
      }
      this.form.patchValue({
        recordPlanId: record?.generalMetadata?.filePlan?.filePlanNumber,
        recordPlanSubject: record?.generalMetadata?.filePlan?.subject,
        fileId: record?.generalMetadata?.recordNumber,
        subject: record?.generalMetadata?.subject,
        leadership: record?.generalMetadata?.leadership,
        fileManager: record?.generalMetadata?.fileManager,
        processType: record?.type,
        lifeStart: this.messageService.getDateText(record?.lifetime?.start),
        lifeEnd: this.messageService.getDateText(record?.lifetime?.end),
        appraisalRecomm: this.appraisalService.getAppraisalDescription(appraisalRecomm)?.shortDesc,
        confidentiality,
        medium,
      });
    });
  }

  /** Updates the form when `appraisal` or `appraisalComplete` changes. */
  private registerAppraisal(): void {
    effect(() => {
      this.form.patchValue({
        appraisal: this.appraisalComplete()
          ? this.appraisalService.getAppraisalDescription(this.appraisal()?.decision)?.shortDesc
          : this.appraisal()?.decision,
        appraisalNote: this.appraisal()?.note,
      });
    });
  }

  /** Sends the appraisal note to the backend when the value of the form field changes. */
  private registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value) => {
        if (value !== this.appraisal()?.note && !this.appraisalComplete()) {
          this.setAppraisalNote(value);
        }
      });
  }

  setAppraisal(decision: AppraisalCode): void {
    this.messagePage.setAppraisalDecision(this.record()!.recordId, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.record()!.recordId, note);
  }
}
