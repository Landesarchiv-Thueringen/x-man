import { CommonModule } from '@angular/common';
import { Component, Signal, computed, effect, inject, resource } from '@angular/core';
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
  AppraisalService,
  appraisalDescriptions,
} from '../../../../services/appraisal.service';
import { MessageService } from '../../../../services/message.service';
import {
  PackagingChoice,
  PackagingService,
  packagingChoices,
} from '../../../../services/packaging.service';
import { MessagePageService } from '../../message-page.service';
import { printPackagingStats } from '../../packaging-stats.pipe';
import { confidentialityLevels } from '../confidentiality-level.pipe';
import { media } from '../medium.pipe';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
})
export class FileMetadataComponent {
  private appraisalService = inject(AppraisalService);
  private formBuilder = inject(FormBuilder);
  private messagePage = inject(MessagePageService);
  private messageService = inject(MessageService);
  private packagingService = inject(PackagingService);
  private route = inject(ActivatedRoute);

  // Signals
  private params = toSignal(this.route.params, { requireSync: true });
  readonly process = this.messagePage.process;
  readonly message = this.messagePage.message;
  readonly recordId = computed<string>(() => this.params()['id']);
  /** The page's file record. Might update on page changes. */
  readonly record = computed(() => this.messagePage.fileRecords().get(this.recordId()));
  readonly appraisal = computed(() => this.messagePage.appraisals().get(this.recordId()));
  readonly appraisalComplete = this.messagePage.appraisalComplete;
  readonly canBeAppraised: Signal<boolean>;
  readonly canChoosePackaging: Signal<boolean>;
  readonly hasUnresolvedError = this.messagePage.hasUnresolvedError;
  readonly packagingDecision = computed(
    () => this.messagePage.packagingDecisions()[this.recordId()] ?? '',
  );
  readonly selectionActive = this.messagePage.selectionActive;

  readonly form = this.formBuilder.group({
    recordPlanId: new FormControl<string | null>(null),
    recordPlanSubject: new FormControl<string | null>(null),
    fileId: new FormControl<string | null>(null),
    subject: new FormControl<string | null>(null),
    leadership: new FormControl<string | null>(null),
    fileManager: new FormControl<string | null>(null),
    fileType: new FormControl<string | null>(null),
    lifeStart: new FormControl<string | null>(null),
    lifeEnd: new FormControl<string | null>(null),
    appraisal: new FormControl<string | null>(null),
    appraisalRecomm: new FormControl<string | null>(null),
    appraisalNote: new FormControl<string>('', { nonNullable: true }),
    confidentiality: new FormControl<string | null>(null),
    medium: new FormControl<string | null>(null),
    packaging: new FormControl<PackagingChoice | null>(null),
  });

  readonly appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({
    code,
    ...d,
  }));

  /** Whether the packaging select box is visible and enabled. */
  packagingEnabled = computed(
    () =>
      this.canChoosePackaging() &&
      !this.selectionActive() &&
      !this.hasUnresolvedError() &&
      this.process()?.processState.archiving.progress == null,
  );
  /**
   * Packaging stats for the current record.
   *
   * We use these stats to enrich available choices of the packaging select box
   * with statistics values. We don't fetch stats when the packaging select box
   * is disabled since we already have the required data to enrich the currently
   * selected choice.
   */
  private packagingStats = resource({
    request: () => (this.packagingEnabled() ? this.recordId() : undefined),
    loader: async ({ request: recordId }) =>
      this.packagingService.getPackagingStats(this.messagePage.process()!.processId, [recordId]),
  });
  /** Packaging choices enriched with stats values. */
  packagingChoices = computed(() => this.getEnrichPackagingChoices());

  constructor() {
    // Define signals.
    const structureNode = computed(() => this.messagePage.treeNodes().get(this.recordId()));
    this.canBeAppraised = computed(() => structureNode()?.canBeAppraised ?? false);
    this.canChoosePackaging = computed(
      () =>
        (this.messagePage.messageType() === '0503' && structureNode()?.canChoosePackaging) ?? false,
    );
    // Update the form on changes.
    this.registerRecord();
    this.registerAppraisal();
    // Register inputs.
    this.registerAppraisalNoteChanges();
    // Disable individual appraisal controls while selection is active.
    effect(() => {
      if (
        !this.appraisalComplete() && // If the appraisal is complete, appraisal fields are readonly anyway.
        (this.selectionActive() || this.hasUnresolvedError())
      ) {
        this.form.get('appraisal')?.disable();
        this.form.get('appraisalNote')?.disable();
      } else {
        this.form.get('appraisal')?.enable();
        this.form.get('appraisalNote')?.enable();
      }
    });
    // Set the packaging select field to the current value.
    effect(() => {
      const packaging = this.messagePage.packagingChoices()?.[this.recordId()] ?? 'root';
      this.form.patchValue({ packaging });
    });
    // Disable individual packaging controls while selection is active or the
    // message is already archived.
    effect(() =>
      this.packagingEnabled()
        ? this.form.get('packaging')?.enable()
        : this.form.get('packaging')?.disable(),
    );
  }

  /** Updates the form when `record` changes. */
  private registerRecord(): void {
    effect(() => {
      const record = this.record();
      const appraisalRecomm = record?.archiveMetadata?.appraisalRecommCode;
      let confidentiality: string | undefined;
      if (record?.generalMetadata?.confidentialityLevel) {
        confidentiality =
          confidentialityLevels[record.generalMetadata?.confidentialityLevel].shortDesc;
      }
      let medium: string | undefined;
      if (record?.generalMetadata?.medium) {
        medium = media[record.generalMetadata?.medium].shortDesc;
      }
      this.form.patchValue({
        recordPlanId: record?.generalMetadata?.filePlan?.filePlanNumber,
        recordPlanSubject: record?.generalMetadata?.filePlan?.subject,
        fileId: record?.generalMetadata?.recordNumber,
        subject: record?.generalMetadata?.subject,
        leadership: record?.generalMetadata?.leadership,
        fileManager: record?.generalMetadata?.fileManager,
        fileType: record?.type,
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
        if (value !== this.appraisal()?.note && this.appraisalComplete() === false) {
          this.setAppraisalNote(value);
        }
      });
  }

  /**
   * Uses packaging stats to enrich the option labels with the additional
   * information.
   */
  private getEnrichPackagingChoices() {
    const choices = packagingChoices.map((option) => ({ ...option }));
    const stats = this.packagingStats.value();
    if (stats) {
      // Packaging stats for all packaging choices available. Enrich all choices.
      for (const choice of choices) {
        choice.disabled = !stats[choice.value].deepestLevelHasItems;
        choice.label = choice.label + ` (${printPackagingStats(stats[choice.value])})`;
      }
    } else {
      // The request for all packaging choices is still in flight. Use the
      // available information the enrich only the currently selected choice.
      const packaging = this.messagePage.packagingChoices()?.[this.recordId()] ?? 'root';
      const packagingStats = this.messagePage.packagingStats()?.[this.recordId()];
      if (packagingStats) {
        const choice = choices.find((choice) => choice.value === packaging)!;
        choice.label = choice.label + ` (${printPackagingStats(packagingStats)})`;
      }
    }
    return choices;
  }

  setAppraisal(decision: AppraisalCode): void {
    this.messagePage.setAppraisalDecision(this.record()!.recordId, decision);
  }

  setAppraisalNote(note: string): void {
    this.messagePage.setAppraisalInternalNote(this.record()!.recordId, note);
  }

  setPackaging(value: PackagingChoice): void {
    this.messagePage.setPackaging([this.record()!.recordId], value);
  }
}
