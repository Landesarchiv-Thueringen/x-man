import { CommonModule } from '@angular/common';
import { Component, Inject, TemplateRef, ViewChild } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatExpansionModule, MatExpansionPanel } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { Observable, firstValueFrom, map, startWith, switchMap, take } from 'rxjs';
import { AgenciesService, Agency } from '../../../services/agencies.service';
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
    transferDir: new FormControl(this.agency.transferDir, {
      nonNullable: true,
      validators: Validators.required,
    }),
    filesystemTransferDir: new FormGroup({
      path: new FormControl('', { nonNullable: true }),
    }),
    webDAVTransferDir: new FormGroup({
      url: new FormControl('', { nonNullable: true }),
      user: new FormControl(),
      password: new FormControl(),
    }),
    collectionId: new FormControl(this.agency.collectionId, {
      nonNullable: true,
    }),
    userIds: new FormControl(this.agency.users?.map((user) => user.id) ?? [], { nonNullable: true }),
  });
  archivistsFilterControl = new FormControl('');
  filteredArchivists: Observable<User[]>;
  assignedArchivists: Observable<User[]>;
  users = this.usersService.getUsers();
  collections = this.collectionsService.getCollections();
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
    private dialog: MatDialog,
    private usersService: UsersService,
    private collectionsService: CollectionsService,
    private agenciesService: AgenciesService,
    private transferDirectoryService: TransferDirService,
  ) {
    // Reset 'testResult' when the value of 'transferDir' changes
    this.form.get('transferDir')?.valueChanges.subscribe(() => (this.testResult = 'not-tested'));
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
    this.transferDirPanel.open();
    const value = this.form.getRawValue().transferDir;
    if (this.form.valid && !this.loadingTestResult) {
      this.loadingTestResult = true;
      const observable = this.transferDirectoryService.testTransferDir(value);
      try {
        const testResult = await firstValueFrom(observable);
        this.testResult = testResult.result;
      } catch {
        this.testResult = 'failed';
      } finally {
        this.loadingTestResult = false;
      }
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
        const { userIds, filesystemTransferDir, webDAVTransferDir, ...agency } = this.form.getRawValue();
        const updateAgency: Omit<Agency, 'id'> = {
          ...agency,
          users: userIds.map((userId) => ({ id: userId }) as User),
        };
        console.log(updateAgency);
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
}
