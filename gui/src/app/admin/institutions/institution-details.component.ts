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
import { CollectionsService } from '../collections/collections.service';
import { User, UsersService } from '../users/users.service';
import { Institution, InstitutionsService } from './institutions.service';
import { TransferDirectoryService } from './transfer-directory.service';

/**
 * Institution metadata and associations shown in a dialog.
 */
@Component({
  selector: 'app-institution-details',
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
  templateUrl: './institution-details.component.html',
  styleUrl: './institution-details.component.scss',
})
export class InstitutionDetailsComponent {
  @ViewChild('deleteDialog') deleteDialogTemplate!: TemplateRef<unknown>;
  @ViewChild('transferDirectoryPanel') transferDirectoryPanel!: MatExpansionPanel;

  readonly oldName = this.institution.name;
  form = new FormGroup({
    name: new FormControl(this.institution.name, { nonNullable: true, validators: Validators.required }),
    abbreviation: new FormControl(this.institution.abbreviation, {
      nonNullable: true,
      validators: Validators.required,
    }),
    transferDirectory: new FormGroup({
      uri: new FormControl(this.institution.transferDirectory.uri, {
        nonNullable: true,
        validators: Validators.required,
      }),
      username: new FormControl(this.institution.transferDirectory.username, {
        nonNullable: true,
        validators: Validators.required,
      }),
      password: new FormControl(this.institution.transferDirectory.password, {
        nonNullable: true,
        validators: Validators.required,
      }),
    }),
    collectionId: new FormControl(this.institution.collectionId, {
      nonNullable: true,
    }),
    userIds: new FormControl(this.institution.userIds ?? [], { nonNullable: true }),
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
  isNew = this.institution.id == null;

  constructor(
    private dialogRef: MatDialogRef<InstitutionDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) private institution: Institution,
    private dialog: MatDialog,
    private usersService: UsersService,
    private collectionsService: CollectionsService,
    private institutionsService: InstitutionsService,
    private transferDirectoryService: TransferDirectoryService,
  ) {
    // Reset 'testResult' when any value of 'transferDirectory' changes
    this.form.get('transferDirectory')?.valueChanges.subscribe(() => (this.testResult = 'not-tested'));
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
    this.transferDirectoryPanel.open();
    const value = this.form.getRawValue().transferDirectory;
    if (this.form.valid && !this.loadingTestResult) {
      this.loadingTestResult = true;
      const observable = this.transferDirectoryService.testTransferDirectory(value);
      try {
        const testResult = await firstValueFrom(observable);
        this.testResult = testResult;
      } catch {
        this.testResult = 'failed';
      } finally {
        this.loadingTestResult = false;
      }
    }
  }

  /**
   * Assigns the given archivist as responsible for this institution.
   */
  addArchivist(archivistId: string) {
    const currentIds = this.form.getRawValue().userIds;
    if (!currentIds.includes(archivistId)) {
      this.form.patchValue({ userIds: [...currentIds, archivistId] });
    }
    this.archivistsFilterControl.setValue('');
  }

  /**
   * Removes the given archivist's assignment to this institution.
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
        this.dialogRef.close(this.form.value);
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
   * Deletes this institution after getting user confirmation and closes the dialog.
   */
  deleteInstitution() {
    const dialogRef = this.dialog.open(this.deleteDialogTemplate);
    dialogRef.afterClosed().subscribe((confirmed) => {
      if (confirmed) {
        this.institutionsService.deleteInstitution(this.institution);
        this.dialogRef.close();
      }
    });
  }
}
