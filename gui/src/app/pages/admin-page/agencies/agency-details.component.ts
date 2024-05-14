import { CommonModule } from '@angular/common';
import { Component, ElementRef, Inject, TemplateRef, ViewChild } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialog, MatDialogContent, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule, MatExpansionPanel } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { Observable, firstValueFrom, map, startWith, switchMap, take } from 'rxjs';
import { AgenciesService, Agency } from '../../../services/agencies.service';
import { ConfigService } from '../../../services/config.service';
import { User, UsersService } from '../../../services/users.service';
import { CollectionsService } from '../collections/collections.service';
import { TransferDirService } from './transfer-dir.service';

/**
 * Agency metadata and associations.
 *
 * Shown in a dialog.
 */
@Component({
  selector: 'app-agency-details',
  standalone: true,
  imports: [
    CommonModule,
    MatAutocompleteModule,
    MatButtonModule,
    MatChipsModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    ReactiveFormsModule,
  ],
  templateUrl: './agency-details.component.html',
  styleUrl: './agency-details.component.scss',
})
export class AgencyDetailsComponent {
  @ViewChild('deleteDialog') deleteDialogTemplate!: TemplateRef<unknown>;
  @ViewChild('transferDirPanel') transferDirPanel!: MatExpansionPanel;
  @ViewChild(MatDialogContent, { read: ElementRef }) dialogContent!: ElementRef<HTMLDivElement>;

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
      protocol: new FormControl('', {
        nonNullable: true,
        validators: Validators.required,
      }),
      host: new FormControl('', { validators: [Validators.required] }),
      path: new FormControl('', { validators: [Validators.required] }),
      username: new FormControl(''),
      password: new FormControl(''),
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
  config = toSignal(this.configService.config);
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
  testResult: 'success' | 'failed' | 'not-tested' | 'unchanged' = 'unchanged';
  loadingTestResult = false;
  isNew = this.agency.id == null;

  constructor(
    private dialogRef: MatDialogRef<AgencyDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) private agency: Agency,
    private agenciesService: AgenciesService,
    private collectionsService: CollectionsService,
    private configService: ConfigService,
    private dialog: MatDialog,
    private transferDirectoryService: TransferDirService,
    private usersService: UsersService,
  ) {
    // Reset 'testResult' when the value of 'transferDir' changes
    this.form.get('transferDir')?.valueChanges.subscribe(() => (this.testResult = 'not-tested'));
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
    this.fixupTransferDirInputs();
    this.transferDirPanel.open();
    if (this.form.get('transferDir')?.valid && !this.loadingTestResult) {
      this.loadingTestResult = true;
      const observable = this.transferDirectoryService.testTransferDir(this.getTransferDirURI());
      try {
        const testResult = await firstValueFrom(observable);
        this.testResult = testResult.result;
      } catch {
        this.testResult = 'failed';
      } finally {
        this.loadingTestResult = false;
      }
      this.scrollToBottom();
    }
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
      if (this.testResult === 'not-tested') {
        await this.testTransferDirectory();
      }
      if (this.testResult !== 'failed') {
        const { userIds, transferDir, ...agency } = this.form.getRawValue();
        const updateAgency: Omit<Agency, 'id'> = {
          ...agency,
          users: userIds,
          transferDirURL: this.getTransferDirURI(),
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
            (filterStringLower == null || a.displayName.toLocaleLowerCase().includes(filterStringLower)) &&
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
    const dialogRef = this.dialog.open(this.deleteDialogTemplate);
    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.agenciesService.deleteAgency(this.agency);
        this.dialogRef.close();
      }
    });
  }

  /** Trims and removes superfluous characters likely to be inserted by users. */
  fixupTransferDirInputs(): void {
    const transferDir = this.form.get('transferDir');
    let host = transferDir?.getRawValue().host;
    host = host?.trim();
    transferDir?.get('host')?.setValue(host);
    let path = transferDir?.getRawValue().path;
    path = path?.trim();
    path = path?.replace(/^\/|\/$/g, ''); // trim leading and trailing slashes
    transferDir?.get('path')?.setValue(path);
  }

  /** Combines information from the transfer-dir form group to a URI string. */
  private getTransferDirURI(): string {
    const transferDir = this.form.get('transferDir')!.value;
    // Create the URL as 'http' instead of 'dav' since URL will not behave correctly with 'dav'.
    const dummyProtocol = transferDir.protocol?.startsWith('dav') ? 'http' : transferDir.protocol;
    const transferDirURL = new URL(dummyProtocol + '://' + (transferDir.host ?? '') + '/' + transferDir.path);
    transferDirURL.username = transferDir.username ?? '';
    transferDirURL.password = transferDir.password ?? '';
    return transferDirURL.href.replace(/^http/, transferDir.protocol!);
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
        const username = this.form.get('transferDir')?.get('username');
        const password = this.form.get('transferDir')?.get('password');
        switch (value) {
          case 'file':
            path?.enable();
            path?.setValidators(Validators.required);
            host?.disable();
            host?.clearValidators();
            host?.setValue(null);
            username?.disable();
            username?.setValue(null);
            password?.disable();
            password?.setValue(null);
            break;
          case 'dav':
          case 'davs':
            path?.enable();
            path?.clearValidators();
            host?.enable();
            host?.setValidators(Validators.required);
            username?.enable();
            password?.enable();
            break;
        }
        host?.updateValueAndValidity();
      });
    // Populate fields with initial values from the database
    if (this.agency.transferDirURL) {
      const [protocol, rest] = this.agency.transferDirURL.split('://');
      try {
        // Create the URL as 'http' instead of 'dav' since URL will not behave correctly with 'dav'.
        const dummyProtocol = protocol?.startsWith('dav') ? 'http' : protocol;
        const url = new URL(dummyProtocol + '://' + rest);
        const formGroup = this.form.get('transferDir')!;
        formGroup.get('protocol')?.setValue(protocol);
        formGroup.get('username')?.setValue(url.username);
        formGroup.get('password')?.setValue(url.password);
        formGroup.get('host')?.setValue(url.host);
        formGroup.get('path')?.setValue(url.pathname.replace(/^\//, '')); // trim leading slash
      } catch (e) {
        console.warn('Failed to parse transfer-dir URI', this.agency.transferDirURL, e);
      }
    }
  }

  private scrollToBottom(): void {
    const scrollParent = this.dialogContent.nativeElement;
    function scroll() {
      scrollParent.scroll({ top: 1000000, behavior: 'smooth' });
    }
    window.requestAnimationFrame(scroll);
  }
}
