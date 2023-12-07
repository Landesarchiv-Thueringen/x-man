// angular
import { NgModule, LOCALE_ID } from '@angular/core';
import { DatePipe, registerLocaleData } from '@angular/common';
import { BrowserModule } from '@angular/platform-browser';
import { ClipboardModule } from '@angular/cdk/clipboard'; 
import { ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import localeDe from '@angular/common/locales/de';
registerLocaleData(localeDe);

// material
import { MatButtonModule } from '@angular/material/button';
import { MatCheckboxModule } from '@angular/material/checkbox'; 
import { MatDialogModule } from '@angular/material/dialog'; 
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field'; 
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu'; 
import { MatPaginatorModule, MatPaginatorIntl} from '@angular/material/paginator';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner'; 
import { MatSelectModule } from '@angular/material/select';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatSortModule} from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table'; 
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTreeModule } from '@angular/material/tree';

// project
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { DocumentMetadataComponent } from './metadata/document-metadata/document-metadata.component';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { FileFeaturePipe } from './utility/localization/file-attribut-de.pipe';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { FeatureBreakPipe } from './utility/formatting/feature-break.pipe';
import { InstitutMetadataComponent } from './metadata/institution-metadata/institution-metadata.component';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { MessageTreeComponent } from './message/message-tree/message-tree.component';
import { PaginatorDeService } from './utility/localization/paginator-de.service';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';
import { RecordObjectAppraisalPipe } from './metadata/record-object-appraisal-pipe';
import { ProcessTableComponent } from './process/process-table/process-table.component';
import { DocumentVersionMetadataComponent } from './metadata/document-version-metadata/document-version-metadata.component';
import { ClearingTableComponent } from './clearing/clearing-table/clearing-table.component';
import { PrimaryDocumentsTableComponent } from './metadata/primary-documents-table/primary-documents-table.component';
import { FileOverviewComponent } from './metadata/primary-document/primary-document-metadata.component';

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
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    ClipboardModule,
    HttpClientModule,
    MatButtonModule,
    MatCheckboxModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatMenuModule,
    MatPaginatorModule,
    MatProgressSpinnerModule,
    MatSelectModule,
    MatSidenavModule,
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
    DatePipe,
    FeatureBreakPipe,
    FileFeaturePipe,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
