import { ClipboardModule } from '@angular/cdk/clipboard';
import { DatePipe, registerLocaleData } from '@angular/common';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import localeDe from '@angular/common/locales/de';
import { LOCALE_ID, NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS, MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatPaginatorIntl, MatPaginatorModule } from '@angular/material/paginator';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSelectModule } from '@angular/material/select';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatSortModule } from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTreeModule } from '@angular/material/tree';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { ClearingTableComponent } from './clearing/clearing-table/clearing-table.component';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { AppraisalFormComponent } from './message/appraisal-form/appraisal-form.component';
import { FinalizeAppraisalDialogComponent } from './message/finalize-appraisal-dialog/finalize-appraisal-dialog.component';
import { MessageTreeComponent } from './message/message-tree/message-tree.component';
import { StartArchivingDialogComponent } from './message/start-archiving-dialog/start-archiving-dialog.component';
import { DocumentMetadataComponent } from './metadata/document-metadata/document-metadata.component';
import { DocumentVersionMetadataComponent } from './metadata/document-version-metadata/document-version-metadata.component';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { InstitutMetadataComponent } from './metadata/institution-metadata/institution-metadata.component';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { FileOverviewComponent } from './metadata/primary-document/primary-document-metadata.component';
import { PrimaryDocumentsTableComponent } from './metadata/primary-documents-table/primary-documents-table.component';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';
import { RecordObjectAppraisalPipe } from './metadata/record-object-appraisal-pipe';
import { ProcessTableComponent } from './process/process-table/process-table.component';
import { AuthInterceptor } from './utility/authorization/auth-interceptor';
import { FeatureBreakPipe } from './utility/formatting/feature-break.pipe';
import { FileFeaturePipe } from './utility/localization/file-attribut-de.pipe';
import { PaginatorDeService } from './utility/localization/paginator-de.service';
registerLocaleData(localeDe);

@NgModule({
  declarations: [
    AppComponent,
    DocumentMetadataComponent,
    MainNavigationComponent,
    MessageTreeComponent,
    MessageMetadataComponent,
    FeatureBreakPipe,
    FileFeaturePipe,
    FileMetadataComponent,
    FileOverviewComponent,
    InstitutMetadataComponent,
    ProcessMetadataComponent,
    RecordObjectAppraisalPipe,
    ProcessTableComponent,
    DocumentVersionMetadataComponent,
    ClearingTableComponent,
    PrimaryDocumentsTableComponent,
    AppraisalFormComponent,
    FinalizeAppraisalDialogComponent,
    StartArchivingDialogComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    ClipboardModule,
    FormsModule,
    HttpClientModule,
    MatButtonModule,
    MatCheckboxModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatListModule,
    MatMenuModule,
    MatPaginatorModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    MatSidenavModule,
    MatSlideToggleModule,
    MatSnackBarModule,
    MatSortModule,
    MatTableModule,
    MatTreeModule,
    MatToolbarModule,
    ReactiveFormsModule,
  ],
  providers: [
    { provide: LOCALE_ID, useValue: 'de' },
    { provide: MatPaginatorIntl, useClass: PaginatorDeService },
    { provide: MAT_FORM_FIELD_DEFAULT_OPTIONS, useValue: { appearance: 'outline' } },
    { provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true },
    DatePipe,
    FeatureBreakPipe,
    FileFeaturePipe,
  ],
  bootstrap: [AppComponent],
})
export class AppModule {}
