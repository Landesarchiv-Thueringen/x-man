import { CommonModule } from '@angular/common';
import { Component, Query, computed, effect, signal } from '@angular/core';
import { takeUntilDestroyed, toSignal } from '@angular/core/rxjs-interop';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { ActivatedRoute, Params } from '@angular/router';
import { combineLatest, firstValueFrom, switchMap } from 'rxjs';
import { debounceTime, shareReplay, skip } from 'rxjs/operators';
import {
  Appraisal,
  AppraisalCode,
  AppraisalService,
  appraisalDescriptions,
} from '../../../../services/appraisal.service';
import { MessageService } from '../../../../services/message.service';
import {
  PackagingOption,
  PackagingService,
  packagingOptions,
} from '../../../../services/packaging.service';
import { FileRecord } from '../../../../services/records.service';
import { MessagePageService } from '../../message-page.service';
import { MessageProcessorService, StructureNode } from '../../message-processor.service';
import { printPackagingStats } from '../../packaging-stats.pipe';
import { confidentialityLevels } from '../confidentiality-level.pipe';
import { media } from '../medium.pipe';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss'],
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
export class FileMetadataComponent {
  metadataQuery?: Query;
  /** The pages file record object. Might update on page changes. */
  record?: FileRecord;
  appraisal?: Appraisal | null;
  appraisalCodes = Object.entries(appraisalDescriptions).map(([code, d]) => ({ code, ...d }));
  appraisalComplete = signal(false);
  form: FormGroup;
  canBeAppraised = false;
  canChoosePackaging = false;
  selectionActive = this.messagePage.selectionActive;
  hasUnresolvedError = this.messagePage.hasUnresolvedError;
  message = this.messagePage.message;
  packagingOptions = [...packagingOptions.map((option) => ({ ...option }))];
  /**
   * When all packaging-option labels have been enriched with packaging stats,
   * this variable holds the record ID that the information belongs to.
   */
  private packagingOptionsPopulated = '';
  private params = toSignal(this.route.params, { initialValue: {} as Params });
  private recordId = computed<string>(() => this.params()['id']);
  packagingDecision = computed(() => this.messagePage.packagingDecisions()[this.recordId()] ?? '');

  constructor(
    private appraisalService: AppraisalService,
    private formBuilder: FormBuilder,
    private messagePage: MessagePageService,
    private messageService: MessageService,
    private messageProcessor: MessageProcessorService,
    private packagingService: PackagingService,
    private route: ActivatedRoute,
  ) {
    this.form = this.formBuilder.group({
      recordPlanId: new FormControl<string | null>(null),
      recordPlanSubject: new FormControl<string | null>(null),
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
      packaging: new FormControl<PackagingOption | null>(null),
    });
    const record = this.route.params.pipe(
      switchMap((params: Params) => this.messagePage.getFileRecord(params['id'])),
      shareReplay(1),
    );
    const structureNode = record.pipe(
      switchMap((record) => this.messageProcessor.getNodeWhenReady(record.recordId)),
    );
    const appraisal = record.pipe(
      switchMap((record) => this.messagePage.observeAppraisal(record.recordId)),
    );
    // Update the form and local properties on changes.
    combineLatest([record, structureNode, appraisal, this.messagePage.observeAppraisalComplete()])
      .pipe(takeUntilDestroyed())
      .subscribe(([recordObject, structureNode, appraisal, appraisalComplete]) =>
        this.setMetadata(recordObject, structureNode, appraisal, appraisalComplete),
      );
    this.registerAppraisalNoteChanges();
    // Disable individual appraisal controls while selection is active.
    effect(() => {
      if (
        !this.appraisalComplete() && // If the appraisal is complete, appraisal fields are readonly anyway.
        (this.messagePage.selectionActive() || this.messagePage.hasUnresolvedError())
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
      const packaging = this.messagePage.packagingOptions()?.[this.recordId()] ?? 'root';
      this.form.patchValue({ packaging });
    });
    // Reset packaging options when the record changes.
    effect(() => {
      this.recordId();
      this.packagingOptionsPopulated = '';
      this.packagingOptions = [...packagingOptions.map((option) => ({ ...option }))];
      this.enrichPackagingOptions();
    });
    // Enrich the currently selected packaging options with known stats while
    // the request for the remaining stats is in flight.
    effect(() => {
      if (this.packagingOptionsPopulated == this.recordId()) {
        return;
      }
      const packaging = this.messagePage.packagingOptions()?.[this.recordId()] ?? 'root';
      const packagingStats = this.messagePage.packagingStats()?.[this.recordId()];
      this.packagingOptions = [...packagingOptions.map((option) => ({ ...option }))];
      if (packagingStats) {
        const option = this.packagingOptions.find((option) => option.value === packaging)!;
        option.label = option.label + ` (${printPackagingStats(packagingStats)})`;
      }
    });
    // Disable individual packaging controls while selection is active or the
    // message is already archived.
    effect(() => {
      if (
        this.messagePage.selectionActive() ||
        this.messagePage.hasUnresolvedError() ||
        this.messagePage.process()?.processState.archiving.complete
      ) {
        this.form.get('packaging')?.disable();
      } else {
        this.form.get('packaging')?.enable();
      }
    });
  }

  /**
   * Fetches packaging stats for all packaging options from the backend and
   * enriches the option labels with the additional information.
   */
  private async enrichPackagingOptions(): Promise<void> {
    if (this.packagingOptionsPopulated || !this.messagePage.process()) {
      return;
    }
    this.packagingOptionsPopulated = this.recordId();
    const statsMap = await firstValueFrom(
      this.packagingService.getPackagingStats(this.messagePage.process()!.processId, [
        this.recordId(),
      ]),
    );
    for (const option of this.packagingOptions) {
      option.disabled = !statsMap[option.value].deepestLevelHasItems;
      if (option.label.includes('(')) {
        continue; // already includes stats
      }
      option.label = option.label + ` (${printPackagingStats(statsMap[option.value])})`;
    }
  }

  registerAppraisalNoteChanges(): void {
    this.form.controls['appraisalNote'].valueChanges
      .pipe(skip(1), debounceTime(400))
      .subscribe((value) => {
        if (value !== this.appraisal?.note && this.appraisalComplete() === false) {
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
    this.canChoosePackaging =
      this.messagePage.message()?.messageType === '0503' && structureNode.canChoosePackaging;
    this.appraisal = appraisal;
    this.appraisalComplete.set(appraisalComplete);
    const appraisalRecomm = this.record.archiveMetadata?.appraisalRecommCode;
    let confidentiality: string | undefined;
    if (record.generalMetadata?.confidentialityLevel) {
      confidentiality =
        confidentialityLevels[record.generalMetadata?.confidentialityLevel].shortDesc;
    }
    let medium: string | undefined;
    if (record.generalMetadata?.medium) {
      medium = media[record.generalMetadata?.medium].shortDesc;
    }
    this.form.patchValue({
      recordPlanId: this.record.generalMetadata?.filePlan?.filePlanNumber,
      recordPlanSubject: this.record.generalMetadata?.filePlan?.subject,
      fileId: this.record.generalMetadata?.recordNumber,
      subject: this.record.generalMetadata?.subject,
      fileType: this.record.type,
      lifeStart: this.messageService.getDateText(this.record.lifetime?.start),
      lifeEnd: this.messageService.getDateText(this.record.lifetime?.end),
      appraisal: this.appraisalComplete()
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

  setPackaging(value: PackagingOption): void {
    this.messagePage.setPackaging([this.record!.recordId], value);
  }
}
