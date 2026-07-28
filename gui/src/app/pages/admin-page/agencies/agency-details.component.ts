import { CommonModule } from '@angular/common';
import { Component, ElementRef, TemplateRef, inject, signal, viewChild } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators, AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatChipsModule } from '@angular/material/chips';
import {
  MAT_DIALOG_DATA,
  MatDialog,
  MatDialogContent,
  MatDialogModule,
  MatDialogRef,
} from '@angular/material/dialog';
import { MatExpansionModule, MatExpansionPanel } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { MatTooltipModule } from '@angular/material/tooltip';
import { Observable, firstValueFrom, map, startWith, switchMap, take } from 'rxjs';
import { AgenciesService, Agency, TransferDir, TransferProtocol } from '../../../services/agencies.service';
import { ConfigService } from '../../../services/config.service';
import { User, UsersService } from '../../../services/users.service';
import { CollectionsService } from '../collections/collections.service';
import { TestResult, TransferDirService } from './transfer-dir.service';

/**
 * Agency metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-agency-details',
  imports: [
    CommonModule,
    MatAutocompleteModule,
    MatButtonModule,
    MatCheckboxModule,
    MatChipsModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    MatTooltipModule,
    ReactiveFormsModule,
  ],
  templateUrl: './agency-details.component.html',
  styleUrl: './agency-details.component.scss',
})
export class AgencyDetailsComponent {
  private dialogRef = inject<MatDialogRef<AgencyDetailsComponent>>(MatDialogRef);
  private agency = inject<Agency>(MAT_DIALOG_DATA);
  private agenciesService = inject(AgenciesService);
  private collectionsService = inject(CollectionsService);
  private configService = inject(ConfigService);
  private dialog = inject(MatDialog);
  private transferDirectoryService = inject(TransferDirService);
  private usersService = inject(UsersService);

  readonly deleteDialogTemplate = viewChild.required<TemplateRef<unknown>>('deleteDialog');
  readonly transferDirPanel = viewChild.required<MatExpansionPanel>('transferDirPanel');
  readonly dialogContent = viewChild.required(MatDialogContent, { read: ElementRef });

  readonly oldName = this.agency.name;
  form = new FormGroup({
    name: new FormControl(this.agency.name, { nonNullable: true, validators: Validators.required }),
    abbreviation: new FormControl(this.agency.abbreviation, {
      nonNullable: true,
      validators: Validators.required,
    }),
    prefix: new FormControl(this.agency.prefix, { nonNullable: true }),
    code: new FormControl(this.agency.code, { nonNullable: true }),
    contactEmail: new FormControl(this.agency.contactEmail, { nonNullable: true }),
    transferDir: new FormGroup({
      protocol: new FormControl<TransferProtocol>('file', {
        nonNullable: true,
        validators: Validators.required,
      }),
      host: new FormControl<string>('', { nonNullable: true, validators: [Validators.required] }),
      path: new FormControl<string>('', { nonNullable: true, validators: [Validators.required] }),
      user: new FormControl<string>('', { nonNullable: true }),
      password: new FormControl<string>('', { nonNullable: true }),
      allowInsecureTLS: new FormControl<boolean>(false, { nonNullable: true }),
      path0502: new FormControl<string>('', { nonNullable: true, validators: [this.testResultValidator('path0502Exists')] }),
      path0504: new FormControl<string>('', { nonNullable: true, validators: [this.testResultValidator('path0504Exists')] }),
      path0506: new FormControl<string>('', { nonNullable: true, validators: [this.testResultValidator('path0506Exists')] }),
      path0507: new FormControl<string>('', { nonNullable: true, validators: [this.testResultValidator('path0507Exists')] }),
    }),
    collectionId: new FormControl(this.agency.collectionId, {
      nonNullable: true,
    }),
    userIds: new FormControl(this.agency.users ?? [], { nonNullable: true }),
  });
  archivistsFilterControl = new FormControl('');
  filteredArchivists: Observable<User[]>;
  assignedArchivists: Observable<User[]>;
  users = this.usersService.getUsers();
  collections = this.collectionsService.getCollections();
  config = this.configService.config;

  showPassword = signal(false);
  togglePasswordVisibility(event: MouseEvent) {
    this.showPassword.set(!this.showPassword());
    event.stopPropagation();
  }

  /**
   * The result of testing the configuration of the transfer-directory.
   *
   * - 'success' / 'failed': the test has run successfully / unsuccessfully with
   *   the current configuration as reflected by the form group
   * - 'unchanged': the configuration has not been modified since opening the
   *   dialog
   * - 'not-tested': the configuration has changed since opening the dialog and
   *   the test was not yet run
   */
  testState: 'success' | 'failed' | 'not-tested' | 'unchanged' = 'unchanged';
  testResult?: TestResult;
  loadingTestResult = false;
  isNew = this.agency.id == null;

  testResultValidator(resultKey: keyof TestResult): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      if (!control.value || this.testState === 'not-tested' || !this.testResult) {
        return null;
      }
      return this.testResult[resultKey] != undefined && this.testResult[resultKey]
        ? null
        : { pathNotFound: true };
    };
  }

  constructor() {
    // Reset 'testResult' when the value of 'transferDir' changes
    this.form.get('transferDir')?.valueChanges.subscribe(() => {
      this.testState = 'not-tested'
    });
    ['path0502', 'path0504', 'path0506', 'path0507'].forEach(config => {
      this.form.get('transferDir')?.get(config)?.valueChanges.subscribe(() => {
        this.testState = 'not-tested'
        this.form.get('transferDir')?.get(config)?.updateValueAndValidity({
          onlySelf: true,
          emitEvent: false
        })
      });
    })
    this.initTransferDirGroup();
    // Disable close on backdrop click as soon as the user modifies any value
    this.form.valueChanges.pipe(take(1)).subscribe(() => (this.dialogRef.disableClose = true));
    // Bind autocomplete results for archivists
    this.filteredArchivists = this.archivistsFilterControl.valueChanges.pipe(
      startWith(null),
      switchMap((filterString: string | null) => this.filterArchivists(filterString)),
    );
    // Resolve userIds to archivist objects
    this.assignedArchivists = this.form.get('userIds')!.valueChanges.pipe(
      startWith(this.form.getRawValue().userIds),
      switchMap((userIds) => this.usersService.getUsersByIds(userIds)),
    );
  }

  /**
   * Tests whether the transfer-directory configuration currently reflected by
   * `form` is reachable and allows read/write access.
   *
   * Saves the result to `testResult`.
   *
   * Sets `loadingTestResult` to true while running.
   */
  async testTransferDirectory() {
    this.trimTransferDirInputs();
    this.transferDirPanel().open();
    if (this.isTransferDirConfigTestable() && !this.loadingTestResult) {
      this.loadingTestResult = true;
      const observable = this.transferDirectoryService.testTransferDir(this.form.getRawValue().transferDir);
      try {
        this.testResult = await firstValueFrom(observable);
        this.testState = this.testResult?.success ? 'success' : 'failed';
        ['path0502', 'path0504', 'path0506', 'path0507'].forEach(config => {
          this.form.get('transferDir')?.get(config)?.updateValueAndValidity({
            onlySelf: true,
            emitEvent: false
          })
        });
      } catch {
        this.testState = 'failed';
      } finally {
        this.loadingTestResult = false;
      }
      this.scrollToBottom();
    }
  }

  isTransferDirConfigTestable(): boolean {
    const transferDirFormGroup = this.form.get('transferDir') as FormGroup
    return !Object.values(transferDirFormGroup.controls)
      .some(control => control.hasError('required'));
  }

  /**
   * Assigns the given archivist as responsible for this agency.
   */
  addArchivist(archivistId: string) {
    const currentIds = this.form.getRawValue().userIds;
    if (!currentIds.includes(archivistId)) {
      this.form.patchValue({ userIds: [...currentIds, archivistId] });
    }
    this.archivistsFilterControl.setValue('');
  }

  /**
   * Removes the given archivist's assignment to this agency.
   */
  removeArchivist(archivist: User) {
    const currentIds = this.form.getRawValue().userIds;
    this.form.patchValue({ userIds: currentIds.filter((id) => id !== archivist.id) });
  }

  /**
   * Saves the dialog data and closes the dialog.
   *
   * If the transfer directory has not yet been tested with the current
   * configuration, runs the test before saving.
   *
   * If the transfer directory could not be tested successfully (either by this
   * function or before), aborts.
   */
  async save() {
    if (this.form.valid) {
      if (this.testState === 'not-tested') {
        await this.testTransferDirectory();
      }
      if (this.testState !== 'failed') {
        const { userIds, transferDir, ...agency } = this.form.getRawValue();
        const updateAgency: Omit<Agency, 'id'> = {
          ...agency,
          users: userIds,
          transferDir: transferDir,
        };
        this.dialogRef.close(updateAgency);
      }
    }
  }

  /**
   * Resolves to the list of archivists to show in the autocomplete panel.
   */
  private filterArchivists(filterString: string | null): Observable<User[]> {
    const filterStringLower = filterString?.toLowerCase() ?? null;
    return this.users.pipe(
      map((archivists) =>
        archivists.filter(
          (a) =>
            (filterStringLower == null ||
              a.displayName.toLocaleLowerCase().includes(filterStringLower)) &&
            // Filter archivists that are already assigned
            !this.form.getRawValue().userIds.includes(a.id),
        ),
      ),
    );
  }

  /**
   * Deletes this agency after getting user confirmation and closes the dialog.
   */
  deleteAgency() {
    const dialogRef = this.dialog.open(this.deleteDialogTemplate());
    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.agenciesService.deleteAgency(this.agency);
        this.dialogRef.close();
      }
    });
  }

  /** Trims and removes superfluous characters likely to be inserted by users. */
  trimTransferDirInputs(): void {
    const transferDir = this.form.get('transferDir');
    let host = transferDir?.getRawValue().host;
    host = host?.trim();
    transferDir?.get('host')?.setValue(host);
    transferDir?.get('path')?.setValue(this.trimPath(transferDir?.getRawValue().path));
    transferDir?.get('path0502')?.setValue(this.trimPath(transferDir?.getRawValue().path0502));
    transferDir?.get('path0504')?.setValue(this.trimPath(transferDir?.getRawValue().path0504));
    transferDir?.get('path0506')?.setValue(this.trimPath(transferDir?.getRawValue().path0506));
    transferDir?.get('path0507')?.setValue(this.trimPath(transferDir?.getRawValue().path0507));
  }

  /** Trims leading and trailing slashes and whitespaces from a path. */
  trimPath(path: string | null) {
    return path ? path.trim().replace(/^\/|\/$/g, '') : ''
  }

  /**
   * Initial setup for the transfer-dir form group.
   *
   * - Registers change listeners that update the form group based on the
   *   selected protocol.
   * - Initializes the form fields with values extracted from the transfer-dir
   *   URI saved in the database.
   */
  private initTransferDirGroup(): void {
    // Update fields based on selected protocol
    this.form
      .get('transferDir')
      ?.get('protocol')
      ?.valueChanges.subscribe((value) => {
        const path = this.form.get('transferDir')?.get('path');
        const host = this.form.get('transferDir')?.get('host');
        const user = this.form.get('transferDir')?.get('user');
        const password = this.form.get('transferDir')?.get('password');
        switch (value) {
          case 'file':
            path?.enable();
            path?.setValidators(Validators.required);
            host?.disable();
            host?.clearValidators();
            host?.setValue('');
            user?.disable();
            user?.setValue('');
            password?.disable();
            password?.setValue('');
            break;
          case 'dav':
          case 'davs':
            path?.enable();
            path?.clearValidators();
            host?.enable();
            host?.setValidators(Validators.required);
            user?.enable();
            password?.enable();
            // remove the prefilled default root dir when switching to webDAV
            if (path?.value === this.transferDirectoryService.getDefaultRootDir()) {
              path?.patchValue('')
            }
            break;
        }
        host?.updateValueAndValidity();
        this.form.get('transferDir')?.get('allowInsecureTLS')?.patchValue(false)
      });
    // Populate fields with initial values from the database
    const formGroup = this.form.get('transferDir')!;
    formGroup.get('protocol')?.setValue(this.agency.transferDir.protocol);
    formGroup.get('user')?.setValue(this.agency.transferDir.user);
    formGroup.get('password')?.setValue(this.agency.transferDir.password);
    formGroup.get('host')?.setValue(this.agency.transferDir.host);
    formGroup.get('path')?.setValue(this.trimPath(this.agency.transferDir.path));
    formGroup.get('allowInsecureTLS')?.setValue(this.agency.transferDir.allowInsecureTLS);
    formGroup.get('path0502')?.setValue(this.agency.transferDir.path0502);
    formGroup.get('path0504')?.setValue(this.agency.transferDir.path0504);
    formGroup.get('path0506')?.setValue(this.agency.transferDir.path0506);
    formGroup.get('path0507')?.setValue(this.agency.transferDir.path0507);
  }

  private scrollToBottom(): void {
    const scrollParent = this.dialogContent().nativeElement;
    function scroll() {
      scrollParent.scroll({ top: 1000000, behavior: 'smooth' });
    }
    window.requestAnimationFrame(scroll);
  }
}
